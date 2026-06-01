// Package models maps the standard FreeRADIUS SQL tables to Go structs. The
// schema itself is owned by the goose migrations under /migrations.
package models

import "time"

// RadCheck holds per-user check attributes (e.g. Cleartext-Password, Expiration).
type RadCheck struct {
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	Username  string `gorm:"column:username" json:"username"`
	Attribute string `gorm:"column:attribute" json:"attribute"`
	Op        string `gorm:"column:op" json:"op"`
	Value     string `gorm:"column:value" json:"value"`
}

// TableName maps to the FreeRADIUS table.
func (RadCheck) TableName() string { return "radcheck" }

// RadReply holds per-user reply attributes.
type RadReply struct {
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	Username  string `gorm:"column:username" json:"username"`
	Attribute string `gorm:"column:attribute" json:"attribute"`
	Op        string `gorm:"column:op" json:"op"`
	Value     string `gorm:"column:value" json:"value"`
}

func (RadReply) TableName() string { return "radreply" }

// RadGroupCheck holds per-group check attributes (e.g. Simultaneous-Use).
type RadGroupCheck struct {
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	GroupName string `gorm:"column:groupname" json:"groupname"`
	Attribute string `gorm:"column:attribute" json:"attribute"`
	Op        string `gorm:"column:op" json:"op"`
	Value     string `gorm:"column:value" json:"value"`
}

func (RadGroupCheck) TableName() string { return "radgroupcheck" }

// RadGroupReply holds per-group reply attributes (rate limit, timeout, quota).
type RadGroupReply struct {
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	GroupName string `gorm:"column:groupname" json:"groupname"`
	Attribute string `gorm:"column:attribute" json:"attribute"`
	Op        string `gorm:"column:op" json:"op"`
	Value     string `gorm:"column:value" json:"value"`
}

func (RadGroupReply) TableName() string { return "radgroupreply" }

// RadUserGroup maps a user to a group (profile).
type RadUserGroup struct {
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	Username  string `gorm:"column:username" json:"username"`
	GroupName string `gorm:"column:groupname" json:"groupname"`
	Priority  int    `gorm:"column:priority" json:"priority"`
}

func (RadUserGroup) TableName() string { return "radusergroup" }

// RadAcct is an accounting record (one row per session). Mapped read-mostly.
type RadAcct struct {
	RadAcctID        uint64     `gorm:"column:radacctid;primaryKey" json:"radacctid"`
	AcctSessionID    string     `gorm:"column:acctsessionid" json:"acct_session_id"`
	AcctUniqueID     string     `gorm:"column:acctuniqueid" json:"acct_unique_id"`
	Username         string     `gorm:"column:username" json:"username"`
	NASIPAddress     string     `gorm:"column:nasipaddress" json:"nas_ip_address"`
	NASPortID        string     `gorm:"column:nasportid" json:"nas_port_id"`
	AcctStartTime    *time.Time `gorm:"column:acctstarttime" json:"acct_start_time"`
	AcctUpdateTime   *time.Time `gorm:"column:acctupdatetime" json:"acct_update_time"`
	AcctStopTime     *time.Time `gorm:"column:acctstoptime" json:"acct_stop_time"`
	AcctSessionTime  *uint      `gorm:"column:acctsessiontime" json:"acct_session_time"`
	AcctInputOctets  *int64     `gorm:"column:acctinputoctets" json:"acct_input_octets"`
	AcctOutputOctets *int64     `gorm:"column:acctoutputoctets" json:"acct_output_octets"`
	CallingStationID string     `gorm:"column:callingstationid" json:"calling_station_id"`
	FramedIPAddress  string     `gorm:"column:framedipaddress" json:"framed_ip_address"`
}

func (RadAcct) TableName() string { return "radacct" }

// Nas is a registered network access server (the Mikrotik router).
type Nas struct {
	ID          uint   `gorm:"column:id;primaryKey" json:"id"`
	NASName     string `gorm:"column:nasname" json:"nasname"`
	ShortName   string `gorm:"column:shortname" json:"shortname"`
	Type        string `gorm:"column:type" json:"type"`
	Ports       *int   `gorm:"column:ports" json:"ports"`
	Secret      string `gorm:"column:secret" json:"secret"`
	Server      string `gorm:"column:server" json:"server"`
	Community   string `gorm:"column:community" json:"community"`
	Description string `gorm:"column:description" json:"description"`
}

func (Nas) TableName() string { return "nas" }
