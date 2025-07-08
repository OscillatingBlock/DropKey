package router

import (
	"Drop-Key/internal/middleware"
	"Drop-Key/internal/paste"
	"Drop-Key/internal/user"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Router(pasteHandler paste.PasterHandlerInterface, userHandler user.UserHandler) *echo.Echo {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())

	publicPasteGroup := e.Group("/api/pastes", custom_middleware.Logger)
	publicPasteGroup.GET("/:id", pasteHandler.GetPaste)
	publicPasteGroup.GET("", pasteHandler.GetByPublicKey)

	protectedPasteGroup := e.Group("/api/pastes", custom_middleware.Logger, custom_middleware.JwtAuth)
	protectedPasteGroup.POST("", pasteHandler.CreatePaste)
	protectedPasteGroup.PUT("/:id", pasteHandler.UpdatePaste)

	userGroup := e.Group("/api/users", custom_middleware.Logger)
	userGroup.POST("", userHandler.RegisterHandler)
	userGroup.POST("/auth", userHandler.AuthenticateHandler)
	userGroup.GET("/:id", userHandler.GetByIDHandler)
	userGroup.GET("", userHandler.GetByPublicKeyHandler)
	return e
}
