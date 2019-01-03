package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
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

const PUSH_URL = "https://www.pushbullet.com/authorize?client_id=" +
	"PR0sGjjxNmfu8OwRrawv2oxgZllvsDm1&redirect_uri=http%3A%2F%2F" +
	"localhost%3A1234%2Fhandle_token&response_type=code&scope=everything"

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
	e.GET("/searchPage", mainPage)
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
		if err.Error() != "http: name cookie not present" {
			fmt.Fprintf(os.Stderr, "Error getting cookie user_token.")
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
	return c.Redirect(http.StatusFound, "/searchPage")
}

func index(c echo.Context) error {
	userToken, err := c.Cookie("user_token")
	data := make(map[string]interface{})
	if err != nil {
		data["login"] = "Login"
		data["url"] = PUSH_URL
		return c.Render(http.StatusOK, "index.html", data)
	}
	name := RROnline.GetUserName(userToken.Value)
	data["login"] = name
	data["url"] = "/searchPage"
	return c.Render(http.StatusOK, "index.html", data)
}

func gettingStarted(c echo.Context) error {
	return c.Render(http.StatusOK, "gettingStarted.html", nil)
}

func mainPage(c echo.Context) error {
	userToken, err := c.Cookie("user_token")
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	name := RROnline.GetUserName(userToken.Value)
	data := make(map[string]interface{})
	data["name"] = name
	return c.Render(http.StatusOK, "searchPage.html", data)
}
