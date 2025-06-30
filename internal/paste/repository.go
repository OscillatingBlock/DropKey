package paste

import (
	"context"
	"log/slog"

	"github.com/uptrace/bun"
)

type PasteRepository interface {
	Create(ctx context.Context, paste *Paste) error
	GetById(ctx context.Context, id string) (*Paste, error)
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
	_, err := r.db.NewCreateTable().Model(paste).Exec(ctx)
	if err != nil {
		slog.Error("Error while inserting paste", "pasteid", paste.ID)
		return err
	}
	return nil
}

func (r *pasteRepository) GetById(ctx context.Context, id string) (*Paste, error) {
	var paste Paste
	err := r.db.NewSelect().Model(&paste).Where("id = ?", id).Scan(ctx)
	if err != nil {
		slog.Error("Error while getting paste", "pasteid", id, "error", err)
		return nil, err
	}
	return &paste, nil
}
