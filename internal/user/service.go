package user

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"Drop-Key/internal/models"
	"Drop-Key/internal/utils"

	"github.com/google/uuid"
)

type UserService interface {
	Create(ctx context.Context, user *models.User) (string, error)
	Authenticate(ctx context.Context, userID, signature, challenge string) (bool, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByPublicKey(ctx context.Context, publicKey string) (*models.User, error)
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *userService {
	return &userService{
		repo: repo,
	}
}

func (u *userService) Create(ctx context.Context, user *models.User) (string, error) {
	if user.PublicKey == "" {
		return "", utils.WrapError(utils.ErrEmptyPublicKey, "Cannot create user, error")
	}
	fmt.Println(user.PublicKey)

	existingUser, err := u.repo.GetByPublicKey(ctx, user.PublicKey)
	if err == nil && existingUser != nil {
		return "", utils.WrapError(utils.ErrDuplicatePublicKey, "Cannot create user, error")
	}
	publicKey, err := base64.StdEncoding.DecodeString(user.PublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return "", utils.WrapError(utils.ErrInvalidPublicKey, "Cannot create user, error")
	}

	user.ID = uuid.NewString()

	err = u.repo.Create(ctx, user)
	if err != nil {
		return "", utils.WrapError(utils.ErrUserCreationFailed, "Failed to create user")
	}

	return user.ID, nil
}

func (u *userService) Authenticate(ctx context.Context, userID, signature, challenge string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if userID == "" {
		return false, utils.WrapError(utils.ErrEmptyUserID, "Authentication failed, error ")
	}
	if challenge == "" {
		return false, utils.WrapError(utils.ErrValidationError, "Auth failed, challenge empty, error ")
	}

	if signature == "" {
		return false, utils.WrapError(utils.ErrEmptySignature, "Auth failed, error ")
	}
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return false, utils.WrapError(utils.ErrInvalidSignature, "Auth failed, error ")
	}

	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return false, utils.WrapError(utils.ErrUserNotFound, "Auth failed, error ")
	}
	pubKeyBytes, err := base64.StdEncoding.DecodeString(user.PublicKey)
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		return false, utils.WrapError(utils.ErrInvalidPublicKey, "Auth failed, public key invalid")
	}

	decodedChallenge, err := base64.StdEncoding.DecodeString(challenge)
	if err != nil {
		return false, utils.WrapError(utils.ErrValidationError, "Auth failed, invalid base64 challenge")
	}

	ok := ed25519.Verify(pubKeyBytes, decodedChallenge, sig)

	if !ok {
		return false, utils.WrapError(utils.ErrInvalidSignature, "Auth failed, error ")
	}

	return true, nil
}

func (u *userService) GetByPublicKey(ctx context.Context, publicKey string) (*models.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if publicKey == "" {
		return nil, utils.WrapError(utils.ErrEmptyPublicKey, "Cannot get user by publicKey, error ")
	}
	pub, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return nil, utils.WrapError(utils.ErrInvalidPublicKey, "Cannot get user by publicKey, error ")
	}

	user, err := u.repo.GetByPublicKey(ctx, publicKey)
	if err != nil {
		return nil, utils.WrapError(utils.ErrUserNotFound, "Cannot get user by public key")
	}

	return user, nil
}

func (u *userService) GetByID(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, utils.WrapError(utils.ErrEmptyUserID, "cannot get user, error")
	}

	_, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.WrapError(utils.ErrInvalidUserID, "cannot get user, error")
	}

	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, utils.WrapError(utils.ErrUserNotFound, "cannot get user by id, error")
	}

	return user, nil
}
