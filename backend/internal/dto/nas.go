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

	HotspotConfig NASHotspotConfigInput `json:"hotspot_config"`
}

type RadiusServerInput struct {
	Name        string `json:"name" binding:"required,max=120"`
	APIURL      string `json:"api_url" binding:"required,max=255"`
	APIKey      string `json:"api_key" binding:"max=255"`
	RadiusIP    string `json:"radius_ip" binding:"max=128"`
	CoAPort     string `json:"coa_port" binding:"max=10"`
	Timeout     string `json:"timeout" binding:"max=20"`
	Description string `json:"description" binding:"max=200"`
	IsDefault   bool   `json:"is_default"`
}

type RadiusServerOutput struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	APIURL      string `json:"api_url"`
	APIKey      string `json:"api_key"`
	RadiusIP    string `json:"radius_ip"`
	CoAPort     string `json:"coa_port"`
	Timeout     string `json:"timeout"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

// NASHotspotConfigInput stores billing-owned Mikrotik deployment settings. The
// FreeRADIUS NAS record intentionally does not carry these fields.
type NASHotspotConfigInput struct {
	RadiusAPIURL     string `json:"radius_api_url" binding:"max=255"`
	RadiusAPIKey     string `json:"radius_api_key" binding:"max=255"`
	RadiusServerID   *uint  `json:"radius_server_id" binding:"omitempty,min=1"`
	RadiusIP         string `json:"radius_ip" binding:"max=128"`
	FrontendURL      string `json:"frontend_url" binding:"max=255"`
	BackendURL       string `json:"backend_url" binding:"max=255"`
	FrontendHost     string `json:"frontend_host" binding:"max=128"`
	CoAPort          string `json:"coa_port" binding:"max=10"`
	WANInterface     string `json:"wan_interface" binding:"max=60"`
	HotspotInterface string `json:"hotspot_interface" binding:"max=60"`
	BridgePorts      string `json:"bridge_ports" binding:"max=200"`
	HotspotNetwork   string `json:"hotspot_network" binding:"max=64"`
	HotspotGateway   string `json:"hotspot_gateway" binding:"max=64"`
	HotspotPoolRange string `json:"hotspot_pool_range" binding:"max=128"`
	HotspotDNS       string `json:"hotspot_dns" binding:"max=128"`
}

// NASOutput is the admin-facing NAS view: the RADIUS client fields plus local
// hotspot deployment settings used by script generation.
type NASOutput struct {
	ID            uint                   `json:"id"`
	NASName       string                 `json:"nasname"`
	ShortName     string                 `json:"shortname"`
	Type          string                 `json:"type"`
	Ports         *int                   `json:"ports"`
	Secret        string                 `json:"secret"`
	Server        string                 `json:"server"`
	Community     string                 `json:"community"`
	Description   string                 `json:"description"`
	HotspotConfig NASHotspotConfigOutput `json:"hotspot_config"`
}

// NASHotspotConfigOutput is the persisted script-generation profile for one
// NAS. Empty radius/frontend hosts mean the frontend can derive local defaults.
type NASHotspotConfigOutput struct {
	RadiusAPIURL     string `json:"radius_api_url"`
	RadiusAPIKey     string `json:"radius_api_key"`
	RadiusServerID   *uint  `json:"radius_server_id"`
	RadiusIP         string `json:"radius_ip"`
	FrontendURL      string `json:"frontend_url"`
	BackendURL       string `json:"backend_url"`
	FrontendHost     string `json:"frontend_host"`
	CoAPort          string `json:"coa_port"`
	WANInterface     string `json:"wan_interface"`
	HotspotInterface string `json:"hotspot_interface"`
	BridgePorts      string `json:"bridge_ports"`
	HotspotNetwork   string `json:"hotspot_network"`
	HotspotGateway   string `json:"hotspot_gateway"`
	HotspotPoolRange string `json:"hotspot_pool_range"`
	HotspotDNS       string `json:"hotspot_dns"`
}
