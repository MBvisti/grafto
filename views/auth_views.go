package views

import (
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
)

func (v Views) LoginPage(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "user/login", RenderOpts{
		Layout: BaseLayout,
		Data: LoginForm{
			CsrfField: template.HTML(csrf.TemplateField(ctx.Request())),
		},
	})
}

type LoginForm struct {
	EmailNeedsVerification bool
	CouldNotAuthenticate   bool
	CsrfField              template.HTML
}

func (v Views) LoginForm(ctx echo.Context, data LoginForm) error {
	return ctx.Render(http.StatusOK, "user/__login_form", RenderOpts{
		Data: data,
	})
}

func (v Views) Authenticated(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "user/__authenticated", RenderOpts{
		Data: nil,
	})
}

type EmailValidationData struct {
	TokenInvalid bool
}

func (v Views) EmailValidation(ctx echo.Context, data EmailValidationData) error {
	return ctx.Render(http.StatusOK, "user/email_validation", RenderOpts{
		Layout: BaseLayout,
		Data:   data,
	})
}