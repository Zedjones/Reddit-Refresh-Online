package RROnline

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"../Reddit-Refresh-Go/reddit_refresh_go/reddit_refresh"
)

type Device struct {
	Email    string `db:"email"`
	DeviceId string `db:"device_id"`
	Nickname string `db:"nickname"`
	Active   bool   `db:"active"`
}

type Search struct {
	Email      string `db:"email"`
	Sub        string `db:"sub"`
	Search     string `db:"search"`
	LastResult string `db:"last_result"`
}

type UserInfo struct {
	Email    string  `db:"email"`
	Interval float32 `db:"interval_sec"`
}

const PASSWD_FILE = "username"
const USER_FILE = "password"
const CONN_STR = "postgres://%s:%s@traphouse.us/reddit_refresh_online"

const DEFAULT_INTERVAL = 600

const SEARCH_QUERY_STR = "SELECT email, sub, search, last_result " +
	"FROM search WHERE email = $1 ORDER BY create_time"
const SEARCH_IND_QUERY_STR = "SELECT email, sub, search, last_result " +
	"FROM search WHERE email = $1 and sub = $2 and search = $3"
const SEARCH_DEL_STR = "DELETE FROM search " +
	"WHERE email = ? AND sub = ? AND search NOT IN (?)"
const SEARCH_DEL_SUB_STR = "DELETE FROM search " +
	"WHERE email = $1 AND sub = $2"
const SEARCH_INS_STR = "INSERT INTO search (email, sub, search, last_result)" +
	"	VALUES ($1, $2, $3, $4)"
const SEARCH_UPD_STR = "UPDATE search SET last_result = $1" +
	"	WHERE email = $2 AND sub = $3 AND search = $4"
const DUP_SEARCH_ERR = "pq: duplicate key value violates unique constraint \"search_pk\""

const USER_INFO_QUERY_STR = "SELECT email, interval_sec FROM user_info" +
	"	WHERE email = $1"
const USER_INFO_INS_STR = "INSERT INTO user_info (email, interval_sec)" +
	"	VALUES ($1, $2)"

const DEVICES_INS_STR = "INSERT INTO device (email, device_id, nickname)" +
	"	VALUES ($1, $2, $3)"
const DEVICES_QUERY_STR = "SELECT email, device_id, nickname, active" +
	"	FROM device WHERE email = $1"
const DEVICE_DEL_STR = "DELETE FROM device WHERE device_id = $1"
const DEVICE_DEL_ALL_STR = "DELETE FROM device WHERE email = $1"

func Connect() *sqlx.DB {
	username, _ := ioutil.ReadFile(USER_FILE)
	password, _ := ioutil.ReadFile(PASSWD_FILE)
	_ = pq.Efatal //weird fix for bug with pq
	fullConStr := fmt.Sprintf(CONN_STR, string(password), string(username))
	db, err := sqlx.Open("postgres", fullConStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to PGSQL DB.\n")
	}
	return db
}

func RefreshDevices(token string) map[string]string {
	db := Connect()
	email := GetEmail(token)
	devices := reddit_refresh.GetDevices(token)
	_, err := db.Exec(DEVICE_DEL_ALL_STR, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting devices for %s", email)
	}
	for nickname, iden := range devices {
		AddDevice(email, iden, nickname)
	}
	return devices
}

func GetDevices(email string) []Device {
	db := Connect()
	devices := []Device{}
	err := db.Select(&devices, DEVICES_QUERY_STR, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting devices for %s\n", email)
	}
	return devices
}

func AddDevice(email string, deviceID string, nickname string) {
	db := Connect()
	_, err := db.Exec(DEVICES_INS_STR, email, deviceID, nickname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inserting device %s for %s\n",
			deviceID, email)
		fmt.Println(err)
	}
}

func DeleteDevice(deviceID string) {
	db := Connect()
	_, err := db.Exec(DEVICE_DEL_STR, deviceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting device %s\n", deviceID)
	}
}

func GetSearches(email string) []Search {
	db := Connect()
	searches := []Search{}
	err := db.Select(&searches, SEARCH_QUERY_STR, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting searches for %s\n", email)
	}
	return searches
}

func GetLastRes(email string, sub string, search string) string {
	db := Connect()
	searches := []Search{}
	err := db.Select(&searches, SEARCH_IND_QUERY_STR, email, sub, search)
	if err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Error getting search (%s, %s, %s)",
			email, sub, search))
	}
	return searches[0].LastResult
}

func DeleteMissingSearches(email string, sub string, searches []string) error {
	db := Connect()
	query, args, err := sqlx.In(SEARCH_DEL_STR, email, sub, searches)
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	_, err = db.Exec(query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting old searches for (%s, %s)\n",
			email, sub)
		fmt.Println(err)
		return errors.New("Could not delete old searches")
	}
	return nil
}

func DeleteSub(email string, sub string) error {
	db := Connect()
	_, err := db.Exec(SEARCH_DEL_SUB_STR, email, sub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting sub (%s, %s)\n",
			email, sub)
		return errors.New("Could not delete sub")
	}
	return nil
}

func AddSearch(email string, sub string, search string) {
	db := Connect()
	_, err := db.Exec(SEARCH_INS_STR, email, sub, search, "")
	if err != nil {
		if err.Error() != DUP_SEARCH_ERR {
			fmt.Fprintf(os.Stderr, "Error inserting search for %s\n", email)
			fmt.Println(err)
		}
	}
	//TODO: add code to start a search thread
}

func UserExists(email string) bool {
	db := Connect()
	users := []UserInfo{}
	err := db.Select(&users, USER_INFO_QUERY_STR, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting interval for %s\n", email)
	}
	return len(users) != 0
}

func GetInterval(email string) float32 {
	db := Connect()
	users := []UserInfo{}
	err := db.Select(&users, USER_INFO_QUERY_STR, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting interval for %s\n", email)
	}
	return users[0].Interval
}

func UpdateLastRes(email string, sub string, search string, url string) {
	db := Connect()
	_, err := db.Exec(SEARCH_UPD_STR, url, email, sub, search)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating (%s, %s, %s) to %s\n",
			email, sub, search, url)
	}
}

func AddUser(email string, interval float32) {
	db := Connect()
	_, err := db.Exec(USER_INFO_INS_STR, email, interval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user %s\n", email)
	}
}
