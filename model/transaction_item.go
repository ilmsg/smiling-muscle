package model

// TransactionItem: รายละเอียด เดบิต/เครดิต
type TransactionItem struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	TransactionID uint    `json:"-"` // ผูกกับ Transaction
	AccountID     uint    `json:"account_id"`
	Account       Account `json:"-"`
	Debit         float64 `json:"debit"`
	Credit        float64 `json:"credit"`
}
