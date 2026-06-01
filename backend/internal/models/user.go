package models

// User is an administrator/operator of the billing panel. Hotspot customers are
// NOT users — they authenticate against FreeRADIUS via voucher codes.
type User struct {
	Base
	Name     string `gorm:"size:120;not null" json:"name"`
	Username string `gorm:"size:60;uniqueIndex;not null" json:"username"`
	Email    string `gorm:"size:160;uniqueIndex" json:"email"`
	Password string `gorm:"size:255;not null" json:"-"` // bcrypt hash, never serialized
	Role     string `gorm:"size:20;not null;default:operator" json:"role"`
	IsActive bool   `gorm:"not null;default:true" json:"is_active"`
}
