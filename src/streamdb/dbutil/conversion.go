package dbutil

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"text/template"
	//"path/filepath"

	//"github.com/kardianos/osext"
)

const (
	versionString = "DBVersion"
	defaultDbversion = "00000000"
)

// TODO @josephlewis42, add daniel's upgrade for version number on timebatch stuff.
// TODO @josephlewis42, add read markers to timebatchdb on read push to aux table, on update aux delete old

// the database meta type
type meta struct {
	key   string
	value string
}

// Gets the conversion script for the given params.
func getConversion(dbtype string, dbversion string, dropOld bool) (string, error) {
	templateParams := make(map[string]string)

	if dbversion == "" {
		dbversion = defaultDbversion
	}

	templateParams["DBVersion"] = dbversion
	templateParams["DBType"] = dbtype
	if dropOld {
		templateParams["DroppingTables"] = "true"
	} else {
		templateParams["DroppingTables"] = "false"
	}

	if dbtype == "postgres" {
		templateParams["pkey_exp"] = "SERIAL PRIMARY KEY"
	} else {
		templateParams["pkey_exp"] = "INTEGER PRIMARY KEY AUTOINCREMENT"
	}

	conversion_template, err := template.New("modifier").Parse(dbconversion)

	if err != nil {
		return "", err
	}

	var doc bytes.Buffer
	conversion_template.Execute(&doc, templateParams)

	return doc.String(), nil
}

func getSqlite3Location() string {
	return "sqlite3"
}

/** Upgrades the database with the given connection string, returns an error if anything goes wrong.


Note that sqlite3 databases rely on a lot of moving parts and can go very wrong
due to an implementation decision that they can't do more than one statement
in an Exec() call; thus we dump the update to a file, close the database,
invoke sqlite3 with the proper command to execute the sql file.

Postgres just uses the existing connection.

**/
func UpgradeDatabase(cxnstring string, dropold bool) error {

	db, driver, err := OpenSqlDatabase(cxnstring)
	if err != nil {
		return err
	}

	// Check version of database
	version := GetDatabaseVersion(db, driver)
	log.Printf("Upgrading DB From Version: %v\n", version)

	conversionstr, err := getConversion(driver, version, dropold)

	if err != nil {
		return err
	}

	switch driver {
	case SQLITE3:

		sqliteLocation := getSqlite3Location()
		log.Printf("Sqlite Location is: %v\n", sqliteLocation)
		// sqlite doesn't allow direct exec of multiple lines, so we do it
		// from the cli and hope for the best.

		f, err := ioutil.TempFile("", "initdb_")
		if err != nil {
			return err
		}

		// Tempfile says we're responsible for closing and deleting this
		defer f.Close()
		defer os.Remove(f.Name())

		//log.Printf("Doing Conversion, script is:\n%v\n\n", conversionstr)

		_, err = f.WriteString(conversionstr)
		if err != nil {
			return err
		}

		// So we don't get any race conditions on the database
		db.Close()

		// Print sqlite version
		log.Printf("Sqlite Version\n")
		cmd := exec.Command(sqliteLocation, "--version")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()

		// Strip anything from sqlite connection string that isn't a path
		cmd = exec.Command(sqliteLocation, "-init", f.Name(), SqliteURIToPath(cxnstring))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			return err
		}

	case POSTGRES:
		defer db.Close()
		_, err = db.Exec(conversionstr)

		if err != nil {
			return err
		}

	default:
		log.Printf("Unknown Driver %v\n", driver)
		return errors.New("The connection driver is unknown, cowardly failing.")
	}

	return nil
}
