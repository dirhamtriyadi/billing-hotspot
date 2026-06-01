package dto

// CheckoutRequest is the public self-service purchase request.
type CheckoutRequest struct {
	PackageID     uint   `json:"package_id" binding:"required"`
	CustomerName  string `json:"customer_name" binding:"required,max=120"`
	CustomerPhone string `json:"customer_phone" binding:"required,max=30"`
	CustomerEmail string `json:"customer_email" binding:"omitempty,email,max=160"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=cash xendit midtrans tripay"`
	// Channel is optional and provider-specific (e.g. Tripay "QRIS","BRIVA").
	Channel string `json:"channel" binding:"omitempty,max=20"`
}

// OrderListQuery filters the admin order list.
type OrderListQuery struct {
	PageQuery
	Status        string `form:"status"`
	PaymentMethod string `form:"payment_method"`
}
