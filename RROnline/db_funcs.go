package RROnline

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

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
	Interval float32 `db:"interval_min"`
	Token    string  `db:"access_token"`
}

const PASSWD_FILE = "username"
const USER_FILE = "password"
const CONN_STR = "postgres://%s:%s@traphouse.us/reddit_refresh_online"

const DEFAULT_INTERVAL = 10

const SearchQueryAllStr = "SELECT email, sub, search, last_result FROM search"
const SearchQueryStr = "SELECT email, sub, search, last_result " +
	"FROM search WHERE email = $1 ORDER BY create_time"
const SearchIndQueryStr = "SELECT email, sub, search, last_result " +
	"FROM search WHERE email = $1 and sub = $2 and search = $3"
const SearchDelStr = "DELETE FROM search " +
	"WHERE email = ? AND sub = ? AND search NOT IN (?)"
const SearchDelSubStr = "DELETE FROM search " +
	"WHERE email = $1 AND sub = $2"
const SearchInsStr = "INSERT INTO search (email, sub, search, last_result)" +
	"	VALUES ($1, $2, $3, $4)"
const SearchUpdStr = "UPDATE search SET last_result = $1" +
	"	WHERE email = $2 AND sub = $3 AND search = $4"
const DupSearchErr = "pq: duplicate key value violates unique constraint \"search_pk\""

const UserInfoQueryStr = "SELECT email, interval_min, access_token FROM user_info" +
	"	WHERE email = $1"
const UserInfoInsStr = "INSERT INTO user_info (email, interval_min, access_token)" +
	"	VALUES ($1, $2)"
const UserInfoUpdStr = "UPDATE user_info SET access_token = $1" +
	"	WHERE email = $2"

const DevicesInsStr = "INSERT INTO device (email, device_id, nickname)" +
	"	VALUES ($1, $2, $3)"
const DevicesQueryStr = "SELECT email, device_id, nickname, active" +
	"	FROM device WHERE email = $1"
const DevicesDelStr = "DELETE FROM device WHERE device_id = $1"
const DevicesDelAllStr = "DELETE FROM device WHERE email = $1"

/*
Connect returns a sqlx database connection for the database
*/
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

/*
RefreshDevices refreshes the user's devices in the database
token is the user's Pushbullet API token
db is the database to use for connection, or nil
*/
func RefreshDevices(token string, db *sqlx.DB) map[string]string {
	if db == nil {
		db = Connect()
	}
	email := GetEmail(token)
	devices := reddit_refresh.GetDevices(token)
	_, err := db.Exec(DevicesDelAllStr, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting devices for %s", email)
	}
	var wg sync.WaitGroup
	wg.Add(len(devices))
	fmt.Println(devices)
	for nickname, iden := range devices {
		go AddDevice(email, iden, nickname, db, &wg)
	}
	//wait for all AddDevice calls to exit
	wg.Wait()
	return devices
}

/*
GetDevices gets all devices in the DB for a given user
db is the database to use for connection, or nil
*/
func GetDevices(email string, db *sqlx.DB) []Device {
	if db == nil {
		db = Connect()
	}
	devices := []Device{}
	err := db.Select(&devices, DevicesQueryStr, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting devices for %s\n", email)
	}
	return devices
}

/*
AddDevice adds a given device to the database for the specified user
deviceID is the UUID for the device given by Pushbullet
nickname is the nickname that Pushbullet gives
db can be a sqlx DB connection or nil
wg is the WaitGroup to use when this is a goroutine
*/
func AddDevice(email string, deviceID string, nickname string, db *sqlx.DB, wg *sync.WaitGroup) {
	//wait until end of function to tell wait group that we're exiting
	defer wg.Done()
	if db == nil {
		db = Connect()
	}
	_, err := db.Exec(DevicesInsStr, email, deviceID, nickname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inserting device %s for %s\n",
			deviceID, email)
		fmt.Println(err)
	}
}

func DeleteDevice(deviceID string) {
	db := Connect()
	_, err := db.Exec(DevicesDelStr, deviceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting device %s\n", deviceID)
	}
}

func GetAllSearches() []Search {
	db := Connect()
	searches := []Search{}
	err := db.Select(&searches, SearchQueryAllStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting all searches\n")
	}
	return searches
}

func GetSearches(email string, db *sqlx.DB) []Search {
	if db == nil {
		db = Connect()
	}
	searches := []Search{}
	err := db.Select(&searches, SearchQueryStr, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting searches for %s\n", email)
	}
	return searches
}

func GetLastRes(email string, sub string, search string) string {
	db := Connect()
	defer db.Close()
	searches := []Search{}
	err := db.Select(&searches, SearchIndQueryStr, email, sub, search)
	if err != nil || len(searches) == 0 {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Error getting search (%s, %s, %s)",
			email, sub, search))
	}
	return searches[0].LastResult
}

func DeleteMissingSearches(email string, sub string, searches []string) error {
	db := Connect()
	query, args, err := sqlx.In(SearchDelStr, email, sub, searches)
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	_, err = db.Exec(query, args...)
	// TODO: figure out how to delete old search goroutines
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
	_, err := db.Exec(SearchDelSubStr, email, sub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting sub (%s, %s)\n",
			email, sub)
		return errors.New("Could not delete sub")
	}
	// TODO: add code to delete all sub goroutines
	return nil
}

func AddSearch(email string, sub string, search string) {
	db := Connect()
	_, err := db.Exec(SearchInsStr, email, sub, search, "")
	if err != nil {
		if err.Error() != DupSearchErr {
			fmt.Fprintf(os.Stderr, "Error inserting search for %s\n", email)
			fmt.Println(err)
		}
	}
	//TODO: add code to start a search goroutine
}

func UserExists(email string) bool {
	db := Connect()
	users := []UserInfo{}
	err := db.Select(&users, UserInfoQueryStr, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting interval for %s\n", email)
	}
	return len(users) != 0
}

func GetInterval(email string) float32 {
	db := Connect()
	users := []UserInfo{}
	err := db.Select(&users, UserInfoQueryStr, email)
	if err != nil || len(users) == 0 {
		fmt.Fprintf(os.Stderr, "Error getting interval for %s\n", email)
	}
	return users[0].Interval
}

func GetUserToken(email string) string {
	db := Connect()
	users := []UserInfo{}
	err := db.Select(&users, UserInfoQueryStr, email)
	if err != nil || len(users) == 0 {
		fmt.Fprintf(os.Stderr, "Error getting token for %s\n", email)
	}
	return users[0].Token
}

func UpdateUserToken(email string, token string) {
	db := Connect()
	_, err := db.Exec(UserInfoUpdStr, token, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating %s to %s\n", email, token)
	}
}

func UpdateLastRes(email string, sub string, search string, url string) {
	db := Connect()
	_, err := db.Exec(SearchUpdStr, url, email, sub, search)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating (%s, %s, %s) to %s\n",
			email, sub, search, url)
	}
}

func AddUser(email string, interval float32, token string, db *sqlx.DB) {
	if db == nil {
		db = Connect()
	}
	_, err := db.Exec(UserInfoInsStr, email, interval, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user %s\n", email)
	}
}
