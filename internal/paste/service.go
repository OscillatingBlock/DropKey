package paste

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"time"

	"Drop-Key/internal/models"
	"Drop-Key/internal/user"

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
	expires_at := time.Now().UTC().Add(time.Duration(expires_in))
	if time.Now().UTC().Compare(expires_at) != -1 {
		return "", fmt.Errorf("Invalid paste, expired already.")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if time.Now().UTC().Add(time.Second*604800).Compare(expires_at) != 1 {
		return "", fmt.Errorf("Expiry date of paste too long.")
	}

	if paste.Ciphertext == "" {
		return "", fmt.Errorf("Invalid paste, empty Ciphertext")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(paste.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("Invalid paste, Ciphertext not base64 encrypted.")
	}

	if paste.Signature == "" {
		return "", fmt.Errorf("Invalid paste, empty Signature")
	}
	signature, err := base64.StdEncoding.DecodeString(paste.Signature)
	if err != nil {
		return "", fmt.Errorf("Invalid paste, Signature not base64 encrypted.")
	}
	if len(signature) != ed25519.SignatureSize {
		return "", fmt.Errorf("Invalid size of signature")
	}

	if paste.PublicKey == "" {
		return "", fmt.Errorf("Invalid Paste, empty PublicKey.")
	}
	publicKey, err := base64.StdEncoding.DecodeString(paste.PublicKey)
	if err != nil {
		return "", fmt.Errorf("Invalid paste, publicKey not base64 encrypted.")
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return "", fmt.Errorf("Invalid size of public key")
	}

	_, err = p.userRepo.GetByPublicKey(ctx, paste.PublicKey)
	if err != nil {
		return "", fmt.Errorf("Could not insert paste, user does not exist.")
	}

	ok := ed25519.Verify(publicKey, ciphertext, signature)
	if !ok {
		return "", fmt.Errorf("Invalid Signature.")
	}

	paste.ID = uuid.NewString()
	paste.ExpiresAt = expires_at
	err = p.repo.Create(ctx, paste)
	if err != nil {
		return "", err
	}
	return paste.ID, nil
}

func (p *pasteService) GetByID(ctx context.Context, id string) (paste *models.Paste, err error) {
	if id == "" {
		return nil, fmt.Errorf("Cannot get paste, Invalid id, id = %v", id)
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid paste id: %v", err)
	}
	paste, err = p.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if paste == nil {
		return nil, fmt.Errorf("Paste not found, paste id = %v", id)
	}
	return paste, err
}

func (p *pasteService) GetByPublicKey(ctx context.Context, publicKey string) ([]*models.Paste, error) {
	if publicKey == "" {
		return nil, fmt.Errorf("Cannot get pastes, Empty publicKey")
	}

	base64publicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("Cannot get pastes, Invalid publicKey, not base64 encrypted.")
	}
	if len(base64publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("Invalid size of public key")
	}

	userWithPublicKey, err := p.userRepo.GetByPublicKey(ctx, publicKey)
	if err != nil {
		return nil, fmt.Errorf("Cannot get pastes, user with this PublicKey not found.")
	}

	if publicKey != userWithPublicKey.PublicKey {
		return nil, fmt.Errorf("Cannot get pastes, UNauthorized access.")
	}

	pastes, err := p.repo.GetByPublicKey(ctx, publicKey)
	if err != nil {
		return nil, err
	}

	return pastes, nil
}
