package utils

import "errors"

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
)

var (
	ErrEmptyPublicKey           = errors.New("public key is empty")
	ErrInvalidPublicKey         = errors.New("public key is not base64 encoded or has invalid size")
	ErrUserNotFoundForPublicKey = errors.New("user with specified public key not found")
	ErrUnauthorizedAccess       = errors.New("unauthorized access to pastes")
)
