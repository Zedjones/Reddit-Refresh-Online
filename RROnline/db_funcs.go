package RROnline

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Device struct {
	Email    string `db:"email"`
	DeviceId string `db:"device_id"`
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

const SEARCH_QUERY_STR = "SELECT email, sub, search, last_result " +
	"FROM search WHERE email = $1 ORDER BY create_time"
const SEARCH_DEL_STR = "DELETE FROM search WHERE email = $1"
const SEARCH_INS_STR = "INSERT INTO search (email, sub, search, last_result)" +
	"	VALUES ($1, $2, $3, $4)"
const SEARCH_UPD_STR = "UPDATE search SET last_result = $1" +
	"	WHERE email = $2 AND sub = $3 AND search = $4"

const USER_INFO_QUERY_STR = "SELECT email, interval_sec FROM user_info" +
	"	WHERE email = $1"
const USER_INFO_INS_STR = "INSERT INTO user_info (email, interval_sec)" +
	"	VALUES ($1, $2)"

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

func GetSearches(email string) []Search {
	db := Connect()
	searches := []Search{}
	err := db.Select(&searches, SEARCH_QUERY_STR, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting searches for %s\n", email)
	}
	return searches
}

func DeleteSearches(email string) {
	db := Connect()
	_, err := db.Exec(SEARCH_DEL_STR, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting searches for %s\n", email)
	}
}

func AddSearch(email string, sub string, search string) {
	db := Connect()
	_, err := db.Exec(SEARCH_INS_STR, email, sub, search, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inserting search for %s\n", email)
	}
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
