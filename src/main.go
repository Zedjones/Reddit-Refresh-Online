package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"

	"github.com/Zedjones/Reddit-Refresh-Online/RROnline"
	"github.com/Zedjones/Reddit-Refresh-Online/logger"
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

/*
TemplateRenderer is exported for the Echo template renderer to use
*/
type TemplateRenderer struct {
	templates *template.Template
}

type searches struct {
	Sub      string   `json:"subreddit"`
	Searches []string `json:"searches"`
}

type search struct {
	Sub    string `json:"subreddit"`
	Search string `json:"search"`
}

type device struct {
	ID     string `json:"id"`
	Active bool   `json:"active"`
}

type sub struct {
	Sub string `json:"subreddit"`
}

type interval struct {
	Interval float32 `json:"interval"`
}

/*
Render is exported for the Echo template renderer to use
*/
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	logger.Init()
	logger.Log.Println("Starting server")
	RROnline.LoadDBConfig()
	RROnline.LoadPushConfig()
	getPushURL()
	go startSearches()
	e := echo.New()
	e.Use(middleware.CSRF())
	e.Use(middleware.Logger())
	sessionStr := make([]byte, 32)
	rand.Read(sessionStr)
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(sessionStr))))
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("../templates/*.html")),
	}
	e.Renderer = renderer
	e.Static("/", "..")
	e.GET("/handle_token", handleToken)
	e.GET("/", index)
	e.GET("/searchPage", mainPage)
	e.POST("/addSearch", addSearch)
	e.POST("/deleteSub", deleteSub)
	e.POST("/updateInterval", updateInterval)
	e.POST("/validateSubreddit", validateRoute)
	e.POST("/editDevice", editDevice)
	e.Start(":1234")
}

func getPushURL() {
	redirectURI := strings.Replace(RROnline.PushConf.RedirectURI, ":", "%3A", -1)
	redirectURI = strings.Replace(redirectURI, "/", "%2F", -1)
	pushURL = fmt.Sprintf(pushURLTemplate, RROnline.PushConf.ClientID, redirectURI)
}

func startSearches() {
	logger.Log.Println("Starting searches")
	routineManager = *RROnline.CreateManager()
	searches, err := RROnline.GetAllSearches()
	if err != nil {
		RROnline.LogDBError(err)
	}
	for _, item := range searches {
		token, err := RROnline.GetUserToken(item.Email)
		if err != nil {
			RROnline.LogDBError(err)
		}
		routineManager.RMAddSearch(token, item.Sub, item.Search)
	}
}

func updateInterval(c echo.Context) error {
	interval := new(interval)
	if err := c.Bind(interval); err != nil {
		logger.Log.Println("Error binding JSON body to interval.")
		return c.NoContent(http.StatusBadRequest)
	}
	userToken, err := c.Cookie("user_token")
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	email := RROnline.GetEmail(userToken.Value)
	logger.Log.Printf("Updating interval for %s to %f\n", email, interval.Interval)
	RROnline.UpdateInterval(email, interval.Interval)
	return c.NoContent(http.StatusOK)
}

func editDevice(c echo.Context) error {
	device := new(device)
	if err := c.Bind(device); err != nil {
		logger.Log.Println("Error binding JSON body to device.")
		return c.NoContent(http.StatusBadRequest)
	}
	if _, err := c.Cookie("user_token"); err != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	logger.Log.Printf("Updating device %s to %t\n", device.ID, device.Active)
	RROnline.UpdateDevice(device.ID, device.Active)
	return c.NoContent(http.StatusOK)
}

func validateRoute(c echo.Context) error {
	sub := new(sub)
	if err := c.Bind(sub); err != nil {
		logger.Log.Println("Error binding JSON body to sub.")
		return c.NoContent(http.StatusBadRequest)
	}
	if sub.Sub == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	jsonBody := &isValid{RROnline.ValidateSub(sub.Sub)}
	return c.JSON(http.StatusOK, jsonBody)
}

func handleToken(c echo.Context) error {
	code := c.QueryParam("code")
	sess, _ := session.Get("session", c)
	if code == "" {
		sess.Values["error"] = "Error handling token or 'Deny' was selected"
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, "/")
	}
	userTok := RROnline.GetToken(code)
	_, err := c.Cookie("user_token")
	if err != nil {
		if err.Error() != "http: named cookie not present" {
			RROnline.LogDBError(&RROnline.DBError{Err: "Error getting cookie user_token.\n",
				Reason: err.Error()})
			sess.Values["error"] = "Server error occurred, please try again."
			sess.Save(c.Request(), c.Response())
			return c.Redirect(http.StatusFound, "/")
		}
	}
	cookie := new(http.Cookie)
	cookie.Name = "user_token"
	cookie.Value = userTok
	for _, envVar := range os.Environ() {
		pair := strings.Split(envVar, "=")
		if pair[0] == "SECURE_COOKIES" && pair[1] == "True" {
			cookie.Secure = true
		}
	}
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HttpOnly = true
	email := RROnline.GetEmail(userTok)
	exists, dbErr := RROnline.UserExists(email)
	if dbErr != nil {
		RROnline.LogDBError(dbErr)
		sess.Values["error"] = "Server error occurred, please try again."
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, "/")
	}
	if !exists {
		db, err := RROnline.Connect()
		if err != nil {
			RROnline.LogDBError(err)
			sess.Values["error"] = "Server error occurred, please try again."
			sess.Save(c.Request(), c.Response())
			return c.Redirect(http.StatusFound, "/")
		}
		RROnline.AddUser(email, userTok, db)
		RROnline.RefreshDevices(userTok, db, nil)
	} else {
		RROnline.UpdateUserToken(email, userTok)
	}
	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/searchPage")
}

func index(c echo.Context) error {
	data := make(map[string]interface{})
	sess, _ := session.Get("session", c)
	if err, exists := sess.Values["error"]; exists {
		data["login"] = "Login"
		data["url"] = pushURL
		data["error"] = err.(string)
		delete(sess.Values, "error")
		sess.Save(c.Request(), c.Response())
		return c.Render(http.StatusOK, "index.html", data)
	}
	userToken, err := c.Cookie("user_token")
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

func mainPage(c echo.Context) error {
	userToken, err := c.Cookie("user_token")
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	refreshChan := make(chan *RROnline.DBError)
	db, dbErr := RROnline.Connect()
	// TODO: Same deal as two functions up, should return a 5XX status code or render a special page
	if err != nil {
		RROnline.LogDBError(dbErr)
	}
	go RROnline.RefreshDevices(userToken.Value, db, refreshChan)
	name := RROnline.GetUserName(userToken.Value)
	email := RROnline.GetEmail(userToken.Value)
	searches, dbErr := RROnline.GetSearches(email, db)
	// TODO: Not sure what dbErr is, I think it changes? Either way if it being present undermines the following
	// functions, it should probably return some sort of status code or result in a page render
	if err != nil {
		RROnline.LogDBError(dbErr)
	}
	// TODO: Same as above
	interval, dbErr := RROnline.GetInterval(email)
	if dbErr != nil {
		RROnline.LogDBError(dbErr)
	}
	searchMap := make(map[string][]string)
	for _, search := range searches {
		if _, ok := searchMap[search.Sub]; ok {
			searchMap[search.Sub] = append(searchMap[search.Sub], search.Search)
		} else {
			searchMap[search.Sub] = []string{search.Search}
		}
	}
	dbErr = <-refreshChan
	// TODO: Same as above
	if dbErr != nil {
		RROnline.LogDBError(dbErr)
	}
	devices, dbErr := RROnline.GetDevices(email, db)
	// TODO: Same as above
	if dbErr != nil {
		RROnline.LogDBError(dbErr)
	}
	data := make(map[string]interface{})
	data["name"] = name
	data["searches"] = searchMap
	data["devices"] = devices
	data["interval"] = interval
	data["inc"] = func(i int) int {
		return i + 1
	}
	return c.Render(http.StatusOK, "searchPage.html", data)
}

func addSearch(c echo.Context) error {
	searches := new(searches)
	if err := c.Bind(searches); err != nil {
		logger.Log.Println("Error binding JSON body to searches.")
		return c.NoContent(http.StatusBadRequest)
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
	return c.NoContent(http.StatusOK)
}

func deleteSub(c echo.Context) error {
	sub := new(sub)
	if err := c.Bind(sub); err != nil {
		logger.Log.Println("Error binding JSON body to search.")
		return c.NoContent(http.StatusBadRequest)
	}
	userToken, err := c.Cookie("user_token")
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	email := RROnline.GetEmail(userToken.Value)
	dbErr := RROnline.DeleteSub(email, sub.Sub, routineManager)
	if dbErr != nil {
		RROnline.LogDBError(dbErr)
		return c.NoContent(http.StatusBadRequest)
	}
	return nil
}
