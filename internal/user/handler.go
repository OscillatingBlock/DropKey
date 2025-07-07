package user

import (
	"errors"
	"net/http"

	"Drop-Key/internal/models"
	"Drop-Key/internal/utils"

	"github.com/labstack/echo/v4"
)

type UserHandler interface {
	RegisterHandler(c echo.Context) error
	AuthenticateHandler(c echo.Context) error
	GetByIDHandler(c echo.Context) error
	GetByPublicKeyHandler(c echo.Context) error
}

type userHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *userHandler {
	return &userHandler{
		service: service,
	}
}

type pub struct {
	PublicKey string `json:"publickey"`
}

type ID struct {
	Id string `json:"id"`
}

type AuthRequest struct {
	ID        string `json:"id"`
	Signature string `json:"signature"`
	Challenge string `json:"challenge"`
}

func (h *userHandler) RegisterHandler(c echo.Context) error {
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
			Id: id,
		}
		return c.JSON(http.StatusCreated, response)

	case utils.ErrEmptyPublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "Empty public key")

	case utils.ErrInvalidPublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid public key")

	case utils.ErrUserCreationFailed:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")

	case utils.ErrDuplicatePublicKey:
		return echo.NewHTTPError(http.StatusBadRequest, "User already exists")

	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}
}

func (h *userHandler) AuthenticateHandler(c echo.Context) error {
	req := &AuthRequest{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}

	ok, err := h.service.Authenticate(c.Request().Context(), req.ID, req.Signature, req.Challenge)

	switch {
	case err == nil && ok:
		return c.JSON(http.StatusOK, map[string]string{"message": "Authentication successful"})

	case errors.Is(err, utils.ErrEmptyUserID):
		return echo.NewHTTPError(http.StatusBadRequest, "Missing user ID")

	case errors.Is(err, utils.ErrValidationError):
		return echo.NewHTTPError(http.StatusBadRequest, "Missing or invalid challenge")

	case errors.Is(err, utils.ErrEmptySignature):
		return echo.NewHTTPError(http.StatusBadRequest, "Missing signature")

	case errors.Is(err, utils.ErrInvalidSignature):
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid signature")

	case errors.Is(err, utils.ErrUserNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "User not found")

	case errors.Is(err, utils.ErrInvalidPublicKey):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid public key")

	case err != nil:
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return echo.NewHTTPError(http.StatusUnauthorized, "Authentication failed")
}

func (h *userHandler) GetByPublicKeyHandler(c echo.Context) error {
	publicKey := c.QueryParam("publickey")
	if publicKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing public key")
	}
	err := c.Bind(publicKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}

	user, err := h.service.GetByPublicKey(c.Request().Context(), publicKey)

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

func (h *userHandler) GetByIDHandler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing user ID")
	}
	err := c.Bind(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}
	user, err := h.service.GetByID(c.Request().Context(), id)

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
