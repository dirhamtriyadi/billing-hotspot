package models

// RadiusServer is one managed radius-api endpoint, usually one branch-local
// FreeRADIUS management API.
type RadiusServer struct {
	Base
	Name        string `gorm:"size:120;uniqueIndex;not null" json:"name"`
	APIURL      string `gorm:"column:api_url;size:255;not null" json:"api_url"`
	APIKey      string `gorm:"column:api_key;size:255;not null;default:''" json:"api_key"`
	RadiusIP    string `gorm:"column:radius_ip;size:128;not null;default:''" json:"radius_ip"`
	CoAPort     string `gorm:"column:coa_port;size:10;not null;default:'3799'" json:"coa_port"`
	Timeout     string `gorm:"size:20;not null;default:'10s'" json:"timeout"`
	Description string `gorm:"size:200;not null;default:''" json:"description"`
	IsDefault   bool   `gorm:"column:is_default;not null;default:false" json:"is_default"`
}

// NASHotspotConfig stores billing-owned deployment settings for one Mikrotik
// NAS, including the local radius-api endpoint for that branch. The FreeRADIUS
// NAS row is still synced into the branch radius-api database, but the billing
// app keeps this local copy so it can support multiple independent RADIUS
// servers.
type NASHotspotConfig struct {
	Base
	NASName          string `gorm:"column:nasname;size:128;uniqueIndex;not null" json:"nasname"`
	ShortName        string `gorm:"column:shortname;size:32;not null;default:''" json:"shortname"`
	Type             string `gorm:"column:type;size:30;not null;default:'mikrotik'" json:"type"`
	Ports            *int   `gorm:"column:ports" json:"ports"`
	Secret           string `gorm:"column:secret;size:60;not null;default:''" json:"secret"`
	Description      string `gorm:"column:description;size:200;not null;default:''" json:"description"`
	RadiusAPIURL     string `gorm:"column:radius_api_url;size:255;not null;default:''" json:"radius_api_url"`
	RadiusAPIKey     string `gorm:"column:radius_api_key;size:255;not null;default:''" json:"radius_api_key"`
	RadiusServerID   *uint  `gorm:"column:radius_server_id" json:"radius_server_id"`
	RadiusIP         string `gorm:"column:radius_ip;size:128;not null;default:''" json:"radius_ip"`
	FrontendURL      string `gorm:"column:frontend_url;size:255;not null;default:''" json:"frontend_url"`
	BackendURL       string `gorm:"column:backend_url;size:255;not null;default:''" json:"backend_url"`
	FrontendHost     string `gorm:"size:128;not null;default:''" json:"frontend_host"`
	CoAPort          string `gorm:"column:coa_port;size:10;not null;default:'3799'" json:"coa_port"`
	WANInterface     string `gorm:"column:wan_interface;size:60;not null;default:'ether1'" json:"wan_interface"`
	HotspotInterface string `gorm:"size:60;not null;default:'bridge-hotspot'" json:"hotspot_interface"`
	BridgePorts      string `gorm:"size:200;not null;default:'wlan1,wlan2'" json:"bridge_ports"`
	HotspotNetwork   string `gorm:"size:64;not null;default:'10.5.50.0/24'" json:"hotspot_network"`
	HotspotGateway   string `gorm:"size:64;not null;default:'10.5.50.1'" json:"hotspot_gateway"`
	HotspotPoolRange string `gorm:"size:128;not null;default:'10.5.50.10-10.5.50.254'" json:"hotspot_pool_range"`
	HotspotDNS       string `gorm:"column:hotspot_dns;size:128;not null;default:'8.8.8.8,1.1.1.1'" json:"hotspot_dns"`
}
