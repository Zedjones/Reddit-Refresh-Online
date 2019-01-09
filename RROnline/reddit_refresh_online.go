package RROnline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"../Reddit-Refresh-Go/reddit_refresh_go/reddit_refresh"
)

const ABOUT_URL = "https://www.reddit.com/%s/about.json"
const USER_URL = "https://api.pushbullet.com/v2/users/me"
const OAUTH_URL = "https://api.pushbullet.com/oauth2/token"
const CLIENT_ID = "PR0sGjjxNmfu8OwRrawv2oxgZllvsDm1"
const CLIENT_SECRET = "VdoOJb5BVCPNjqD0b02dVrIVZzkVD2oY"
const TOKEN = "o.OldUc0rKEAt9xhYaHpfeXlUksvVBNKzv"

func ValidateSub(sub string) bool {
	if !strings.Contains(sub, "r/") {
		sub = fmt.Sprintf("r/%s", sub)
	}

	url := fmt.Sprintf(ABOUT_URL, sub)
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

func GetUserName(token string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", USER_URL, nil)
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

func GetEmail(token string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", USER_URL, nil)
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

func GetToken(token string) string {
	dataMap := make(map[string]string)
	dataMap["client_secret"] = CLIENT_SECRET
	dataMap["client_id"] = CLIENT_ID
	dataMap["grant_type"] = "authorization_code"
	dataMap["code"] = token
	jsonBuf, err := json.Marshal(dataMap)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error converting data map in JSON string.\n")
		return ""
	}
	req, err := http.NewRequest("POST", OAUTH_URL, bytes.NewBuffer(jsonBuf))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating OAuth request.\n")
		return ""
	}
	client := &http.Client{}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Token", TOKEN)
	resp, err := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	accessToken := result["access_token"].(string)
	return accessToken
}

func CheckResult(token string, email string, sub string, search string) {
	newResult := reddit_refresh.GetResult(sub, search)
	oldResult := GetLastRes(email, sub, search)
	if oldResult != newResult.Url {
		devices := GetDevices(email, nil)
		for _, device := range devices {
			reddit_refresh.SendPushLink(device.DeviceId, token, newResult)
			UpdateLastRes(email, sub, search, newResult.Url)
		}
	}
}
