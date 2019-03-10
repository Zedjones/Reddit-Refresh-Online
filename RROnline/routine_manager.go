package RROnline

import (
	"fmt"
	"time"

	"github.com/zedjones/Reddit-Refresh-Go/reddit_refresh_go/reddit_refresh"
)

//subChanMap maps searches to channels
type subChanMap map[string]chan<- bool

//emailSubMap maps emails (users) to subChanMaps
type emailSubMap map[string]subChanMap

/*
RoutineManager controls operations on all goroutines for checking search results
*/
type RoutineManager struct {
	masterMap map[string]emailSubMap
}

/*
CreateManager make a new RoutineManager and initializes the masterMap
*/
func CreateManager() *RoutineManager {
	rm := RoutineManager{}
	rm.masterMap = make(map[string]emailSubMap)
	return &rm
}

/*
RMAddSub creates the goroutines for a subreddit and list of searches,
also creating the necessary maps
*/
func (rm RoutineManager) RMAddSub(token string, sub string, searches []string) {
	email := GetEmail(token)
	if _, ok := rm.masterMap[email]; !ok {
		rm.masterMap[email] = make(emailSubMap)
	}
	for _, search := range searches {
		if _, ok := rm.masterMap[email][sub]; !ok {
			rm.masterMap[email][sub] = make(subChanMap)
		}
		rm.RMAddSearch(token, sub, search)
	}
}

/*
RMAddSearch creates a goroutine for the provided subreddit and search,
also creating the necesssary maps
*/
func (rm RoutineManager) RMAddSearch(token string, sub string, search string) {
	email := GetEmail(token)
	if _, ok := rm.masterMap[email]; !ok {
		rm.masterMap[email] = make(emailSubMap)
	}
	if _, ok := rm.masterMap[email][sub]; !ok {
		rm.masterMap[email][sub] = make(subChanMap)
	}
	if _, ok := rm.masterMap[email][sub][search]; !ok {
		searchChan := make(chan bool)
		go checkResult(token, sub, search, searchChan)
		rm.masterMap[email][sub][search] = searchChan
	}
}

/*
RMDeleteSub kills all goroutines for a given subreddit
*/
func (rm RoutineManager) RMDeleteSub(email string, sub string) {
	for search := range rm.masterMap[email][sub] {
		rm.RMDeleteSearch(email, sub, search)
	}
}

/*
RMDeleteSearch kills a search goroutine for the given user by sending a
signal over the channel
*/
func (rm RoutineManager) RMDeleteSearch(email string, sub string, search string) {
	rm.masterMap[email][sub][search] <- true
}

/*
checkResult checks the latest result for the search term and compares it to the
last result stored in the database. If it's different (i.e. newer), then it
sends a push to each of the user's active devices and updates the last
result in the database to the new URL.
*/
func checkResult(token string, sub string, search string, listen <-chan bool) {
	email := GetEmail(token)
	for {
		interval, err := GetInterval(email)
		if err != nil {
			//TODO: Logging
		}
		oldResult, err := GetLastRes(email, sub, search)
		if err != nil {
			//TODO: Logging
		}
		newResult := reddit_refresh.GetResult(sub, search)
		if oldResult != newResult.Url {
			devices, err := GetDevices(email, nil)
			if err != nil {
				//TODO: Logging
			}
			for _, device := range devices {
				if device.Active {
					reddit_refresh.SendPushLink(device.DeviceID, token, newResult)
				}
			}
			UpdateLastRes(email, sub, search, newResult.Url)
		}
		//either we get a value over the channel or the temporary timeout
		//channel returns
		select {
		case <-listen:
			fmt.Println("Deleting search: " + search)
			return
		case <-time.After(time.Duration(interval*60) * time.Second):
			continue
		}
	}
}

/*
A testing version of checkResult that only prints the new result and doesn't
send a push or update the link in the database
*/
func checkResultTesting(token string, sub string, search string, listen <-chan bool) {
	email := GetEmail(token)
	interval, err := GetInterval(email)
	if err != nil {
		fmt.Printf(err.err, err.reason)
	}
	for {
		oldResult, err := GetLastRes(email, sub, search)
		if err != nil {
			fmt.Printf(err.err, err.reason)
		}
		newResult := reddit_refresh.GetResult(sub, search)
		fmt.Println(oldResult, newResult)
		select {
		case <-listen:
			fmt.Println("Deleting search: " + search)
			return
		case <-time.After(time.Duration(interval*60) * time.Second):
			continue
		}
	}
}
