package model

import "time"

// Transaction: รายการค้า (Header)
type Transaction struct {
	ID          uint              `gorm:"primaryKey" json:"id"`
	Date        time.Time         `json:"date"`
	Description string            `json:"description"`
	Items       []TransactionItem `gorm:"foreignKey:TransactionID" json:"items"`
	CreatedAt   time.Time         `json:"created_at"`
}

// Input struct สำหรับรับข้อมูล JSON
type CreateTransactionInput struct {
	Description string `json:"description" binding:"required"`
	Items       []struct {
		AccountID uint    `json:"account_id" binding:"required"`
		Debit     float64 `json:"debit"`
		Credit    float64 `json:"credit"`
	} `json:"items" binding:"required,min=2"` // ต้องมีอย่างน้อย 2 ขา (เดบิต/เครดิต)
}

// โครงสร้างข้อมูลสำหรับส่งกลับ งบกำไรขาดทุน
type IncomeStatementResponse struct {
	Revenue   float64 `json:"total_revenue"`
	Expense   float64 `json:"total_expense"`
	NetIncome float64 `json:"net_income"` // กำไรสุทธิ
}

// โครงสร้างข้อมูลสำหรับส่งกลับ งบดุล
type BalanceSheetResponse struct {
	Assets           float64 `json:"total_assets"`
	Liabilities      float64 `json:"total_liabilities"`
	Equity           float64 `json:"total_equity"`      // ทุนเดิม
	RetainedEarnings float64 `json:"retained_earnings"` // กำไรสะสม (มาจาก Net Income)
	TotalLiabEquity  float64 `json:"total_liabilities_and_equity"`
}
