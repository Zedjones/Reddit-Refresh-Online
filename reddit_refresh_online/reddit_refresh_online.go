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
