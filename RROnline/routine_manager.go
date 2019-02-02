package RROnline

import (
	"fmt"
	"time"

	"../Reddit-Refresh-Go/reddit_refresh_go/reddit_refresh"
)

type subChanMap map[string]chan<- bool
type emailSubMap map[string]subChanMap

type RoutineManager struct {
	masterMap map[string]emailSubMap
}

func CreateManager() *RoutineManager {
	rm := RoutineManager{}
	rm.masterMap = make(map[string]emailSubMap)
	return &rm
}

func (rm RoutineManager) RMAddSub(token string, sub string, searches []string) {
	if _, ok := rm.masterMap[token]; !ok {
		rm.masterMap[token] = make(emailSubMap)
	}
	for _, search := range searches {
		rm.masterMap[token][sub] = make(subChanMap)
		rm.RMAddSearch(token, sub, search)
	}
}

func (rm RoutineManager) RMAddSearch(token string, sub string, search string) {
	searchChan := make(chan bool)
	fmt.Println("Starting goroutine")
	go CheckResult(token, sub, search, searchChan)
	rm.masterMap[token][sub][search] = searchChan
}

func (rm RoutineManager) RMDeleteSub(token string, sub string) {
	for search := range rm.masterMap[token][sub] {
		rm.RMDeleteSearch(token, sub, search)
	}
}

func (rm RoutineManager) RMDeleteSearch(token string, sub string, search string) {
	rm.masterMap[token][sub][search] <- true
}

func CheckResult(token string, sub string, search string, listen <-chan bool) {
	fmt.Println(token)
	email := GetEmail(token)
	fmt.Println(email)
	//interval := GetInterval(email)
	for {
		fmt.Println("getting old result")
		oldResult := GetLastRes(email, sub, search)
		fmt.Println(oldResult)
		newResult := reddit_refresh.GetResult(sub, search)
		fmt.Println(newResult)
		/*
				if oldResult != newResult.Url {
					fmt.Println(newResult)
						devices := GetDevices(email, nil)
						for _, device := range devices {
							reddit_refresh.SendPushLink(device.DeviceId, token, newResult)
							UpdateLastRes(email, sub, search, newResult.Url)
						}
				}
			select {
			case <-listen:
				return
			case <-time.After(1 * time.Minute):
				//continue
			}
		*/
		time.Sleep(1 * time.Minute)
	}
}
