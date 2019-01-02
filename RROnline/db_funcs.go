package RROnline

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jmoiron/sqlx"
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

const PASSWD_FILE = "password"
const USER_FILE = "username"
const CONN_STR = "host='traphouse.us'" +
	" dbname=reddit_refresh_online" +
	" user=%s" +
	" password=%s"

const SEARCH_QUERY_STR = "SELECT email, sub, search FROM search ORDER BY create_time"
const SEARCH_DEL_STR = "DELETE FROM search WHERE email = %s"
const SEARCH_INS_STR = "INSERT INTO search (email, sub, search, last_result)" +
	"	VALUES (%s, %s, %s, %s)"
const SEARCH_UPD_STR = "UPDATE search SET last_result = %s" +
	"	WHERE email = %s AND sub = %s AND search = %s"

const USER_INFO_QUERY_STR = "SELECT email, interval_sec FROM user_info" +
	"	WHERE email=%s"

func Connect() *sqlx.DB {
	username, _ := ioutil.ReadFile(USER_FILE)
	password, _ := ioutil.ReadFile(PASSWD_FILE)
	fullConStr := fmt.Sprintf(CONN_STR, username, password)
	db, err := sqlx.Open("postgres", fullConStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to PGSQL DB.")
	}
	return db
}

func GetSearches(email string) []Search {
	db := Connect()
	searches := []Search{}
	err := db.Select(&searches, SEARCH_QUERY_STR)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting searches for %s", email)
	}
	return searches
}

func DeleteSearches(email string) {
	db := Connect()
	_, err := db.Exec(SEARCH_DEL_STR, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting searches for %s", email)
	}
}

func AddSearch(email string, sub string, search string) {
	db := Connect()
	_, err := db.Exec(SEARCH_INS_STR, email, sub, search, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inserting search for %s", email)
	}
}

func GetInterval(email string) float32 {
	db := Connect()
	user := UserInfo{}
	err := db.Select(user, USER_INFO_QUERY_STR, email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting interval for %s", email)
	}
	return user.Interval
}

func UpdateLastRes(email string, sub string, search string, url string) {
	db := Connect()
	_, err := db.Exec(SEARCH_UPD_STR, url, email, sub, search)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating (%s, %s, %s) to %s",
			email, sub, search, url)
	}
}
