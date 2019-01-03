package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"../RROnline"
)

type isValid struct {
	IsValid bool `json:"is_valid"`
}

type userToken struct {
	UserToken string `json:"user_token"`
}

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Use(middleware.CSRF())
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("../templates/*.html")),
	}
	e.Renderer = renderer
	e.Static("/", "..")
	e.POST("/validate_subreddit", validateRoute)
	e.GET("/handle_token", handleToken)
	e.GET("/", index)
	e.GET("/gettingStarted", gettingStarted)
	e.Start(":1234")
}

func validateRoute(c echo.Context) error {
	sub := c.QueryParam("sub")
	fmt.Printf(sub)
	if sub == "" {
		return c.NoContent(http.StatusNotFound)
	}
	jsonBody := &isValid{RROnline.ValidateSub(sub)}
	return c.JSON(http.StatusOK, jsonBody)
}

func handleToken(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.NoContent(http.StatusNotFound)
	}
	userTok := RROnline.GetToken(code)
	_, err := c.Cookie("user_token")
	if err != nil {
		if err.Error() == "http: name cookie not present" {
			return c.Render(http.StatusOK, "mainPage.html", nil)
		}
	}
	cookie := new(http.Cookie)
	cookie.Name = "user_token"
	cookie.Value = userTok
	//cookie.Secure = true
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HttpOnly = true
	c.SetCookie(cookie)
	email := RROnline.GetEmail(userTok)
	if !RROnline.UserExists(email) {
		RROnline.AddUser(email, RROnline.DEFAULT_INTERVAL)
	}
	return c.Render(http.StatusOK, "mainPage.html", nil)
}

func index(c echo.Context) error {
	userToken, err := c.Cookie("user_token")
	fmt.Println("Rendering index")
	data := make(map[string]interface{})
	if err != nil {
		data["login"] = "Login"
		return c.Render(http.StatusOK, "index.html", data)
	}
	name := RROnline.GetUserName(userToken.Value)
	data["login"] = name
	return c.Render(http.StatusOK, "index.html", data)
}

func gettingStarted(c echo.Context) error {
	return c.Render(http.StatusOK, "gettingStarted.html", nil)
}
