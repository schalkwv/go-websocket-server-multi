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
	e.Validator = &Validator{validator: validator.New()}

	e.GET("/version", s.versionHandler)
	e.GET("/", s.rootHandler)

	return e
}

func (s *Server) versionHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.appVersion)
}

type account struct {
	Name string `json:"name" validate:"required"`
	Port int    `json:"port" validate:"required"`
}

type accountList map[string]account

func (s *Server) rootHandler(c echo.Context) error {
	tmpl := getTemplate()

	accounts := accountList{
		"account1": {
			Name: "Account 1",
			Port: 5001,
		},
		"account2": {
			Name: "Account 2",
			Port: 5002,
		},
		"account3": {
			Name: "Account 3",
			Port: 5003,
		},
	}

	err := tmpl.ExecuteTemplate(c.Response().Writer, "Base",
		map[string]any{
			"Accounts": accounts,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return nil
}
