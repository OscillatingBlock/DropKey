package user

import (
	"net/http"

	"Drop-Key/internal/models"
	"Drop-Key/internal/utils"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service UserService
}

type pub struct {
	PublicKey string `json:"publickey"`
}

type ID struct {
	ID string `json:"id"`
}

type AuthRequest struct {
	ID        string `json:"ID"`
	Signature string `json:"signature"`
	Challenge string `json:"challenge"`
}

func (h *UserHandler) RegisterHandler(c echo.Context) error {
	pub := &pub{}
	err := c.Bind(pub)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}

	user := &models.User{
		PublicKey: pub.PublicKey,
	}
	id, err := h.service.Create(c.Request().Context(), user)
	switch err {

	case nil:
		response := &ID{
			ID: id,
		}
		return c.JSON(http.StatusCreated, response)

	case utils.ErrEmptyPublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "Empty public key")

	case utils.ErrInvalidPublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid public key")

	case utils.ErrUserCreationFailed:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")

	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}
}

func (h *UserHandler) AuthenticateHandler(c echo.Context) error {
	req := &AuthRequest{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}

	ok, err := h.service.Authenticate(c.Request().Context(), req.ID, req.Signature, req.Challenge)

	switch {
	case err == nil && ok:
		return c.JSON(http.StatusOK, map[string]string{"message": "Authentication successful"})

	case err == utils.ErrEmptyUserID:
		return echo.NewHTTPError(http.StatusBadRequest, "Missing user ID")

	case err == utils.ErrValidationError:
		return echo.NewHTTPError(http.StatusBadRequest, "Missing or invalid challenge")

	case err == utils.ErrEmptySignature:
		return echo.NewHTTPError(http.StatusBadRequest, "Missing signature")

	case err == utils.ErrInvalidSignature:
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid signature")

	case err == utils.ErrUserNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "User not found")

	case err == utils.ErrInvalidPublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid public key")

	case err != nil:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return echo.NewHTTPError(http.StatusUnauthorized, "Authentication failed")
}

func (h *UserHandler) GetByPublicKeyHandler(c echo.Context) error {
	publicKey := &pub{}
	err := c.Bind(publicKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}

	user, err := h.service.GetByPublicKey(c.Request().Context(), publicKey.PublicKey)

	switch err {
	case nil:
		return c.JSON(http.StatusOK, user)
	case utils.ErrEmptyPublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "Empty publickey")

	case utils.ErrInvalidPublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid publickey")

	case utils.ErrNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "User not found")

	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}
}

func (h *UserHandler) GetByIDHandler(c echo.Context) error {
	id := &ID{}
	err := c.Bind(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}
	user, err := h.service.GetByID(c.Request().Context(), id.ID)

	switch err {
	case nil:
		return c.JSON(http.StatusOK, user)
	case utils.ErrEmptyUserID:
		return echo.NewHTTPError(http.StatusBadRequest, "Empty id")
	case utils.ErrInvalidUserID:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid id")
	case utils.ErrUserNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "User not found")

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Internal server error")
	}
}
