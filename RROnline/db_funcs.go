package RROnline

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
)

const PASSWD_FILE = "password"
const USER_FILE = "username"
const CONN_STR = "host='traphouse.us'" +
	" dbname=reddit_refresh_online" +
	" user=%s" +
	" password=%s"

func Connect() *sql.DB {
	username, _ := ioutil.ReadFile(USER_FILE)
	password, _ := ioutil.ReadFile(PASSWD_FILE)
	fullConStr := fmt.Sprintf(CONN_STR, username, password)
	db, err := sql.Open("postgres", fullConStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to PGSQL DB.")
	}
	return db
}
