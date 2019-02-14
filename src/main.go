package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/Zedjones/Reddit-Refresh-Online/RROnline"
)

const pushURLTemplate = "https://www.pushbullet.com/authorize?client_id=%s" +
	"&redirect_uri=%s&response_type=code&scope=everything"

var routineManager RROnline.RoutineManager
var pushURL string

type isValid struct {
	IsValid bool `json:"valid"`
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
	RROnline.LoadDBConfig()
	RROnline.LoadPushConfig()
	getPushURL()
	go startSearches()
	e := echo.New()
	e.Use(middleware.CSRF())
	e.Use(middleware.Logger())
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("../templates/*.html")),
	}
	e.Renderer = renderer
	e.Static("/", "..")
	e.POST("/validateSubreddit", validateRoute)
	e.GET("/handle_token", handleToken)
	e.GET("/", index)
	e.GET("/gettingStarted", gettingStarted)
	e.GET("/searchPage", mainPage)
	e.POST("/addSearch", addSearch)
	e.POST("/deleteSub", deleteSub)
	e.Start(":1234")
}

func getPushURL() {
	redirectURI := strings.Replace(RROnline.PushConf.RedirectURI, ":", "%3A", -1)
	redirectURI = strings.Replace(redirectURI, "/", "%2F", -1)
	pushURL = fmt.Sprintf(pushURLTemplate, RROnline.PushConf.ClientID, redirectURI)
}

func startSearches() {
	routineManager = *RROnline.CreateManager()
	searches := RROnline.GetAllSearches()
	for _, item := range searches {
		token := RROnline.GetUserToken(item.Email)
		routineManager.RMAddSearch(token, item.Sub, item.Search)
	}
}

func validateRoute(c echo.Context) error {
	sub := new(Sub)
	if err := c.Bind(sub); err != nil {
		fmt.Fprintln(os.Stderr, "Error binding JSON body to sub.")
	}
	if sub.Sub == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	jsonBody := &isValid{RROnline.ValidateSub(sub.Sub)}
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
		RROnline.AddUser(email, userTok, db)
		RROnline.RefreshDevices(userTok, db, nil)
	} else {
		RROnline.UpdateUserToken(email, userTok)
	}
	return c.Redirect(http.StatusFound, "/searchPage")
}

func index(c echo.Context) error {
	userToken, err := c.Cookie("user_token")
	data := make(map[string]interface{})
	if err != nil {
		data["login"] = "Login"
		data["url"] = pushURL
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
	refreshChan := make(chan bool)
	db := RROnline.Connect()
	go RROnline.RefreshDevices(userToken.Value, db, refreshChan)
	name := RROnline.GetUserName(userToken.Value)
	email := RROnline.GetEmail(userToken.Value)
	searches := RROnline.GetSearches(email, db)
	searchMap := make(map[string][]string)
	for _, search := range searches {
		if _, ok := searchMap[search.Sub]; ok {
			searchMap[search.Sub] = append(searchMap[search.Sub], search.Search)
		} else {
			searchMap[search.Sub] = []string{search.Search}
		}
	}
	<-refreshChan
	devices := RROnline.GetDevices(email, db)
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
		fmt.Fprintln(os.Stderr, "Error binding JSON body to searches.")
	}
	userToken, err := c.Cookie("user_token")
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	email := RROnline.GetEmail(userToken.Value)
	for _, search := range searches.Searches {
		RROnline.AddSearch(userToken.Value, email, searches.Sub,
			search, routineManager)
	}
	RROnline.DeleteMissingSearches(email, searches.Sub,
		searches.Searches, routineManager)
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
	err = RROnline.DeleteSub(email, sub.Sub, routineManager)
	if err != nil {
		//TODO: handle error here
	}
	return nil
}
