package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"Drop-Key/internal/models"
	"Drop-Key/internal/utils"

	"github.com/uptrace/bun"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByPublicKey(ctx context.Context, public_key string) (*models.User, error)
}

type userRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) *userRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		slog.Error("Error while inserting user", "operation", "create", "userid", user.ID)
		return err
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.NewSelect().Model(&user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		slog.Error("Error while getting user", "operation", "get", "userID", id, "error", err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPublicKey(ctx context.Context, public_key string) (*models.User, error) {
	var user models.User
	err := r.db.NewSelect().Model(&user).Where("public_key = ?", public_key).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.ErrUserNotFound
		}
		slog.Error("Error while getting user", "opearation", "get", "public_key", public_key[:8], "error", err)
		return nil, utils.ErrUserNotFound
	}
	return &user, nil
}
