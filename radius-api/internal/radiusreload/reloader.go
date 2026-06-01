// Package radiusreload restarts the FreeRADIUS container so it re-reads its SQL
// client list (the `nas` table) after a NAS is added, changed, or removed.
//
// FreeRADIUS loads SQL clients (read_clients) only at startup, so a new router
// with its own unique secret is not recognised until the server reloads. Rather
// than give radius-api a shell or the docker CLI, we talk to the Docker Engine
// API directly over its unix socket: POST /containers/<name>/restart.
//
// The reloader is best-effort and fully optional: if no container name or socket
// is configured (e.g. local `go run`), Reload is a no-op that logs a notice, so
// nothing breaks outside Docker.
package radiusreload

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Reloader triggers a FreeRADIUS container restart via the Docker socket.
type Reloader struct {
	container string
	socket    string
	http      *http.Client
	enabled   bool
}

// New builds a Reloader. When container or socket is empty it is disabled and
// Reload becomes a logged no-op.
func New(container, socket string) *Reloader {
	r := &Reloader{container: container, socket: socket}
	if container == "" || socket == "" {
		return r
	}
	r.http = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socket)
			},
		},
	}
	r.enabled = true
	return r
}

// Enabled reports whether the reloader can actually talk to Docker.
func (r *Reloader) Enabled() bool { return r.enabled }

// Reload restarts the FreeRADIUS container. It is safe to call when disabled
// (logs and returns nil). Errors are returned so callers may log them, but a
// failed reload should not fail the originating NAS mutation.
func (r *Reloader) Reload(ctx context.Context) error {
	if !r.enabled {
		slog.Info("radius reload skipped (not configured); restart FreeRADIUS manually to pick up new NAS clients")
		return nil
	}

	// Docker Engine API: POST /containers/{id}/restart. Host is ignored for a
	// unix-socket transport but must be a valid URL.
	url := fmt.Sprintf("http://docker/containers/%s/restart?t=3", r.container)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("build docker restart request: %w", err)
	}

	resp, err := r.http.Do(req)
	if err != nil {
		return fmt.Errorf("docker restart call failed (is %s mounted?): %w", r.socket, err)
	}
	defer resp.Body.Close()

	// 204 = restarted; 304 = already in desired state; 404 = unknown container.
	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusNotModified:
		slog.Info("FreeRADIUS reloaded", slog.String("container", r.container))
		return nil
	default:
		return fmt.Errorf("docker restart returned status %d for container %q", resp.StatusCode, r.container)
	}
}
