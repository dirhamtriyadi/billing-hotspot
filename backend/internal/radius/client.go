// Package radius is the backend's client for the radius-api microservice. The
// backend never touches the FreeRADIUS database directly; it provisions and
// revokes hotspot credentials exclusively through these HTTP calls.
package radius

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
)

// Client talks to the radius-api over HTTP, authenticated with a shared API key.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClientWith builds a radius-api client from explicit endpoint values. It is
// used for multi-branch deployments where every branch has its own radius-api.
func NewClientWith(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: timeout},
	}
}

// Profile is a package's radius attribute set (maps to a FreeRADIUS group).
type Profile struct {
	Name            string `json:"name"`
	RateDownKbps    int    `json:"rate_down_kbps"`
	RateUpKbps      int    `json:"rate_up_kbps"`
	BurstEnabled    bool   `json:"burst_enabled"`
	SessionTimeout  int    `json:"session_timeout_secs"`
	DataQuotaMB     int64  `json:"data_quota_mb"`
	SimultaneousUse int    `json:"simultaneous_use"`
}

// User is a single hotspot credential to provision.
type User struct {
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	Profile   string     `json:"profile"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// UpsertProfile creates or updates the radius group for a package.
func (c *Client) UpsertProfile(ctx context.Context, p Profile) error {
	return c.do(ctx, http.MethodPost, "/api/v1/profiles", p, nil)
}

// CreateUser provisions a voucher credential into FreeRADIUS.
func (c *Client) CreateUser(ctx context.Context, u User) error {
	return c.do(ctx, http.MethodPost, "/api/v1/users", u, nil)
}

// CreateUsers bulk-provisions many credentials in a single request.
func (c *Client) CreateUsers(ctx context.Context, users []User) error {
	if len(users) == 0 {
		return nil
	}
	return c.do(ctx, http.MethodPost, "/api/v1/users/bulk", map[string]interface{}{"users": users}, nil)
}

// DeleteUser removes a credential (e.g. when a voucher is revoked/expired).
func (c *Client) DeleteUser(ctx context.Context, username string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/users/"+username, nil, nil)
}

// DisconnectUser forcibly ends a user's active sessions via CoA/PoD.
func (c *Client) DisconnectUser(ctx context.Context, username string) error {
	return c.do(ctx, http.MethodPost, "/api/v1/users/"+username+"/disconnect", nil, nil)
}

// NAS is a registered network access server (the Mikrotik router that talks
// RADIUS). Mirrors the radius-api FreeRADIUS `nas` table row.
type NAS struct {
	ID          uint   `json:"id"`
	NASName     string `json:"nasname"`
	ShortName   string `json:"shortname"`
	Type        string `json:"type"`
	Ports       *int   `json:"ports"`
	Secret      string `json:"secret"`
	Server      string `json:"server"`
	Community   string `json:"community"`
	Description string `json:"description"`
}

// NASInput is the payload to register or update a NAS, keyed by nasname.
type NASInput struct {
	NASName     string `json:"nasname"`
	ShortName   string `json:"shortname"`
	Secret      string `json:"secret"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Ports       *int   `json:"ports"`
}

// ListNAS returns every registered NAS / RADIUS client.
func (c *Client) ListNAS(ctx context.Context) ([]NAS, error) {
	var out []NAS
	if err := c.do(ctx, http.MethodGet, "/api/v1/nas", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpsertNAS registers or updates a NAS and returns the stored record.
func (c *Client) UpsertNAS(ctx context.Context, in NASInput) (*NAS, error) {
	var out NAS
	if err := c.do(ctx, http.MethodPost, "/api/v1/nas", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteNAS removes a NAS client by its id.
func (c *Client) DeleteNAS(ctx context.Context, id uint) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/nas/%d", id), nil, nil)
}

// do performs a request, mapping transport/HTTP failures to AppErrors and
// optionally decoding the envelope's data field into out.
func (c *Client) do(ctx context.Context, method, path string, body, out interface{}) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return apperror.Internal("failed to encode radius request").WithCause(err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return apperror.Internal("failed to build radius request").WithCause(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return apperror.ServiceUnavailable("RADIUS service is unreachable").WithCause(err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		msg := fmt.Sprintf("radius-api returned %d", resp.StatusCode)
		// Surface the upstream message when present.
		var env struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(raw, &env) == nil && env.Message != "" {
			msg = "radius-api: " + env.Message
		}
		return apperror.ServiceUnavailable(msg).WithCause(fmt.Errorf("body: %s", string(raw)))
	}

	if out != nil {
		var env struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			return apperror.Internal("failed to decode radius response").WithCause(err)
		}
		if len(env.Data) > 0 {
			if err := json.Unmarshal(env.Data, out); err != nil {
				return apperror.Internal("failed to decode radius data").WithCause(err)
			}
		}
	}
	return nil
}
