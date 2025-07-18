package utils

import (
	"errors"
	"fmt"
)

var (
	ErrPasteExpiredAlready               = errors.New("paste has already expired")
	ErrPasteExpiryTooLong                = errors.New("paste expiry date is too long")
	ErrPasteEmptyCiphertext              = errors.New("paste has empty ciphertext")
	ErrPasteInvalidCiphertext            = errors.New("paste ciphertext is not base64 encoded")
	ErrPasteEmptySignature               = errors.New("paste has empty signature")
	ErrPasteInvalidSignature             = errors.New("paste signature is not base64 encoded or has invalid size")
	ErrPasteInvalidPublicKey             = errors.New("paste has empty or invalid public key")
	ErrPasteUserNotFound                 = errors.New("user for paste does not exist")
	ErrPasteInvalidSignatureVerification = errors.New("invalid signature verification")
	ErrPasteInvalidID                    = errors.New("invalid paste ID")
	ErrPasteNotFound                     = errors.New("paste not found")
	ErrPasteInvalidExpiryTime            = errors.New("paste has invalid expiry time")
)

var (
	ErrEmptyPublicKey           = errors.New("public key is empty")
	ErrInvalidPublicKey         = errors.New("public key is not base64 encoded or has invalid size")
	ErrUserNotFoundForPublicKey = errors.New("user with specified public key not found")
	ErrUnauthorizedAccess       = errors.New("unauthorized access to pastes")
)

var (
	ErrInvalidSignature = errors.New("Invalid signature")
	ErrEmptySignature   = errors.New("Empty signature")
)

var (
	ErrInvalidInput   = errors.New("invalid input provided")
	ErrDatabase       = errors.New("database operation failed")
	ErrInternalServer = errors.New("internal server error")
	ErrNotFound       = errors.New("resource not found")
)

var (
	ErrDuplicatePublicKey   = errors.New("public key already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidUserID        = errors.New("invalid User ID")
	ErrEmptyUserID          = errors.New("empty user ID")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrValidationError      = errors.New("validation error")
	ErrUserCreationFailed   = errors.New("user creation failed")
	ErrUserAlreadyExists    = errors.New("user already exists, perform login instead")
)

func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
