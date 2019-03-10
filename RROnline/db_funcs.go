package RROnline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/zedjones/Reddit-Refresh-Go/reddit_refresh_go/reddit_refresh"
)

/*
Device represents a single user device in the database
*/
type Device struct {
	Email    string `db:"email"`
	DeviceID string `db:"device_id"`
	Nickname string `db:"nickname"`
	Active   bool   `db:"active"`
}

/*
Search represents a single user search in the database
*/
type Search struct {
	Email      string `db:"email"`
	Sub        string `db:"sub"`
	Search     string `db:"search"`
	LastResult string `db:"last_result"`
}

/*
UserInfo represents a single user in the database
*/
type UserInfo struct {
	Email    string  `db:"email"`
	Interval float32 `db:"interval_min"`
	Token    string  `db:"access_token"`
	DevRef   string  `db:"last_device_refresh"`
}

type dbConfig struct {
	User string `json:"username"`
	Pass string `json:"password"`
	DB   string `json:"db"`
}

/*
DBError defines the basic error type for database operations
err is the operation that the error occurred on
reason is the reason this error occurred
*/
type DBError struct {
	err    string
	reason string
}

const dbConfFile = "../DBSettings.json"
const connStr = "postgres://%s:%s@traphouse.us/%s"

//DefaultInterval is the default interval for a user when first created
const DefaultInterval = 10

const searchQueryAllStr = "SELECT email, sub, search, last_result FROM search"
const searchQueryStr = "SELECT email, sub, search, last_result " +
	"FROM search WHERE email = $1 ORDER BY create_time"
const searchIndQueryStr = "SELECT email, sub, search, last_result " +
	"FROM search WHERE email = $1 and sub = $2 and search = $3"
const searchDelStr = "DELETE FROM search " +
	"WHERE email = ? AND sub = ? AND search NOT IN (?)"
const searchQueryMissingStr = "SELECT email, sub, search, last_result " +
	"FROM search WHERE email = ? AND sub = ? AND search NOT IN (?)"
const searchDelSubStr = "DELETE FROM search " +
	"WHERE email = $1 AND sub = $2"
const searchInsStr = "INSERT INTO search (email, sub, search, last_result)" +
	"	VALUES ($1, $2, $3, $4)"
const searchUpdStr = "UPDATE search SET last_result = $1" +
	"	WHERE email = $2 AND sub = $3 AND search = $4"
const dupSearchErr = "pq: duplicate key value violates unique constraint \"search_pk\""

const userInfoQueryStr = "SELECT email, interval_min, access_token FROM user_info" +
	"	WHERE email = $1"
const userInfoInsStr = "INSERT INTO user_info (email, access_token)" +
	"	VALUES ($1, $2)"
const userInfoUpdStr = "UPDATE user_info SET access_token = $1" +
	"	WHERE email = $2"
const userInfoUpdInt = "UPDATE user_info SET interval_min = $1" +
	"	WHERE email = $2"

const devicesInsStr = "INSERT INTO device (email, device_id, nickname)" +
	"	VALUES ($1, $2, $3)"
const devicesQueryStr = "SELECT email, device_id, nickname, active" +
	"	FROM device WHERE email = $1"
const devicesDelStr = "DELETE FROM device WHERE device_id = $1"
const devicesDelMissingStr = "DELETE FROM device WHERE email = ?" +
	"	AND device_id NOT IN (?)"
const devicesUpdMissingStr1 = "INSERT INTO device (email, device_id, nickname) VALUES "
const devicesUpdMissingStr2 = "		ON CONFLICT (device_id) DO NOTHING"
const devicesDelAllStr = "DELETE FROM device WHERE email = $1"
const devicesUpdActive = "UPDATE device SET active = $1" +
	"	WHERE device_id = $2"

var dbConf dbConfig

/*
LoadDBConfig loads the configuration for the database from the settngs file
specified by dbConfFile
*/
func LoadDBConfig() *DBError {
	content, err := ioutil.ReadFile(dbConfFile)
	if err != nil {
		return &DBError{"Error reading config file.\n", err.Error()}
	}
	dbConf = dbConfig{}
	if err = json.Unmarshal(content, &dbConf); err != nil {
		return &DBError{"Error unmarshalling config file.\n", err.Error()}
	}
	return nil
}

/*
Connect returns a sqlx database connection for the database
*/
func Connect() (*sqlx.DB, *DBError) {
	_ = pq.Efatal //weird fix for bug with pq
	fullConStr := fmt.Sprintf(connStr, dbConf.User, dbConf.Pass, dbConf.DB)
	db, err := sqlx.Open("postgres", fullConStr)
	if err != nil {
		return nil, &DBError{"Error connecting to PGSQL DB.\n", err.Error()}
	}
	return db, nil
}

/*
RefreshDevices refreshes the user's devices in the database
token is the user's Pushbullet API token
db is the database to use for connection, or nil
*/
func RefreshDevices(token string, db *sqlx.DB, rChan chan *DBError) {
	errors := ""
	reasons := ""
	if db == nil {
		db, err := Connect()
		if err != nil {
			rChan <- err
		}
		defer db.Close()
	}
	email := GetEmail(token)
	devices := reddit_refresh.GetDevices(token)
	//first part of string, before values
	insStr := devicesUpdMissingStr1
	insVals := []interface{}{}
	ids := []string{}
	//add each value to the args... and the sqlStr
	for nickname, iden := range devices {
		insStr += "(?, ?, ?),"
		insVals = append(insVals, email, iden, nickname)
		ids = append(ids, iden)
	}
	insStr = strings.TrimSuffix(insStr, ",")
	//change ? into $n
	insStr = sqlx.Rebind(sqlx.DOLLAR, insStr)
	//add last section about ignoring duplicate values in insert
	insStr += devicesUpdMissingStr2
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		if _, err := db.Exec(insStr, insVals...); err != nil {
			errors += fmt.Sprintf("Error updating devices for %s\n", email)
			reasons += err.Error() + "\n"
		}
		wg.Done()
	}()
	rmStr, delVals, err := sqlx.In(devicesDelMissingStr, email, ids)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error constructing query to deleting missing devices\n")
	}
	rmStr = sqlx.Rebind(sqlx.DOLLAR, rmStr)
	go func() {
		if _, err = db.Exec(rmStr, delVals...); err != nil {
			errors += fmt.Sprintf("Error removing old devices for %s\n", email)
			reasons += err.Error() + "\n"
		}
		wg.Done()
	}()
	wg.Wait()
	//tell main routine that we're done
	if rChan != nil {
		if errors != "" {
			rChan <- &DBError{errors, reasons}
		} else {
			rChan <- nil
		}
	}
}

/*
GetDevices gets all devices in the DB for a given user
db is the database to use for connection, or nil
*/
func GetDevices(email string, db *sqlx.DB) ([]Device, *DBError) {
	if db == nil {
		db, err := Connect()
		if err != nil {
			return nil, err
		}
		defer db.Close()
	}
	devices := []Device{}
	err := db.Select(&devices, devicesQueryStr, email)
	if err != nil {
		return nil, &DBError{fmt.Sprintf("Error getting devices for %s\n", email),
			err.Error()}
	}
	return devices, nil
}

/*
AddDevice adds a given device to the database for the specified user
deviceID is the UUID for the device given by Pushbullet
nickname is the nickname that Pushbullet gives
db can be a sqlx DB connection or nil
wg is the WaitGroup to use when this is a goroutine
*/
func AddDevice(email string, deviceID string, nickname string, db *sqlx.DB, wg *sync.WaitGroup) *DBError {
	//wait until end of function to tell wait group that we're exiting
	defer wg.Done()
	if db == nil {
		db, err := Connect()
		if err != nil {
			return err
		}
		defer db.Close()
	}
	_, err := db.Exec(devicesInsStr, email, deviceID, nickname)
	if err != nil {
		return &DBError{fmt.Sprintf("Error inserting device %s for %s\n",
			deviceID, email), err.Error()}
	}
	return nil
}

/*
DeleteDevice deletes a given device from the database
deviceID is the UUID for the device
NOTE: we don't need the email since each device ID is unique
*/
func DeleteDevice(deviceID string) *DBError {
	db, err := Connect()
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec(devicesDelStr, deviceID); err != nil {
		return &DBError{fmt.Sprintf("Error deleting device %s\n", deviceID),
			err.Error()}
	}
	return nil
}

/*
UpdateDevice updates whether or not a device is active in the database
This decides whether or not it will receive new pushes
NOTE: we don't need the email since each device ID is unique
*/
func UpdateDevice(deviceID string, active bool) *DBError {
	db, err := Connect()
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec(devicesUpdActive, active, deviceID); err != nil {
		return &DBError{fmt.Sprintf("Error updating device %s to %t\n", deviceID, active),
			err.Error()}
	}
	return nil
}

/*
GetAllSearches gets every search from the database, used at program startup
*/
func GetAllSearches() ([]Search, *DBError) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	searches := []Search{}
	if err := db.Select(&searches, searchQueryAllStr); err != nil {
		return nil, &DBError{"Error getting all searches\n", err.Error()}
	}
	return searches, nil
}

/*
GetSearches gets all the searches in the DB for the given user
db is the database to use for connection, or nil
*/
func GetSearches(email string, db *sqlx.DB) ([]Search, *DBError) {
	if db == nil {
		db, err := Connect()
		if err != nil {
			return nil, err
		}
		defer db.Close()
	}
	searches := []Search{}
	if err := db.Select(&searches, searchQueryStr, email); err != nil {
		return nil, &DBError{fmt.Sprintf("Error getting searches for %s\n", email),
			err.Error()}
	}
	return searches, nil
}

/*
GetLastRes gets the last search result for the given search and user
sub is the subreddit that the search is for
*/
func GetLastRes(email string, sub string, search string) (string, *DBError) {
	db, err := Connect()
	if err != nil {
		return "", err
	}
	defer db.Close()
	searches := []Search{}
	if err := db.Select(&searches, searchIndQueryStr, email, sub, search); err != nil || len(searches) == 0 {
		return "", &DBError{fmt.Sprintf("Error getting search (%s, %s, %s)",
			email, sub, search), err.Error()}
	}
	return searches[0].LastResult, nil
}

/*
DeleteMissingSearches deletes all searches in the DB not in the searches slice
searches is a slice containing the new searches
*/
func DeleteMissingSearches(email string, sub string, searches []string, rm RoutineManager) *DBError {
	db, dbErr := Connect()
	if dbErr != nil {
		return dbErr
	}
	defer db.Close()
	//create query to get missing searches
	query, args, err := sqlx.In(searchQueryMissingStr, email, sub, searches)
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	searchStructs := []Search{}
	err = db.Select(&searchStructs, query, args...)
	if err != nil {
		return &DBError{"Error getting missing results", err.Error()}
	}
	for _, item := range searchStructs {
		rm.RMDeleteSearch(email, item.Sub, item.Search)
	}
	//create query to delete missing searches
	query, args, err = sqlx.In(searchDelStr, email, sub, searches)
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	_, err = db.Exec(query, args...)
	if err != nil {
		return &DBError{"Could not delete old searches", err.Error()}
	}
	return nil
}

/*
DeleteSub deletes all the searches in the database for a given user and subreddit
*/
func DeleteSub(email string, sub string, rm RoutineManager) *DBError {
	db, dbErr := Connect()
	if dbErr != nil {
		return dbErr
	}
	defer db.Close()
	_, err := db.Exec(searchDelSubStr, email, sub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting sub (%s, %s)\n",
			email, sub)
		return &DBError{"Could not delete sub", err.Error()}
	}
	rm.RMDeleteSub(email, sub)
	return nil
}

/*
AddSearch adds a single search to the DB for a given user and subreddit
*/
func AddSearch(token string, email string, sub string, search string, rm RoutineManager) *DBError {
	db, dbErr := Connect()
	if dbErr != nil {
		return dbErr
	}
	defer db.Close()
	_, err := db.Exec(searchInsStr, email, sub, search, "")
	if err != nil {
		if err.Error() != dupSearchErr {
			return &DBError{fmt.Sprintf("Error inserting search for %s\n", email),
				err.Error()}
		}
	}
	rm.RMAddSearch(token, sub, search)
	return nil
}

/*
UserExists checks whether or not a user exists in the database
*/
func UserExists(email string) (bool, *DBError) {
	db, dbErr := Connect()
	if dbErr != nil {
		return false, dbErr
	}
	defer db.Close()
	users := []UserInfo{}
	err := db.Select(&users, userInfoQueryStr, email)
	if err != nil {
		return false, &DBError{fmt.Sprintf("Error getting interval for %s\n", email),
			err.Error()}
	}
	return len(users) != 0, nil
}

/*
GetInterval gets the refresh interval for a given user
*/
func GetInterval(email string) (float32, *DBError) {
	db, dbErr := Connect()
	if dbErr != nil {
		return 0, dbErr
	}
	defer db.Close()
	users := []UserInfo{}
	err := db.Select(&users, userInfoQueryStr, email)
	if err != nil || len(users) == 0 {
		return 0, &DBError{fmt.Sprintf("Error getting interval for %s\n", email), err.Error()}
	}
	return users[0].Interval, nil
}

/*
UpdateInterval updates the refresh interval for a given user
*/
func UpdateInterval(email string, interval float32) *DBError {
	db, err := Connect()
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec(userInfoUpdInt, interval, email); err != nil {
		return &DBError{fmt.Sprintf("Error updating interval for %s to %f\n",
			email, interval), err.Error()}
	}
	return nil
}

/*
GetUserToken gets a user's Pushbullet access token from the database,
mostly used for starting threads when the application reboots
*/
func GetUserToken(email string) (string, *DBError) {
	db, dbErr := Connect()
	if dbErr != nil {
		return "", dbErr
	}
	defer db.Close()
	users := []UserInfo{}
	err := db.Select(&users, userInfoQueryStr, email)
	if err != nil || len(users) == 0 {
		return "", &DBError{fmt.Sprintf("Error getting token for %s\n", email), err.Error()}
	}
	return users[0].Token, nil
}

/*
UpdateUserToken updates a user's access token in the database
*/
func UpdateUserToken(email string, token string) *DBError {
	db, dbErr := Connect()
	if dbErr != nil {
		return dbErr
	}
	defer db.Close()
	_, err := db.Exec(userInfoUpdStr, token, email)
	if err != nil {
		return &DBError{fmt.Sprintf("Error updating %s to %s\n", email, token), err.Error()}
	}
	return nil
}

/*
UpdateLastRes updates the last result URL for a search for a given user
*/
func UpdateLastRes(email string, sub string, search string, url string) *DBError {
	db, dbErr := Connect()
	if dbErr != nil {
		return dbErr
	}
	defer db.Close()
	_, err := db.Exec(searchUpdStr, url, email, sub, search)
	if err != nil {
		return &DBError{fmt.Sprintf("Error updating (%s, %s, %s) to %s\n",
			email, sub, search, url), err.Error()}
	}
	return nil
}

/*
AddUser adds a user to the database given their interval and token
*/
func AddUser(email string, token string, db *sqlx.DB) *DBError {
	if db == nil {
		db, err := Connect()
		if err != nil {
			return err
		}
		defer db.Close()
	}
	_, err := db.Exec(userInfoInsStr, email, token)
	if err != nil {
		return &DBError{fmt.Sprintf("Error creating user %s\n", email), err.Error()}
	}
	return nil
}
