package dbmaker

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	//"streamdb/config"
	"streamdb/util"
	"github.com/kardianos/osext"
	//"text/template"
)

var (
	//ErrAlreadyRunning is thrown when a database that is already running is started
	ErrAlreadyRunning = errors.New("It looks like the database is already running. If you know it isn't, remove connectordb.pid")

	//ErrFileNotFound thrown when can't find a necessary file
	ErrFileNotFound = errors.New("A required configuration file was not found")

	//FolderPermissions is the folder permissions to use when creating a new database
	FolderPermissions = os.FileMode(0755)

	//FilePermissions refers to the permissions given to a file that is created
	FilePermissions = os.FileMode(0755)
)



/**
// TODO removeme as this can be auto inferred from the connection string

//GetDatabaseType gets the database type used from the folder structure - in particular, if sqlite is used, then there
//will be an sqlite database. If a postgres folder exists, then dbtype is postgres. It returns ErrUnrecognizedDatabase
//if no database is recognized
func GetDatabaseType(streamdbDirectory string, err error) (string, error) {
	if err != nil {
		return "", err
	}

	if util.PathExists(filepath.Join(streamdbDirectory, sqliteDatabaseName)) {
		return "sqlite", nil
	}
	if util.PathExists(filepath.Join(streamdbDirectory, postgresDatabaseName)) {
		return "postgres", nil
	}
	return "", ErrUnrecognizedDatabase
}

**/
//GenerateConfigReplacements generates the replacement variables to use within configuration files
func GenerateConfigReplacements(streamdbDirectory, procname, iface string, port int) map[string]string {
	m := make(map[string]string)

	if len(iface) == 0 {
		iface = "127.0.0.1"
	}

	m["dbdir"] = streamdbDirectory
	m["port"] = strconv.Itoa(port)
	m["interface"] = iface
	m["logfilepath"] = filepath.Join(streamdbDirectory, procname+".log")
	m["logfile"] = procname + ".log"
	m["pidfilepath"] = filepath.Join(streamdbDirectory, procname+".pid")
	m["pidfile"] = procname + ".pid"

	return m
}


//ConfigPath returns the path to the default StreamDB config templates
func ConfigPath() (string, error) {
	execpath, err := osext.ExecutableFolder()
	return filepath.Join(execpath, "config"), err
}

//CopyConfig copies configuration file template from the default config directory of StreamDB to the database folder
func CopyConfig(streamdbPath, configname string, err error) error {
	if err != nil {
		return err
	}

	templatepath := filepath.Join(streamdbPath, configname)
	cpath, err := ConfigPath()
	defaultTemplate := filepath.Join(cpath, configname)
	if !util.PathExists(defaultTemplate) || err != nil {
		return ErrFileNotFound
	}
	log.Printf("Copying %s from '%s'", configname, defaultTemplate)
	return util.CopyFileContents(defaultTemplate, templatepath, err)
}

//SetConfig sets up the given config file with the setting replacements. If a config template is
//not present in the streamdbpath, it looks in the root executable config directory for templates
func SetConfig(streamdbPath, configname string, replacements map[string]string, err error) (string, error) {
	if err != nil {
		return "", err
	}

	log.Printf("Writing %s", configname)

	templatepath := filepath.Join(streamdbPath, configname)

	if !util.PathExists(templatepath) {
		err = CopyConfig(streamdbPath, configname, err)
	}

	configfilecontents, err := ioutil.ReadFile(templatepath)
	if err != nil {
		return "", err
	}

	//Replace stuff in the config file
	for key, value := range replacements {
		configfilecontents = []byte(strings.Replace(string(configfilecontents), "{{"+key+"}}", value, -1))
	}

	outfile := templatepath + ".tmp"
	err = ioutil.WriteFile(outfile, configfilecontents, FilePermissions)

	return outfile, err
}
