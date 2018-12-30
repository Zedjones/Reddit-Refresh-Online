package main

import (
	"net/http"

	"github.com/labstack/echo"

	"../reddit_refresh_online"
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
	jsonBody := &isValid{reddit_refresh_online.ValidateSub(sub)}
	return c.JSON(http.StatusOK, jsonBody)
}

func handleToken(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.NoContent(http.StatusNotFound)
	}
	userTok := reddit_refresh_online.GetToken(code)
	jsonBody := &userToken{userTok}
	return c.JSON(http.StatusOK, jsonBody)
}
