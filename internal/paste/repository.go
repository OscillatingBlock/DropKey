package paste

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"Drop-Key/internal/models"

	"github.com/uptrace/bun"
)

type PasteRepository interface {
	Create(ctx context.Context, paste *models.Paste) error
	GetByID(ctx context.Context, id string) (*models.Paste, error)
	Update(ctx context.Context, paste *models.Paste) error
}

type pasteRepository struct {
	db *bun.DB
}

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

	err := r.db.NewSelect().
		Model(&paste).
		Where("id = ?", id).
		Where("expires_at > ?", time.Now().UTC().Truncate(time.Second)).
		Scan(ctx)
	if err != nil {
		slog.Error("Error while getting paste", "operation", "get", "pasteid", id, "error", err)
		return nil, err
	}
	return &paste, nil
}

func (r *pasteRepository) UpdatePaste(ctx context.Context, paste *models.Paste) error {
	_, err := r.db.NewUpdate().Model(paste).Where("id = ?").Exec(ctx)
	if err != nil {
		slog.Error("Error while updating paste", "operation", "update", "pasteid", paste.ID, "error", err)
		return err
	}
	return nil
}
