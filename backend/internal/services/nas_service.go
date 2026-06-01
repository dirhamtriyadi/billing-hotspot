package services

import (
	"context"

	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/radius"
)

// NasService manages NAS / RADIUS clients (the Mikrotik routers). The records
// live in the FreeRADIUS database owned by the radius-api microservice, so this
// service is a thin authenticated proxy to it.
type NasService struct {
	radius *radius.Client
}

// NewNasService builds a NasService.
func NewNasService(client *radius.Client) *NasService {
	return &NasService{radius: client}
}

// List returns every registered NAS.
func (s *NasService) List(ctx context.Context) ([]radius.NAS, error) {
	return s.radius.ListNAS(ctx)
}

// Upsert registers or updates a NAS keyed by nasname.
func (s *NasService) Upsert(ctx context.Context, in dto.NASInput) (*radius.NAS, error) {
	return s.radius.UpsertNAS(ctx, radius.NASInput{
		NASName:     in.NASName,
		ShortName:   in.ShortName,
		Secret:      in.Secret,
		Type:        in.Type,
		Description: in.Description,
		Ports:       in.Ports,
	})
}

// Delete removes a NAS by id.
func (s *NasService) Delete(ctx context.Context, id uint) error {
	return s.radius.DeleteNAS(ctx, id)
}
