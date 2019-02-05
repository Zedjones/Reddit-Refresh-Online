package RROnline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

//OAuthURL is the only proper way to name this ever though it doesn't need
//to be exported
const OAuthURL = "https://api.pushbullet.com/oauth2/token"
const aboutURL = "https://www.reddit.com/%s/about.json"
const userURL = "https://api.pushbullet.com/v2/users/me"
const clientID = "PR0sGjjxNmfu8OwRrawv2oxgZllvsDm1"
const clientSecret = "VdoOJb5BVCPNjqD0b02dVrIVZzkVD2oY"
const token = "o.OldUc0rKEAt9xhYaHpfeXlUksvVBNKzv"

/*
ValidateSub checks if a subreddit exists
*/
func ValidateSub(sub string) bool {
	if !strings.Contains(sub, "r/") {
		sub = fmt.Sprintf("r/%s", sub)
	}

	url := fmt.Sprintf(aboutURL, sub)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "reddit-refresh-online-1.0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error retrieving webpage.")
		return false
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	url, ok := result["data"].(map[string]interface{})["url"].(string)
	if ok {
		return true
	}
	return false
}

/*
GetUserName gets the user's name given their Pushbullet access token
*/
func GetUserName(token string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating user request.")
		return ""
	}
	req.SetBasicAuth(token, "")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting user response.")
		return ""
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	name, ok := result["name"].(string)
	if !ok {
		return ""
	}
	return name
}

/*
GetEmail gets the user's email given their Pushbullet access token
*/
func GetEmail(token string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating user request.")
		return ""
	}
	req.SetBasicAuth(token, "")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting user response.")
		return ""
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	email, ok := result["email"].(string)
	if !ok {
		return ""
	}
	return email
}

/*
GetToken gets the user's Pushbullet access token given their OAuth token
*/
func GetToken(token string) string {
	dataMap := make(map[string]string)
	dataMap["client_secret"] = clientSecret
	dataMap["client_id"] = clientID
	dataMap["grant_type"] = "authorization_code"
	dataMap["code"] = token
	jsonBuf, err := json.Marshal(dataMap)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error converting data map in JSON string.")
		return ""
	}
	req, err := http.NewRequest("POST", OAuthURL, bytes.NewBuffer(jsonBuf))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating OAuth request.")
		return ""
	}
	client := &http.Client{}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Token", token)
	resp, err := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	accessToken := result["access_token"].(string)
	return accessToken
}
