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

const PUSH_URL = "https://www.pushbullet.com/authorize?client_id=" +
	"PR0sGjjxNmfu8OwRrawv2oxgZllvsDm1&redirect_uri=http%3A%2F%2F" +
	"localhost%3A1234%2Fhandle_token&response_type=code&scope=everything"

type isValid struct {
	IsValid bool `json:"is_valid"`
}

type userToken struct {
	UserToken string `json:"user_token"`
}

type TemplateRenderer struct {
	templates *template.Template
}

type Searches struct {
	Sub      string   `json:"subreddit"`
	Searches []string `json:"searches"`
}

type Search struct {
	Sub    string `json:"subreddit"`
	Search string `json:"search"`
}

type Sub struct {
	Sub string `json:"subreddit"`
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Use(middleware.CSRF())
	e.Use(middleware.Logger())
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
	e.POST("/addSearch", addSearch)
	e.POST("/deleteSub", deleteSub)
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
		if err.Error() != "http: named cookie not present" {
			fmt.Fprintf(os.Stderr, "Error getting cookie user_token.\n")
			fmt.Println(err)
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
		db := RROnline.Connect()
		RROnline.AddUser(email, RROnline.DEFAULT_INTERVAL, db)
		RROnline.RefreshDevices(userTok, db)
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
	email := RROnline.GetEmail(userToken.Value)
	db := RROnline.Connect()
	searches := RROnline.GetSearches(email, db)
	searchMap := make(map[string][]string)
	for _, search := range searches {
		if _, ok := searchMap[search.Sub]; ok {
			searchMap[search.Sub] = append(searchMap[search.Sub], search.Search)
		} else {
			searchMap[search.Sub] = []string{search.Search}
		}
	}
	devices := RROnline.GetDevices(email, db)
	fmt.Println(devices)
	data := make(map[string]interface{})
	data["name"] = name
	data["searches"] = searchMap
	data["devices"] = devices
	data["inc"] = func(i int) int {
		return i + 1
	}
	return c.Render(http.StatusOK, "searchPage.html", data)
}

func addSearch(c echo.Context) error {
	searches := new(Searches)
	if err := c.Bind(searches); err != nil {
		fmt.Fprintln(os.Stderr, "Error binding JSON body to searches.\n")
	}
	userToken, err := c.Cookie("user_token")
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	email := RROnline.GetEmail(userToken.Value)
	for _, search := range searches.Searches {
		RROnline.AddSearch(email, searches.Sub, search)
	}
	RROnline.DeleteMissingSearches(email, searches.Sub, searches.Searches)
	return nil
}

func deleteSub(c echo.Context) error {
	sub := new(Sub)
	if err := c.Bind(sub); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding JSON body to search.\n")
	}
	userToken, err := c.Cookie("user_token")
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	email := RROnline.GetEmail(userToken.Value)
	err = RROnline.DeleteSub(email, sub.Sub)
	if err != nil {
		//TODO: handle error here
	}
	return nil
}
