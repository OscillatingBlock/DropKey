package user

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"Drop-Key/internal/middleware"
	"Drop-Key/internal/models"
	"Drop-Key/internal/utils"

	"github.com/golang-jwt/jwt/v5"
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
	PublicKey string `json:"public_key"`
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
	switch {

	case err == nil:
		response := &ID{
			Id: id,
		}
		return c.JSON(http.StatusCreated, response)

	case errors.Is(err, utils.ErrEmptyPublicKey):
		return echo.NewHTTPError(http.StatusBadRequest, "Empty public key")

	case errors.Is(err, utils.ErrInvalidPublicKey):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid public key")

	case errors.Is(err, utils.ErrUserCreationFailed):
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error.")

	case errors.Is(err, utils.ErrDuplicatePublicKey):
		return echo.NewHTTPError(http.StatusBadRequest, "User already exists")

	default:
		pubKeyB64, _ := base64.StdEncoding.DecodeString(user.PublicKey)
		slog.Error("Error while registering user with publicKey", "publicKey", string(pubKeyB64), "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error..")
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
		jwtSecret := os.Getenv("JWTSECRET")
		if jwtSecret == "" {
			slog.Error("No JWTSECRET found", "JWTSECRET", jwtSecret)
		}

		user, err := h.service.GetByID(c.Request().Context(), req.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
		}

		fmt.Println("creating claims	")
		fmt.Println("pub key", user.PublicKey)
		fmt.Println("userid", user.ID)
		claims := &custom_middleware.JwtCustomClaims{
			UserInfo: custom_middleware.UserInfo{
				UserID:    user.ID,
				Publickey: user.PublicKey,
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		t, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to generate token")
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Authentication successful", "token": t})

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
	publicKey := c.QueryParam("public_key")
	if publicKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing public key")
	}

	user, err := h.service.GetByPublicKey(c.Request().Context(), publicKey)

	switch {
	case err == nil:
		return c.JSON(http.StatusOK, user)

	case errors.Is(err, utils.ErrEmptyPublicKey):
		return echo.NewHTTPError(http.StatusBadRequest, "Empty publickey")

	case errors.Is(err, utils.ErrInvalidPublicKey):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid public key — ensure it is URL-encoded")

	case errors.Is(err, utils.ErrUserNotFound):
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
	user, err := h.service.GetByID(c.Request().Context(), id)

	switch {
	case err == nil:
		return c.JSON(http.StatusOK, user)
	case errors.Is(err, utils.ErrEmptyUserID):
		return echo.NewHTTPError(http.StatusBadRequest, "Empty id")
	case errors.Is(err, utils.ErrInvalidUserID):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid id")
	case errors.Is(err, utils.ErrUserNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "User not found")

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Internal server error")
	}
}
