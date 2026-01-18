package model

import "time"

// Account: ผังบัญชี (เช่น 101-เงินสด, 401-รายได้)
type Account struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"unique" json:"code"` // รหัสบัญชี
	Name      string    `json:"name"`               // ชื่อบัญชี
	Type      string    `json:"type"`               // Asset, Liability, Equity, Revenue, Expense
	CreatedAt time.Time `json:"created_at"`
}

type AccountSummary struct {
	Code        string  `json:"code"`
	AccountName string  `json:"account_name"`
	TotalDebit  float64 `json:"total_debit"`
	TotalCredit float64 `json:"total_credit"`
	Balance     float64 `json:"balance"` // Debit - Credit
}
