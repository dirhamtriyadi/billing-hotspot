package models

// PaymentLog is an audit trail of inbound gateway webhooks/callbacks, retained
// for debugging reconciliation issues.
type PaymentLog struct {
	Base
	OrderID   *uint  `gorm:"index" json:"order_id,omitempty"`
	Provider  string `gorm:"size:20;index" json:"provider"`
	Event     string `gorm:"size:60" json:"event"`
	Reference string `gorm:"size:120;index" json:"reference"`
	Status    string `gorm:"size:30" json:"status"`
	Signature string `gorm:"size:255" json:"signature"`
	Valid     bool   `json:"valid"` // signature verification result
	Payload   string `gorm:"type:text" json:"payload"`
}

// Setting is a key/value store for editable business settings (site name,
// contact, enabled providers, ...). Gateway secrets stay in environment vars.
type Setting struct {
	Key   string `gorm:"primaryKey;size:80" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}
