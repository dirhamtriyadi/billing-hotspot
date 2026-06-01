// Package radiussql implements the management operations against the FreeRADIUS
// SQL tables: provisioning users, defining group profiles, listing live
// sessions, managing NAS clients and disconnecting users via CoA.
package radiussql

import (
	"context"
	"errors"
	"fmt"

	"github.com/dirhamt/billing-hotspot/radius-api/internal/coa"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/dto"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/models"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a requested user/record does not exist.
var ErrNotFound = errors.New("not found")

// FreeRADIUS attribute names provisioned by this service.
const (
	attrPassword       = "Cleartext-Password"
	attrExpiration     = "Expiration"
	attrRateLimit      = "Mikrotik-Rate-Limit"
	attrSessionTimeout = "Session-Timeout"
	attrTotalLimit     = "Mikrotik-Total-Limit"
	attrTotalLimitGiga = "Mikrotik-Total-Limit-Gigawords"
	attrSimultaneous   = "Simultaneous-Use"

	// expirationLayout matches the format FreeRADIUS parses for the date-typed
	// Expiration attribute, e.g. "May 31 2026 14:30:00".
	expirationLayout = "Jan 02 2006 15:04:05"

	gigaword = int64(1) << 32
)

// Service owns the FreeRADIUS database operations.
type Service struct {
	db  *gorm.DB
	coa *coa.Disconnector
}

// New builds a Service.
func New(db *gorm.DB, disconnector *coa.Disconnector) *Service {
	return &Service{db: db, coa: disconnector}
}

// UpsertProfile (re)defines a FreeRADIUS group from package attributes. It is
// idempotent: existing group rows are replaced.
func (s *Service) UpsertProfile(ctx context.Context, p dto.ProfileRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("groupname = ?", p.Name).Delete(&models.RadGroupReply{}).Error; err != nil {
			return err
		}
		if err := tx.Where("groupname = ?", p.Name).Delete(&models.RadGroupCheck{}).Error; err != nil {
			return err
		}

		replies := make([]models.RadGroupReply, 0, 4)
		if p.RateDownKbps > 0 || p.RateUpKbps > 0 {
			replies = append(replies, models.RadGroupReply{
				GroupName: p.Name, Attribute: attrRateLimit, Op: ":=", Value: rateLimit(p),
			})
		}
		if p.SessionTimeout > 0 {
			replies = append(replies, models.RadGroupReply{
				GroupName: p.Name, Attribute: attrSessionTimeout, Op: ":=", Value: fmt.Sprintf("%d", p.SessionTimeout),
			})
		}
		if p.DataQuotaMB > 0 {
			bytes := p.DataQuotaMB * 1024 * 1024
			replies = append(replies, models.RadGroupReply{
				GroupName: p.Name, Attribute: attrTotalLimit, Op: ":=", Value: fmt.Sprintf("%d", bytes%gigaword),
			})
			if giga := bytes / gigaword; giga > 0 {
				replies = append(replies, models.RadGroupReply{
					GroupName: p.Name, Attribute: attrTotalLimitGiga, Op: ":=", Value: fmt.Sprintf("%d", giga),
				})
			}
		}
		if len(replies) > 0 {
			if err := tx.Create(&replies).Error; err != nil {
				return err
			}
		}

		if p.SimultaneousUse > 0 {
			check := models.RadGroupCheck{
				GroupName: p.Name, Attribute: attrSimultaneous, Op: ":=", Value: fmt.Sprintf("%d", p.SimultaneousUse),
			}
			if err := tx.Create(&check).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CreateUser provisions a single credential (idempotent upsert).
func (s *Service) CreateUser(ctx context.Context, u dto.UserRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return upsertUser(tx, u)
	})
}

// CreateUsers bulk-provisions credentials in one transaction.
func (s *Service) CreateUsers(ctx context.Context, users []dto.UserRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		usernames := make([]string, 0, len(users))
		for _, u := range users {
			usernames = append(usernames, u.Username)
		}
		if err := tx.Where("username IN ?", usernames).Delete(&models.RadCheck{}).Error; err != nil {
			return err
		}
		if err := tx.Where("username IN ?", usernames).Delete(&models.RadUserGroup{}).Error; err != nil {
			return err
		}

		checks := make([]models.RadCheck, 0, len(users))
		groups := make([]models.RadUserGroup, 0, len(users))
		for _, u := range users {
			checks = append(checks, models.RadCheck{Username: u.Username, Attribute: attrPassword, Op: ":=", Value: u.Password})
			if u.ExpiresAt != nil {
				checks = append(checks, models.RadCheck{Username: u.Username, Attribute: attrExpiration, Op: ":=", Value: u.ExpiresAt.Format(expirationLayout)})
			}
			groups = append(groups, models.RadUserGroup{Username: u.Username, GroupName: u.Profile, Priority: 1})
		}
		if err := tx.CreateInBatches(&checks, 500).Error; err != nil {
			return err
		}
		return tx.CreateInBatches(&groups, 500).Error
	})
}

// DeleteUser removes a credential's check/reply/group rows (accounting history
// is preserved).
func (s *Service) DeleteUser(ctx context.Context, username string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("username = ?", username).Delete(&models.RadCheck{}).Error; err != nil {
			return err
		}
		if err := tx.Where("username = ?", username).Delete(&models.RadReply{}).Error; err != nil {
			return err
		}
		return tx.Where("username = ?", username).Delete(&models.RadUserGroup{}).Error
	})
}

// GetUser returns a credential's check attributes and groups.
func (s *Service) GetUser(ctx context.Context, username string) (*dto.UserDetail, error) {
	var checks []models.RadCheck
	if err := s.db.WithContext(ctx).Where("username = ?", username).Find(&checks).Error; err != nil {
		return nil, err
	}
	if len(checks) == 0 {
		return nil, ErrNotFound
	}
	var groups []models.RadUserGroup
	if err := s.db.WithContext(ctx).Where("username = ?", username).Find(&groups).Error; err != nil {
		return nil, err
	}
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, g.GroupName)
	}
	return &dto.UserDetail{Username: username, Groups: names, Attributes: checks}, nil
}

// ListSessions returns active accounting sessions (acctstoptime IS NULL),
// optionally filtered by username.
func (s *Service) ListSessions(ctx context.Context, username string) ([]models.RadAcct, error) {
	q := s.db.WithContext(ctx).Model(&models.RadAcct{}).Where("acctstoptime IS NULL")
	if username != "" {
		q = q.Where("username = ?", username)
	}
	var sessions []models.RadAcct
	err := q.Order("acctstarttime DESC").Limit(500).Find(&sessions).Error
	return sessions, err
}

// DisconnectUser sends a PoD for every active session of the user and returns
// the number of sessions successfully disconnected.
func (s *Service) DisconnectUser(ctx context.Context, username string) (int, error) {
	var sessions []models.RadAcct
	if err := s.db.WithContext(ctx).Where("username = ? AND acctstoptime IS NULL", username).Find(&sessions).Error; err != nil {
		return 0, err
	}
	if len(sessions) == 0 {
		return 0, nil
	}

	disconnected := 0
	var lastErr error
	for _, sess := range sessions {
		var nas models.Nas
		if err := s.db.WithContext(ctx).Where("nasname = ?", sess.NASIPAddress).First(&nas).Error; err != nil {
			lastErr = fmt.Errorf("no NAS secret registered for %s: %w", sess.NASIPAddress, err)
			continue
		}
		if err := s.coa.Disconnect(ctx, coa.Target{
			NASIP:     sess.NASIPAddress,
			Secret:    nas.Secret,
			Username:  username,
			SessionID: sess.AcctSessionID,
			FramedIP:  sess.FramedIPAddress,
		}); err != nil {
			lastErr = err
			continue
		}
		disconnected++
	}
	if disconnected == 0 && lastErr != nil {
		return 0, lastErr
	}
	return disconnected, nil
}

// ListNAS returns all registered NAS clients.
func (s *Service) ListNAS(ctx context.Context) ([]models.Nas, error) {
	var list []models.Nas
	err := s.db.WithContext(ctx).Order("id ASC").Find(&list).Error
	return list, err
}

// UpsertNAS registers or updates a NAS keyed by nasname.
func (s *Service) UpsertNAS(ctx context.Context, req dto.NASRequest) (*models.Nas, error) {
	var nas models.Nas
	err := s.db.WithContext(ctx).Where("nasname = ?", req.NASName).First(&nas).Error
	switch {
	case err == nil:
		nas.ShortName = req.ShortName
		nas.Secret = req.Secret
		nas.Type = defaultStr(req.Type, "other")
		nas.Description = req.Description
		nas.Ports = req.Ports
		if err := s.db.WithContext(ctx).Save(&nas).Error; err != nil {
			return nil, err
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		nas = models.Nas{
			NASName:     req.NASName,
			ShortName:   req.ShortName,
			Secret:      req.Secret,
			Type:        defaultStr(req.Type, "other"),
			Description: defaultStr(req.Description, "RADIUS Client"),
			Ports:       req.Ports,
		}
		if err := s.db.WithContext(ctx).Create(&nas).Error; err != nil {
			return nil, err
		}
	default:
		return nil, err
	}
	return &nas, nil
}

// DeleteNAS removes a NAS client by id.
func (s *Service) DeleteNAS(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Delete(&models.Nas{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// upsertUser replaces a single user's check + group rows within a transaction.
func upsertUser(tx *gorm.DB, u dto.UserRequest) error {
	if err := tx.Where("username = ?", u.Username).Delete(&models.RadCheck{}).Error; err != nil {
		return err
	}
	if err := tx.Where("username = ?", u.Username).Delete(&models.RadUserGroup{}).Error; err != nil {
		return err
	}
	checks := []models.RadCheck{
		{Username: u.Username, Attribute: attrPassword, Op: ":=", Value: u.Password},
	}
	if u.ExpiresAt != nil {
		checks = append(checks, models.RadCheck{
			Username: u.Username, Attribute: attrExpiration, Op: ":=", Value: u.ExpiresAt.Format(expirationLayout),
		})
	}
	if err := tx.Create(&checks).Error; err != nil {
		return err
	}
	return tx.Create(&models.RadUserGroup{Username: u.Username, GroupName: u.Profile, Priority: 1}).Error
}

// rateLimit renders the Mikrotik-Rate-Limit value "rxUp/txDown [bursts]".
func rateLimit(p dto.ProfileRequest) string {
	base := fmt.Sprintf("%dk/%dk", p.RateUpKbps, p.RateDownKbps)
	if !p.BurstEnabled {
		return base
	}
	// base-rate burst-rate burst-threshold burst-time(s)
	return fmt.Sprintf("%s %dk/%dk %dk/%dk 8/8",
		base, p.RateUpKbps*2, p.RateDownKbps*2, p.RateUpKbps, p.RateDownKbps)
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
