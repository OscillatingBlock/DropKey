package paste

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"Drop-Key/internal/models"

	"github.com/uptrace/bun"
)

type PasteRepository interface {
	Create(ctx context.Context, paste *models.Paste) error
	GetByID(ctx context.Context, id string) (*models.Paste, error)
	Update(ctx context.Context, paste *models.Paste) error
	GetByPublicKey(ctx context.Context, publicKey string) ([]*models.Paste, error)
}

type pasteRepository struct {
	db *bun.DB
}

var ErrPasteExpired = errors.New("paste has expired")

func NewPasteRepository(db *bun.DB) *pasteRepository {
	return &pasteRepository{
		db: db,
	}
}

func (r *pasteRepository) Create(ctx context.Context, paste *models.Paste) error {
	_, err := r.db.NewInsert().Model(paste).Exec(ctx)
	if err != nil {
		slog.Error("Error while inserting paste", "operation", "Create", "pasteid", paste.ID, "error", err)
		return err
	}
	return nil
}

func (r *pasteRepository) GetByID(ctx context.Context, id string) (*models.Paste, error) {
	var paste models.Paste
	err := r.db.NewSelect().Model(&paste).Where("id = ?", id).Scan(ctx)
	if err != nil {
		slog.Error("Error while getting paste", "operation", "get", "pasteid", id, "error", err)
		return nil, err
	}
	if paste.ExpiresAt.Before(time.Now().UTC().Truncate(time.Second)) {
		slog.Error("Paste has expired", "pasteid", id)
		return nil, ErrPasteExpired
	}
	return &paste, nil
}

func (r *pasteRepository) Update(ctx context.Context, paste *models.Paste) error {
	_, err := r.db.NewUpdate().Model(paste).Where("id = ?", paste.ID).Column("ciphertext", "signature", "public_key", "expires_at").Exec(ctx)
	if err != nil {
		slog.Error("Error while updating paste", "operation", "update", "pasteid", paste.ID, "error", err)
		return err
	}
	return nil
}

func (r *pasteRepository) GetByPublicKey(ctx context.Context, publicKey string) ([]*models.Paste, error) {
	var pastes []*models.Paste

	err := r.db.NewSelect().
		Model(&pastes).
		Where("public_key = ?", publicKey).
		Where("expires_at > ?", time.Now().UTC()).
		Scan(ctx)
	if err != nil {
		slog.Error("Error while getting pastes by public key", "public_key", publicKey, "error", err)
		return nil, err
	}
	return pastes, nil
}
