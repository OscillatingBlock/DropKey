package paste

import (
	"net/http"
	"os"

	"Drop-Key/internal/models"

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
	Ciphertext string `json:"ciphertext"`
	Signature  string `json:"signature"`
	PublicKey  string `json:"publickey"`
}

func (h *PasteHandler) CreatePaste(c echo.Context) error {
	pasteReq := &PasteRequest{}
	err := c.Bind(pasteReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Internal server error")
	}
	paste := &models.Paste{
		Signature:  pasteReq.Signature,
		Ciphertext: pasteReq.Ciphertext,
		PublicKey:  pasteReq.PublicKey,
	}

	ctx := c.Request().Context()
	id, err := h.service.Create(ctx, paste, pasteReq.Expires_in)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	url := getUrl(id, paste.PublicKey)
	return c.JSON(http.StatusOK, url)
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
	var id string
	id = c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing paste ID")
	}

	ctx := c.Request().Context()
	paste, err := h.service.GetByID(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response := PasteResponse{
		Ciphertext: paste.Ciphertext,
		Signature:  paste.Signature,
		PublicKey:  paste.PublicKey,
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
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error, faield to update paste.")
	}

	return c.String(http.StatusOK, "paste updated")
}
