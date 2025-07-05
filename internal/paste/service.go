package paste

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"time"

	"Drop-Key/internal/models"
	"Drop-Key/internal/user"
	"Drop-Key/internal/utils"

	"github.com/google/uuid"
)

type PasteService interface {
	Create(ctx context.Context, paste *models.Paste, expires_in int) (string, error)
	GetByID(ctx context.Context, id string) (*models.Paste, error)
	Update(ctx context.Context, paste *models.Paste) error
	GetByPublicKey(ctx context.Context, publicKey string) ([]*models.Paste, error)
}

type pasteService struct {
	repo     PasteRepository
	userRepo user.UserRepository
}

func NewPasteService(repo PasteRepository, userRepo user.UserRepository) *pasteService {
	return &pasteService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (p *pasteService) Create(ctx context.Context, paste *models.Paste, expires_in int) (string, error) {
	expires_at := time.Now().UTC().Add(time.Second * time.Duration(expires_in)).Truncate(time.Second)
	if time.Now().UTC().Compare(expires_at) != -1 {
		return "", utils.ErrPasteExpiredAlready
	}
	if time.Now().UTC().Add(time.Second*604800).Compare(expires_at) != 1 {
		return "", utils.ErrPasteExpiryTooLong
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	if paste.Ciphertext == "" {
		return "", utils.ErrPasteEmptyCiphertext
	}
	ciphertext, err := base64.StdEncoding.DecodeString(paste.Ciphertext)
	if err != nil {
		return "", utils.ErrPasteInvalidCiphertext
	}

	if paste.Signature == "" {
		return "", utils.ErrPasteEmptySignature
	}
	signature, err := base64.StdEncoding.DecodeString(paste.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return "", utils.ErrPasteInvalidSignature
	}

	if paste.PublicKey == "" {
		return "", utils.ErrPasteInvalidPublicKey
	}
	publicKey, err := base64.StdEncoding.DecodeString(paste.PublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return "", utils.ErrPasteInvalidPublicKey
	}

	_, err = p.userRepo.GetByPublicKey(ctx, paste.PublicKey)
	if err != nil {
		return "", utils.ErrPasteUserNotFound
	}

	if !ed25519.Verify(publicKey, ciphertext, signature) {
		return "", utils.ErrPasteInvalidSignatureVerification
	}

	paste.ID = uuid.NewString()
	paste.ExpiresAt = expires_at

	err = p.repo.Create(ctx, paste)
	if err != nil {
		return "", err
	}
	return paste.ID, nil
}

func (p *pasteService) GetByID(ctx context.Context, id string) (*models.Paste, error) {
	if id == "" {
		return nil, utils.ErrPasteInvalidID
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, utils.ErrPasteInvalidID
	}
	paste, err := p.repo.GetByID(ctx, id)
	if err != nil {
		if err == utils.ErrPasteExpiredAlready {
			return nil, err
		}
		return nil, utils.ErrPasteNotFound
	}
	return paste, nil
}

func (p *pasteService) GetByPublicKey(ctx context.Context, publicKey string) ([]*models.Paste, error) {
	if publicKey == "" {
		return nil, utils.ErrEmptyPublicKey
	}
	base64publicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil || len(base64publicKey) != ed25519.PublicKeySize {
		return nil, utils.ErrInvalidPublicKey
	}

	userWithPublicKey, err := p.userRepo.GetByPublicKey(ctx, publicKey)
	if err != nil {
		return nil, utils.ErrUserNotFoundForPublicKey
	}
	if publicKey != userWithPublicKey.PublicKey {
		return nil, utils.ErrUnauthorizedAccess
	}

	pastes, err := p.repo.GetByPublicKey(ctx, publicKey)
	if err != nil {
		return nil, utils.ErrPasteNotFound
	}
	return pastes, nil
}

func (p *pasteService) Update(ctx context.Context, paste *models.Paste) error {
	if paste.ID == "" {
		return utils.ErrPasteInvalidID
	}
	if _, err := uuid.Parse(paste.ID); err != nil {
		return utils.ErrPasteInvalidID
	}
	if paste.Ciphertext == "" {
		return utils.ErrPasteEmptyCiphertext
	}
	ciphertext, err := base64.StdEncoding.DecodeString(paste.Ciphertext)
	if err != nil {
		return utils.ErrPasteInvalidCiphertext
	}
	if paste.Signature == "" {
		return utils.ErrPasteEmptySignature
	}
	signature, err := base64.StdEncoding.DecodeString(paste.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return utils.ErrPasteInvalidSignature
	}
	if paste.PublicKey == "" {
		return utils.ErrPasteInvalidPublicKey
	}
	publicKey, err := base64.StdEncoding.DecodeString(paste.PublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return utils.ErrPasteInvalidPublicKey
	}
	if !ed25519.Verify(publicKey, ciphertext, signature) {
		return utils.ErrPasteInvalidSignatureVerification
	}
	return p.repo.Update(ctx, paste)
}
