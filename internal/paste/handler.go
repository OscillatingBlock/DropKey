package paste

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	custom_middleware "Drop-Key/internal/middleware"
	"Drop-Key/internal/models"
	"Drop-Key/internal/utils"

	"github.com/labstack/echo/v4"
)

type PasterHandlerInterface interface {
	CreatePaste(c echo.Context) error
	GetPaste(c echo.Context) error
	UpdatePaste(c echo.Context) error
	GetByPublicKey(c echo.Context) error
}

type pasteHandler struct {
	service PasteService
}

func NewPasteHandler(pasteService PasteService) *pasteHandler {
	return &pasteHandler{
		service: pasteService,
	}
}

type PasteRequest struct {
	Ciphertext string `json:"ciphertext"`
	Signature  string `json:"signature"`
	PublicKey  string `json:"public_key"`
	Expires_in int    `json:"expires_in"`
}

type PasteResponse struct {
	ID         string    `json:"ID"`
	Ciphertext string    `json:"ciphertext"`
	Signature  string    `json:"signature"`
	PublicKey  string    `json:"public_key"`
	Expires_in time.Time `json:"expires_in"`
}

type Url struct {
	URL string `json:"url"`
}

type PublicKey struct {
	PublicKey string `json:"public_key"`
}

func (h *pasteHandler) CreatePaste(c echo.Context) error {
	pasteReq := &PasteRequest{}
	if err := c.Bind(pasteReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}

	userInfo, ok := c.Get("userInfo").(custom_middleware.UserInfo)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "User info not found in context")
	}
	if userInfo.Publickey != pasteReq.PublicKey {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized access")
	}

	paste := &models.Paste{
		Signature:  pasteReq.Signature,
		Ciphertext: pasteReq.Ciphertext,
		PublicKey:  pasteReq.PublicKey,
	}

	ctx := c.Request().Context()
	id, err := h.service.Create(ctx, paste, pasteReq.Expires_in)

	switch {
	case err == nil:

	case errors.Is(err, utils.ErrPasteExpiredAlready):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid expiresin, paste expired already")
	case errors.Is(err, utils.ErrPasteExpiryTooLong):
		return echo.NewHTTPError(http.StatusBadRequest, "Expiry date too long")
	case errors.Is(err, utils.ErrPasteEmptyCiphertext):
		return echo.NewHTTPError(http.StatusBadRequest, "Empty ciphertext")
	case errors.Is(err, utils.ErrPasteInvalidCiphertext):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ciphertext")
	case errors.Is(err, utils.ErrEmptySignature):
		return echo.NewHTTPError(http.StatusBadRequest, "Empty signature")
	case errors.Is(err, utils.ErrInvalidSignature):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid signature")
	case errors.Is(err, utils.ErrEmptyPublicKey):
		return echo.NewHTTPError(http.StatusBadRequest, "Empty public key")
	case errors.Is(err, utils.ErrInvalidPublicKey):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid public key")
	case errors.Is(err, utils.ErrPasteUserNotFound):
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized access")
	case errors.Is(err, utils.ErrPasteInvalidSignatureVerification):
		return echo.NewHTTPError(http.StatusBadRequest, "Signature verification failed")

	default:
		slog.Error("error while creating paste", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	slog.Info("paste created with ", "pasteid", paste.ID, "url", getUrl(id, paste.PublicKey))
	return c.JSON(http.StatusCreated, map[string]string{
		"id":  paste.ID,
		"url": getUrl(id, paste.PublicKey),
	})
}

func getUrl(id, pub string) string {
	baseUrl := os.Getenv("BASEURL")
	if baseUrl == "" {
		baseUrl = "https://yourpastebin.com"
	}
	url := baseUrl + "/paste/" + id + "#" + pub
	return url
}

func (h *pasteHandler) GetPaste(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing paste ID")
	}

	ctx := c.Request().Context()
	paste, err := h.service.GetByID(ctx, id)

	switch {
	case errors.Is(err, utils.ErrPasteInvalidID):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid paste ID")
	case errors.Is(err, utils.ErrPasteNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "Paste not found")
	case errors.Is(err, utils.ErrPasteExpiredAlready):
		return echo.NewHTTPError(http.StatusGone, "Paste already expired")
	case err != nil:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSONPretty(http.StatusOK, paste, " ")
}

func (h *pasteHandler) UpdatePaste(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing paste ID")
	}

	pasteReq := &PasteRequest{}
	if err := c.Bind(pasteReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	userInfo, ok := c.Get("userInfo").(custom_middleware.UserInfo)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "User info not found in context")
	}
	if userInfo.Publickey != pasteReq.PublicKey {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized access")
	}

	paste := &models.Paste{
		ID:         id,
		Signature:  pasteReq.Signature,
		Ciphertext: pasteReq.Ciphertext,
		PublicKey:  pasteReq.PublicKey,
	}

	ctx := c.Request().Context()
	err := h.service.Update(ctx, paste, pasteReq.Expires_in)

	switch {
	case errors.Is(err, utils.ErrPasteNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "Paste not found")

	case errors.Is(err, utils.ErrPasteInvalidExpiryTime):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid expires_in")

	case errors.Is(err, utils.ErrUnauthorizedAccess):
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized access")

	case errors.Is(err, utils.ErrPasteInvalidCiphertext):
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request, invalid cipher text")

	case err != nil:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error, failed to update paste")
	}

	return c.String(http.StatusOK, "paste updated")
}

func (h *pasteHandler) GetByPublicKey(c echo.Context) error {
	pubB64 := c.QueryParam("public_key")

	pastes, err := h.service.GetByPublicKey(c.Request().Context(), pubB64)

	switch {
	case err == nil:
		if len(pastes) == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "All pastes have expired for this user")
		}
		return c.JSONPretty(http.StatusOK, pastes, "\t")

	case errors.Is(err, utils.ErrEmptyUserID):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request, empty public key")

	case errors.Is(err, utils.ErrInvalidPublicKey):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid public key")

	case errors.Is(err, utils.ErrUnauthorizedAccess):
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized access")

	case errors.Is(err, utils.ErrPasteNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "Paste not found")

	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}
}
