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

func (p *pasteService) Create(ctx context.Context, paste *models.Paste, expires_in int) (id string, err error) {
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

	publicKey, err := base64.StdEncoding.DecodeString(paste.PublicKey)
	if err != nil {
		return "", fmt.Errorf("Invalid paste, publicKey not base64 encrypted.")
	}

	_, err = p.userRepo.GetByPublicKey(ctx, paste.PublicKey)
	if err != nil {
		return "", fmt.Errorf("Could not insert paste, user does not exist.")
	}

	if len([]byte(publicKey)) != ed25519.PublicKeySize {
		return "", fmt.Errorf("Invalid size of public key")
	}

	if len([]byte(signature)) != ed25519.SignatureSize {
		return "", fmt.Errorf("Invalid size of signature")
	}

	ok := ed25519.Verify([]byte(publicKey), []byte(ciphertext), []byte(signature))
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
