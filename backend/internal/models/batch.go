package models

// VoucherBatch is a group of vouchers generated together for a single package,
// typically printed and sold for cash.
type VoucherBatch struct {
	Base
	Name       string  `gorm:"size:120;not null" json:"name"`
	PackageID  uint    `gorm:"not null;index" json:"package_id"`
	Package    Package `gorm:"constraint:OnDelete:RESTRICT" json:"package,omitempty"`
	Prefix     string  `gorm:"size:12" json:"prefix"`
	Quantity   int     `gorm:"not null" json:"quantity"`
	CodeLength int     `gorm:"not null;default:8" json:"code_length"`
	CreatedBy  uint    `gorm:"index" json:"created_by"`

	// Voucher.BatchID is non-conventional (GORM would expect VoucherBatchID),
	// so the foreign key is declared explicitly.
	Vouchers []Voucher `gorm:"foreignKey:BatchID;references:ID;constraint:OnDelete:CASCADE" json:"vouchers,omitempty"`
}
