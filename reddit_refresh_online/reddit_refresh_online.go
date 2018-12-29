package reddit_refresh_online

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const ABOUT_URL = "https://www.reddit.com/%s/about.json"
const USER_URL = "https://api.pushbullet.com/v2/users/me"

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
