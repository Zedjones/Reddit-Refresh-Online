package RROnline

import (
	"time"

	"../Reddit-Refresh-Go/reddit_refresh_go/reddit_refresh"
)

type subChanMap map[string]<-chan bool
type emailSubMap map[string]subChanMap

type RoutineManager struct {
	masterMap map[string]emailSubMap
}

func (rm RoutineManager) RMAddSubreddit(email string, sub string, searches []string) {
	for _, search := range searches {
		continue
	}
}

func CheckResult(token string, email string, sub string, search string, listen <-chan bool) {
	newResult := reddit_refresh.GetResult(sub, search)
	oldResult := GetLastRes(email, sub, search)
	interval := GetInterval(email)
	for {
		select {
		case <-listen:
			return
		}
		if oldResult != newResult.Url {
			devices := GetDevices(email, nil)
			for _, device := range devices {
				reddit_refresh.SendPushLink(device.DeviceId, token, newResult)
				UpdateLastRes(email, sub, search, newResult.Url)
			}
		}
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}
