package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"

	"../reddit_refresh_online"
)

type isValid struct {
	IsValid bool `json:"is_valid"`
}

func main() {
	e := echo.New()
	e.POST("/validate_subreddit", validateRoute)
	e.Start(":1234")
}

func validateRoute(c echo.Context) error {
	sub := c.QueryParam("sub")
	fmt.Println(sub)
	if sub == "" {
		return c.NoContent(http.StatusNotFound)
	}
	jsonBody := &isValid{reddit_refresh_online.ValidateSub(sub)}
	return c.JSON(http.StatusOK, jsonBody)
}
