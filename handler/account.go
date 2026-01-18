package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ilmsg/smiling-muscle/model"
	"gorm.io/gorm"
)

type IAccountHandler interface {
	CreateAccount(c *gin.Context)
	CreateTransaction(c *gin.Context)
	GetTrialBalance(c *gin.Context)

	GetIncomeStatement(c *gin.Context)
	GetBalanceSheet(c *gin.Context)
}

type AccountHandler struct {
	db *gorm.DB
}

// 1. สร้างบัญชีใหม่
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var account model.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if result := h.db.Create(&account); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusCreated, account)
}

// 2. บันทึกบัญชี (Journal Entry) - *หัวใจสำคัญ*
func (h *AccountHandler) CreateTransaction(c *gin.Context) {
	var input model.CreateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation: ตรวจสอบหลักการบัญชีคู่ (Debit ต้องเท่ากับ Credit)
	var totalDebit, totalCredit float64
	for _, item := range input.Items {
		totalDebit += item.Debit
		totalCredit += item.Credit
	}

	// หมายเหตุ: การเปรียบเทียบ float อาจมีปัญหาทศนิยม ในงานจริงควรใช้ Library 'decimal'
	if totalDebit != totalCredit {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ยอด Debit และ Credit ไม่เท่ากัน (Unbalanced)"})
		return
	}

	// ใช้ Transaction ของ Database (Rollback ถ้าพังกลางทาง)
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// สร้าง Header
		newTx := model.Transaction{
			Date:        time.Now(),
			Description: input.Description,
		}
		if err := tx.Create(&newTx).Error; err != nil {
			return err
		}

		// สร้าง Items
		for _, item := range input.Items {
			newItem := model.TransactionItem{
				TransactionID: newTx.ID,
				AccountID:     item.AccountID,
				Debit:         item.Debit,
				Credit:        item.Credit,
			}
			if err := tx.Create(&newItem).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "บันทึกบัญชีสำเร็จ"})
}

// 3. ดูงบทดลอง (Trial Balance) - สรุปยอดเงินแต่ละบัญชี
func (h *AccountHandler) GetTrialBalance(c *gin.Context) {
	var results []model.AccountSummary

	// Query รวมยอด Debit/Credit ของแต่ละ AccountID
	err := h.db.Table("transaction_items").
		Select("accounts.code, accounts.name as account_name, SUM(transaction_items.debit) as total_debit, SUM(transaction_items.credit) as total_credit").
		Joins("left join accounts on transaction_items.account_id = accounts.id").
		Group("transaction_items.account_id").
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// คำนวณ Balance
	for i := range results {
		results[i].Balance = results[i].TotalDebit - results[i].TotalCredit
	}

	c.JSON(http.StatusOK, results)
}

func (h *AccountHandler) GetIncomeStatement(c *gin.Context) {
	// รับ parameter วันที่ (ถ้าไม่ส่งมา ให้ default เป็นทั้งหมด)
	startDateStr := c.Query("start_date") // format: YYYY-MM-DD
	endDateStr := c.Query("end_date")

	// Base Query
	query := h.db.Table("transaction_items").
		Joins("JOIN accounts ON transaction_items.account_id = accounts.id").
		Joins("JOIN transactions ON transaction_items.transaction_id = transactions.id")

	// Filter วันที่
	if startDateStr != "" {
		query = query.Where("transactions.date >= ?", startDateStr)
	}
	if endDateStr != "" {
		query = query.Where("transactions.date <= ?", endDateStr)
	}

	type Result struct {
		Type   string
		Debit  float64
		Credit float64
	}
	var results []Result

	// Group by Account Type เพื่อรวมยอดแยกตามหมวด
	query.Select("accounts.type, SUM(transaction_items.debit) as debit, SUM(transaction_items.credit) as credit").
		Where("accounts.type IN ?", []string{"Revenue", "Expense"}). // เอาแค่ รายได้ กับ ค่าใช้จ่าย
		Group("accounts.type").
		Scan(&results)

	var report model.IncomeStatementResponse

	for _, r := range results {
		if r.Type == "Revenue" {
			// รายได้: ธรรมชาติอยู่ฝั่ง Credit
			report.Revenue += (r.Credit - r.Debit)
		} else if r.Type == "Expense" {
			// ค่าใช้จ่าย: ธรรมชาติอยู่ฝั่ง Debit
			report.Expense += (r.Debit - r.Credit)
		}
	}

	report.NetIncome = report.Revenue - report.Expense
	c.JSON(http.StatusOK, report)
}

func (h *AccountHandler) GetBalanceSheet(c *gin.Context) {
	asOfDate := c.Query("date") // format: YYYY-MM-DD

	// 1. คำนวณยอดคงเหลือของแต่ละหมวด (Asset, Liability, Equity)
	query := h.db.Table("transaction_items").
		Joins("JOIN accounts ON transaction_items.account_id = accounts.id").
		Joins("JOIN transactions ON transaction_items.transaction_id = transactions.id")

	if asOfDate != "" {
		query = query.Where("transactions.date <= ?", asOfDate)
	}

	type Result struct {
		Type   string
		Debit  float64
		Credit float64
	}
	var results []Result

	// ดึงข้อมูลทุกหมวด
	query.Select("accounts.type, SUM(transaction_items.debit) as debit, SUM(transaction_items.credit) as credit").
		Group("accounts.type").
		Scan(&results)

	var report model.BalanceSheetResponse
	var totalRevenue, totalExpense float64

	for _, r := range results {
		switch r.Type {
		case "Asset":
			// สินทรัพย์: Debit - Credit
			report.Assets += (r.Debit - r.Credit)
		case "Liability":
			// หนี้สิน: Credit - Debit
			report.Liabilities += (r.Credit - r.Debit)
		case "Equity":
			// ทุน: Credit - Debit
			report.Equity += (r.Credit - r.Debit)
		case "Revenue":
			totalRevenue += (r.Credit - r.Debit)
		case "Expense":
			totalExpense += (r.Debit - r.Credit)
		}
	}

	// 2. คำนวณกำไรสะสม (Retained Earnings) เพื่อปิดเข้ากองทุน
	// เพราะในระบบจริง บัญชีรายได้/ค่าใช้จ่าย จะถูกปิดสิ้นปีเข้าสู่กำไรสะสม
	report.RetainedEarnings = totalRevenue - totalExpense

	// 3. รวมยอดฝั่งหนี้สินและทุน
	report.TotalLiabEquity = report.Liabilities + report.Equity + report.RetainedEarnings

	c.JSON(http.StatusOK, report)
}

func NewAccountHandler(db *gorm.DB) IAccountHandler {
	return &AccountHandler{db}
}
