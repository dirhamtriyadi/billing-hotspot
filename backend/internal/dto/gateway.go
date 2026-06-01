package dto

// GatewaySettings is the masked, read-side view of every payment gateway's
// configuration. Secret credentials are never returned in full — only a short
// masked preview and a boolean indicating whether a value is stored.
type GatewaySettings struct {
	DefaultProvider string          `json:"default_provider"`
	EnableCash      bool            `json:"enable_cash"`
	Midtrans        GatewayMidtrans `json:"midtrans"`
	Xendit          GatewayXendit   `json:"xendit"`
	Tripay          GatewayTripay   `json:"tripay"`
}

// GatewayMidtrans is the Midtrans Snap configuration view.
type GatewayMidtrans struct {
	Enabled    bool   `json:"enabled"`    // offered to customers
	Configured bool   `json:"configured"` // has the credentials it needs
	Production bool   `json:"production"` // live vs sandbox
	ServerKey  string `json:"server_key"` // masked preview
	ClientKey  string `json:"client_key"` // public key, shown in full
}

// GatewayXendit is the Xendit configuration view.
type GatewayXendit struct {
	Enabled       bool   `json:"enabled"`
	Configured    bool   `json:"configured"`
	SecretKey     string `json:"secret_key"`     // masked preview
	CallbackToken string `json:"callback_token"` // masked preview
}

// GatewayTripay is the Tripay configuration view.
type GatewayTripay struct {
	Enabled      bool   `json:"enabled"`
	Configured   bool   `json:"configured"`
	Production   bool   `json:"production"`
	APIKey       string `json:"api_key"`       // masked preview
	PrivateKey   string `json:"private_key"`   // masked preview
	MerchantCode string `json:"merchant_code"` // shown in full
}

// GatewayUpdate is the admin request to change gateway settings. All fields are
// optional (pointer): a nil field is left untouched. For secret credential
// fields an empty string is also treated as "leave unchanged" so the masked
// value the UI displays can be submitted back harmlessly.
type GatewayUpdate struct {
	DefaultProvider *string               `json:"default_provider"`
	EnableCash      *bool                 `json:"enable_cash"`
	Midtrans        *GatewayMidtransInput `json:"midtrans"`
	Xendit          *GatewayXenditInput   `json:"xendit"`
	Tripay          *GatewayTripayInput   `json:"tripay"`
}

// GatewayMidtransInput carries Midtrans credential updates.
type GatewayMidtransInput struct {
	Enabled    *bool   `json:"enabled"`
	Production *bool   `json:"production"`
	ServerKey  *string `json:"server_key"`
	ClientKey  *string `json:"client_key"`
}

// GatewayXenditInput carries Xendit credential updates.
type GatewayXenditInput struct {
	Enabled       *bool   `json:"enabled"`
	SecretKey     *string `json:"secret_key"`
	CallbackToken *string `json:"callback_token"`
}

// GatewayTripayInput carries Tripay credential updates.
type GatewayTripayInput struct {
	Enabled      *bool   `json:"enabled"`
	Production   *bool   `json:"production"`
	APIKey       *string `json:"api_key"`
	PrivateKey   *string `json:"private_key"`
	MerchantCode *string `json:"merchant_code"`
}
