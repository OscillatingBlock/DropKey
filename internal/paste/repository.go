package paste

import (
	"context"
	"log/slog"

	"github.com/uptrace/bun"
)

type PasteRepository interface {
	Create(ctx context.Context, paste *Paste) error
	GetByID(ctx context.Context, id string) (*Paste, error)
	Update(ctx context.Context, paste *Paste) error
}

type pasteRepository struct {
	db *bun.DB
}

func NewPasteRepository(db *bun.DB) *pasteRepository {
	return &pasteRepository{
		db: db,
	}
}

func (r *pasteRepository) Create(ctx context.Context, paste *Paste) error {
	_, err := r.db.NewInsert().Model(paste).Exec(ctx)
	if err != nil {
		slog.Error("Error while inserting paste", "operation", "Create", "pasteid", paste.ID, "error", err)
		return err
	}
	return nil
}

func (r *pasteRepository) GetByID(ctx context.Context, id string) (*Paste, error) {
	var paste Paste
	err := r.db.NewSelect().Model(&paste).Where("id = ? AND expires_at > NOW()", id).Scan(ctx)
	if err != nil {
		slog.Error("Error while getting paste", "pasteid", id, "error", err)
		return nil, err
	}
	return &paste, nil
}

func (r *pasteRepository) UpdatePaste(ctx context.Context, paste *Paste) error {
	_, err := r.db.NewUpdate().Model(paste).Where("id = ?").Exec(ctx)
	if err != nil {
		slog.Error("Error while updating paste", "operation", "update", "pasteid", paste.ID, "error", err)
		return err
	}
	return nil
}
