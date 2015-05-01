package dbmaker

import (
	"log"
	"streamdb/dbutil"
	"streamdb/users"
	"streamdb/config"
)

//MakeUser creates a user directly from a streamdb directory, without needing to start up all of streamdb
func MakeUser(username, password, email string) error {
	if err := StartSqlDatabase(); err != nil {
		return err
	}
	defer StopSqlDatabase()

	log.Printf("Creating user %s (%s)\n", username, email)

	spath := config.GetDatabaseConnectionString()

	db, driver, err := dbutil.OpenSqlDatabase(spath)
	if err != nil {
		return err
	}

	var udb users.UserDatabase
	udb.InitUserDatabase(db, string(driver))
	return udb.CreateUser(username, email, password)
}
