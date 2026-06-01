package services

import (
	"time"

	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"github.com/dirhamt/billing-hotspot/backend/internal/radius"
)

// buildProfile maps a package to the radius-api profile payload.
func buildProfile(p models.Package) radius.Profile {
	return radius.Profile{
		Name:            p.Profile,
		RateDownKbps:    p.RateDownKbps,
		RateUpKbps:      p.RateUpKbps,
		BurstEnabled:    p.BurstEnabled,
		SessionTimeout:  p.SessionTimeoutSecs,
		DataQuotaMB:     p.DataQuotaMB,
		SimultaneousUse: p.SimultaneousUse,
	}
}

// computeExpiry returns the validity deadline for a voucher activated at `from`,
// or nil for packages with unlimited validity.
func computeExpiry(p models.Package, from time.Time) *time.Time {
	d := p.Validity()
	if d <= 0 {
		return nil
	}
	t := from.Add(d)
	return &t
}
