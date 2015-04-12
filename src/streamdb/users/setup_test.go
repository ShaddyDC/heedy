package users

/**
This file provides the initialization of the test procedures

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved
**/

import (
    "strconv"
//    "os"
//    "streamdb/dbutil"
//    "log"
    )

var (
    nextNameId = 0
    nextEmailId = 0

	testdb     *UserDatabase
	testdbname = "testing.sqlite3"

    TEST_PASSWORD = "P@$$W0Rd123"
)

func GetNextName() string {
    nextNameId++
    return "name_" + strconv.Itoa(nextNameId)
}

func GetNextEmail() string {
    nextEmailId++
    return "name" + strconv.Itoa(nextNameId) + "@domain.com"
}

/**

func init() {
	var err error

    // may not work if postgres
	_ = os.Remove(testdbname)

    // Init the db
    err = dbutil.UpgradeDatabase(testdbname, true)
    if err != nil {
        log.Panic("Could not set up db for testing: ", err.Error())
    }


	testdb = &UserDatabase{}

	sql, dbtype, err := dbutil.OpenSqlDatabase(testdbname)
	if err != nil {
		log.Panic(err)
	}

	testdb.InitUserDatabase(sql, dbtype.String())

    CleanTestDB()
}

func CleanTestDB(){
    testdb.Exec("DELETE * FROM PhoneCarriers;")
    testdb.Exec("DELETE * FROM Users;")
    testdb.Exec("DELETE * FROM Devices;")
    testdb.Exec("DELETE * FROM Streams;")
    testdb.Exec("DELETE * FROM timeseriestable;")
    testdb.Exec("DELETE * FROM UserKeyValues;")
    testdb.Exec("DELETE * FROM DeviceKeyValues;")
    testdb.Exec("DELETE * FROM StreamKeyValues;")
}


func CreateTestUser() (*User, error) {
    name := GetNextName()
    email := GetNextEmail()

    log.Printf("Creating test user with name: %v, email: %v, pass: %v", name, email, TEST_PASSWORD)

    err := testdb.CreateUser(name, email, TEST_PASSWORD)

    if err != nil {
        return nil, err
    }

    return testdb.ReadUserByName(name)
}


func CreateTestDevice(usr *User) (*Device, error) {
    name := GetNextName()
    err := testdb.CreateDevice(name, usr.UserId)
    if err != nil {
        return nil, err
    }

    return testdb.ReadDeviceForUserByName(usr.UserId, name)
}

func CreateTestStream(dev *Device) (*Stream, error) {
    name := GetNextName()
    err := testdb.CreateStream(name, "", dev.DeviceId)
    if err != nil {
        return nil, err
    }

    return testdb.ReadStreamByDeviceIdAndName(dev.DeviceId, name)
}
**/
