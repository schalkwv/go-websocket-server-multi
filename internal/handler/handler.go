package handler

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	appVersion string
}

// Validator is a custom validator for Echo.
type Validator struct {
	validator *validator.Validate
}

// Validate validates the request according to the required tags.
// Returns HTTPError if the required parameter is missing in the request.
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func NewServer(appVersion string) *Server {
	return &Server{
		appVersion: appVersion,
	}
}

// Router returns the echo router.
func (s *Server) Router() *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORS())

	e.GET("/version", s.versionHandler)

	e.Validator = &Validator{validator: validator.New()}

	return e
}

func (s *Server) versionHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.appVersion)
}
