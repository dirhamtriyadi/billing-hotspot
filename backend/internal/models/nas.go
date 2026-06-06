package models

// NASHotspotConfig stores billing-owned deployment settings for one Mikrotik
// NAS. It is keyed by nasname because the authoritative FreeRADIUS NAS row
// lives in the radius-api database.
type NASHotspotConfig struct {
	Base
	NASName          string `gorm:"column:nasname;size:128;uniqueIndex;not null" json:"nasname"`
	RadiusIP         string `gorm:"size:128;not null;default:''" json:"radius_ip"`
	FrontendHost     string `gorm:"size:128;not null;default:''" json:"frontend_host"`
	CoAPort          string `gorm:"size:10;not null;default:'3799'" json:"coa_port"`
	WANInterface     string `gorm:"size:60;not null;default:'ether1'" json:"wan_interface"`
	HotspotInterface string `gorm:"size:60;not null;default:'bridge-hotspot'" json:"hotspot_interface"`
	BridgePorts      string `gorm:"size:200;not null;default:'wlan1,wlan2'" json:"bridge_ports"`
	HotspotNetwork   string `gorm:"size:64;not null;default:'10.5.50.0/24'" json:"hotspot_network"`
	HotspotGateway   string `gorm:"size:64;not null;default:'10.5.50.1'" json:"hotspot_gateway"`
	HotspotPoolRange string `gorm:"size:128;not null;default:'10.5.50.10-10.5.50.254'" json:"hotspot_pool_range"`
	HotspotDNS       string `gorm:"size:128;not null;default:'8.8.8.8,1.1.1.1'" json:"hotspot_dns"`
}
