package dto

// NASInput registers or updates a NAS / RADIUS client (a Mikrotik router). The
// pairing is keyed by NASName (the router's IP/identity as RADIUS sees it).
type NASInput struct {
	NASName     string `json:"nasname" binding:"required,max=128"`
	ShortName   string `json:"shortname" binding:"max=32"`
	Secret      string `json:"secret" binding:"required,max=60"`
	Type        string `json:"type" binding:"max=30"`
	Description string `json:"description" binding:"max=200"`
	Ports       *int   `json:"ports" binding:"omitempty,min=1,max=65535"`
}
