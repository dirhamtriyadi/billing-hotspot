// Package coa sends RADIUS Disconnect-Request (Packet of Disconnect) packets to
// a NAS, used to forcibly end a hotspot user's active session.
package coa

import (
	"context"
	"fmt"
	"net"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// Disconnector sends PoD packets to NAS devices.
type Disconnector struct {
	port    int
	timeout time.Duration
}

// New builds a Disconnector. port is the NAS CoA/PoD listen port (Mikrotik
// defaults to 3799).
func New(port int, timeout time.Duration) *Disconnector {
	return &Disconnector{port: port, timeout: timeout}
}

// Target identifies the session to disconnect on a specific NAS.
type Target struct {
	NASIP     string
	Secret    string
	Username  string
	SessionID string
	FramedIP  string
}

// Disconnect sends a Disconnect-Request and waits for an ACK.
func (d *Disconnector) Disconnect(ctx context.Context, t Target) error {
	packet := radius.New(radius.CodeDisconnectRequest, []byte(t.Secret))

	if t.Username != "" {
		_ = rfc2865.UserName_SetString(packet, t.Username)
	}
	if t.SessionID != "" {
		_ = rfc2866.AcctSessionID_SetString(packet, t.SessionID)
	}
	if ip := net.ParseIP(t.NASIP); ip != nil {
		_ = rfc2865.NASIPAddress_Set(packet, ip)
	}
	if ip := net.ParseIP(t.FramedIP); ip != nil {
		_ = rfc2865.FramedIPAddress_Set(packet, ip)
	}

	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", t.NASIP, d.port)
	resp, err := radius.Exchange(ctx, packet, addr)
	if err != nil {
		return fmt.Errorf("coa: exchange with %s: %w", addr, err)
	}
	if resp.Code != radius.CodeDisconnectACK {
		return fmt.Errorf("coa: NAS %s responded %v (expected Disconnect-ACK)", addr, resp.Code)
	}
	return nil
}
