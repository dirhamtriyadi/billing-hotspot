package dto

// CreateBatchRequest generates a batch of vouchers for a package. When Name is
// empty an automatic name is assigned. Used by both the "generate vouchers"
// quick action and named batch creation.
type CreateBatchRequest struct {
	Name       string `json:"name" binding:"omitempty,max=120"`
	PackageID  uint   `json:"package_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1,max=2000"`
	Prefix     string `json:"prefix" binding:"omitempty,max=12,alphanum"`
	CodeLength int    `json:"code_length" binding:"omitempty,min=4,max=20"`
}

// UpdateVoucherStatusRequest enables or disables a single voucher. Disabling
// revokes the credential in FreeRADIUS; enabling re-provisions it.
type UpdateVoucherStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active disabled"`
}

// VoucherListQuery filters the voucher list.
type VoucherListQuery struct {
	PageQuery
	Status    string `form:"status"`
	PackageID uint   `form:"package_id"`
	BatchID   uint   `form:"batch_id"`
}
