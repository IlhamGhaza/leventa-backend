package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/pkg/logger"
)

// AuditService handles audit log creation.
type AuditService struct {
	auditRepo repository.AuditLogRepository
}

// NewAuditService creates a new audit service.
func NewAuditService(auditRepo repository.AuditLogRepository) *AuditService {
	return &AuditService{auditRepo: auditRepo}
}

// Log creates an audit log entry (runs in background goroutine).
func (s *AuditService) Log(userID *uuid.UUID, action, module string, entityID *uuid.UUID, metadata interface{}, ipAddress, userAgent string) {
	go func() {
		var metadataJSON []byte
		if metadata != nil {
			metadataJSON, _ = json.Marshal(metadata)
		}

		log := &entity.AuditLog{
			UserID:    userID,
			Action:    action,
			Module:    module,
			EntityID:  entityID,
			Metadata:  metadataJSON,
			IPAddress: ipAddress,
			UserAgent: userAgent,
		}
		if err := s.auditRepo.Create(context.Background(), log); err != nil {
			logger.Error().Err(err).
				Str("action", action).
				Str("module", module).
				Msg("Failed to create audit log")
		}
	}()
}

// LogSync writes an audit entry using ctx. Inside a Transactor.Do callback the
// entry joins the surrounding transaction, so it commits or rolls back with the
// business change. Metadata must never contain secrets or owner PII.
func (s *AuditService) LogSync(ctx context.Context, userID *uuid.UUID, action, module string, entityID *uuid.UUID, metadata interface{}, ipAddress, userAgent string) error {
	var metadataJSON []byte
	if metadata != nil {
		var err error
		if metadataJSON, err = json.Marshal(metadata); err != nil {
			return err
		}
	}
	return s.auditRepo.Create(ctx, &entity.AuditLog{
		UserID:    userID,
		Action:    action,
		Module:    module,
		EntityID:  entityID,
		Metadata:  metadataJSON,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}
