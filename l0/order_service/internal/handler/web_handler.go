package handler

import (
	"embed"
	"net/http"
	"path"
	"text/template"

	"github.com/gin-gonic/gin"
)

//go:embed web/templates/* web/static/*
var webFS embed.FS

type WebHandler struct {
	templates *template.Template
}

func NewWebHandler() (*WebHandler, error) {
	tmpl, err := template.ParseFS(webFS, "web/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &WebHandler{
		templates: tmpl,
	}, nil
}

// index
func (h *WebHandler) Index(c *gin.Context) {
	h.templates.ExecuteTemplate(c.Writer, "index.html", nil)
}

// static
func (h *WebHandler) Static(c *gin.Context) {
	file := c.Param("file")
	c.FileFromFS(path.Join("web/static", file), http.FS(webFS))
}
