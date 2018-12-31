package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo"

	"../RROnline"
)

type isValid struct {
	IsValid bool `json:"is_valid"`
}

type userToken struct {
	UserToken string `json:"user_token"`
}

func main() {
	e := echo.New()
	e.POST("/validate_subreddit", validateRoute)
	e.POST("/handle_token", handleToken)
	e.Start(":1234")
}

func validateRoute(c echo.Context) error {
	sub := c.QueryParam("sub")
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
	cookie := new(http.Cookie)
	cookie.Name = "user_token"
	cookie.Value = userTok
	cookie.Expires = time.Now().Add(24 * time.Hour)
	c.SetCookie(cookie)
	jsonBody := &userToken{userTok}
	return c.JSON(http.StatusOK, jsonBody)
}
