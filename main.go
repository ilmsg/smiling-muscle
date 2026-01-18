package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ilmsg/smiling-muscle/database"
	"github.com/ilmsg/smiling-muscle/handler"
	"github.com/ilmsg/smiling-muscle/model"
)

func main() {
	db := database.NewDatabaseWithSqlite("smiling-muscle.db")
	db.AutoMigrate(model.Account{}, model.Transaction{}, model.TransactionItem{})

	hAccount := handler.NewAccountHandler(db)

	r := gin.Default()
	r.POST("/accounts", hAccount.CreateAccount)              // สร้างผังบัญชี
	r.POST("/transactions", hAccount.CreateTransaction)      // บันทึกรายการค้า
	r.GET("/report/trial-balance", hAccount.GetTrialBalance) // ดูงบ

	// --- Routes ใหม่สำหรับงบการเงิน ---
	r.GET("/report/income-statement", hAccount.GetIncomeStatement) // งบกำไรขาดทุน
	r.GET("/report/balance-sheet", hAccount.GetBalanceSheet)       // งบดุล

	r.Run(":7000")
}
