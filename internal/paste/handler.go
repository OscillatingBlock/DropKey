package paste

import (
	"net/http"
	"os"

	"Drop-Key/internal/models"
	"Drop-Key/internal/utils"

	"github.com/labstack/echo/v4"
)

type PasteHandler struct {
	service PasteService
}

type PasteRequest struct {
	Ciphertext string `json:"ciphertext"`
	Signature  string `json:"signature"`
	PublicKey  string `json:"publickey"`
	Expires_in int    `json:"expiresin"`
}

type PasteResponse struct {
	ID         string `json:"ID"`
	Ciphertext string `json:"ciphertext"`
	Signature  string `json:"signature"`
	PublicKey  string `json:"publickey"`
}

type Url struct {
	URL string `json:"url"`
}

type PublicKey struct {
	PublicKey string `json:"publickey"`
}

func (h *PasteHandler) CreatePaste(c echo.Context) error {
	pasteReq := &PasteRequest{}
	if err := c.Bind(pasteReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}

	paste := &models.Paste{
		Signature:  pasteReq.Signature,
		Ciphertext: pasteReq.Ciphertext,
		PublicKey:  pasteReq.PublicKey,
	}

	ctx := c.Request().Context()
	id, err := h.service.Create(ctx, paste, pasteReq.Expires_in)

	switch err {
	case nil:

	case utils.ErrPasteExpiredAlready:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid expires_in")
	case utils.ErrPasteExpiryTooLong:
		return echo.NewHTTPError(http.StatusBadRequest, "Expiry date too long")
	case utils.ErrPasteEmptyCiphertext:
		return echo.NewHTTPError(http.StatusBadRequest, "Empty ciphertext")
	case utils.ErrPasteInvalidCiphertext:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ciphertext")
	case utils.ErrEmptySignature:
		return echo.NewHTTPError(http.StatusBadRequest, "Empty signature")
	case utils.ErrInvalidSignature:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid signature")
	case utils.ErrEmptyPublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "Empty public key")
	case utils.ErrInvalidPublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid public key")
	case utils.ErrPasteUserNotFound:
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized access")
	case utils.ErrPasteInvalidSignatureVerification:
		return echo.NewHTTPError(http.StatusBadRequest, "Signature verification failed")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	fetchedUrl := getUrl(id, paste.PublicKey)
	url := &Url{URL: fetchedUrl}
	return c.JSON(http.StatusCreated, url)
}

func getUrl(id, pub string) string {
	baseUrl := os.Getenv("BASEURL")
	if baseUrl == "" {
		baseUrl = "https://yourpastebin.com"
	}
	url := baseUrl + "/paste/" + id + "#" + pub
	return url
}

func (h *PasteHandler) GetPaste(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing paste ID")
	}

	ctx := c.Request().Context()
	paste, err := h.service.GetByID(ctx, id)

	switch {
	case err == utils.ErrPasteInvalidID:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid paste ID")
	case err == utils.ErrPasteNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "Paste not found")
	case err == utils.ErrPasteExpiredAlready:
		return echo.NewHTTPError(http.StatusGone, "Paste already expired")
	case err != nil:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response := PasteResponse{
		Ciphertext: paste.Ciphertext,
		Signature:  paste.Signature,
		PublicKey:  paste.PublicKey,
		ID:         paste.ID,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *PasteHandler) UpdatePaste(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing paste ID")
	}

	pasteReq := &PasteRequest{}
	if err := c.Bind(pasteReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	paste := &models.Paste{
		ID:         id,
		Signature:  pasteReq.Signature,
		Ciphertext: pasteReq.Ciphertext,
		PublicKey:  pasteReq.PublicKey,
	}

	ctx := c.Request().Context()
	err := h.service.Update(ctx, paste)

	switch {
	case err == utils.ErrPasteNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "Paste not found")
	case err == utils.ErrUnauthorizedAccess:
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized access")
	case err == utils.ErrPasteInvalidCiphertext:
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request, invalid cipher text")
	case err != nil:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error, failed to update paste")
	}

	return c.String(http.StatusOK, "paste updated")
}

func (h *PasteHandler) GetByPublicKey(c echo.Context) error {
	pubB64 := &PublicKey{}
	err := c.Bind(pubB64)

	switch {
	case err == utils.ErrEmptyUserID:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request, empty public key")
	case err == utils.ErrInvalidPublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid public key")
	case err == utils.ErrUnauthorizedAccess:
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized access")
	case err == utils.ErrPasteNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "Paste not found")
	case err != nil:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	pastes, err := h.service.GetByPublicKey(c.Request().Context(), pubB64.PublicKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error, failed to get pastes by public key")
	}

	return c.JSON(http.StatusOK, pastes)
}
