/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"

	psconfig "github.com/connectordb/pipescript/config"

	log "github.com/Sirupsen/logrus"
)

//SqlType is the type of sql database used
const SqlType = "postgres"

// The header that is written to all config files
var configHeader = `/* ConnectorDB Configuration File

To see an explanation of the configuration, please see:

http://connectordb.github.io/docs/config.html

For an explanation of default values:
https://github.com/connectordb/connectordb/blob/master/src/config/defaultconfig.go
	Look at NewConfiguration() which explains defaults.

Particular configuration options:
frontend options: https://github.com/connectordb/connectordb/blob/master/src/config/frontend.go
	These are the options that pertain to the ConnectorDB server (REST API, web, request logging)
permissions: https://github.com/connectordb/connectordb/blob/master/src/config/permissions/permissions.go
	The permissions and access levels for each user type. All user types in the database are required.
	"default" is the built in permission - a separate permissions file is optional

The configuration file supports javascript style comments.

Several options support live reload. Changing them in the configuration file will automatically update the corresponding setting
in ConnectorDB. The ones that are not live-reloadable will not be reloaded (changing these options will not give any message).
*/
`

// Configuration represents the options which are kept in a config file
type Configuration struct {
	Version int  `json:"version"` // The version of the configuration file
	Watch   bool `json:"watch"`   // Whether or not to watch the config file for changes

	// The permissions file (or "default") to use for setting up user access rights
	Permissions string `json:"permissions"`

	// Options pertaining to the frontend server.
	// These are transparent to json, so they appear directly in the main json.
	Frontend

	// Configuration options for a service
	Redis Service     `json:"redis"`
	Nats  Service     `json:"nats"`
	Sql   *SQLService `json:"sql"`

	// The size of batches and chunks to use with the database
	BatchSize int `json:"batchsize"` // BatchSize is the number of datapoints per database entry
	ChunkSize int `json:"chunksize"` // ChunkSize is number of batches per database insert transaction

	// The cache sizes for users/devices/streams
	UseCache        bool  `json:"cache"`         // Whether or not to enable caching
	CacheTimeout    int64 `json:"cache_timeout"` // Whether the cache times out in seconds
	UserCacheSize   int64 `json:"user_cache_size"`
	DeviceCacheSize int64 `json:"device_cache_size"`
	StreamCacheSize int64 `json:"stream_cache_size"`

	// The default algorithm to use for hashing passwords. Options are SHA512 and bcrypt
	// This can be changed during runtime, and the user passwords will upgrade when they log in
	PasswordHash string `json:"password_hash"`

	// The configuration options for pipescript (https://github.com/connectordb/pipescript)
	PipeScript *psconfig.Configuration `json:"pipescript"`
}

// UserMaker: Since we can't import the *actual* UserMaker from users (since that would give an import loop)
// we need to have our own version here - this version doesn't allow recusrive tree creation
type UserMaker struct {
	Name        string `json:"name"`        // The public username of the user
	Nickname    string `json:"nickname"`    // The nickname of the user
	Email       string `json:"email"`       // The user's email address
	Description string `json:"description"` // A public description
	Icon        string `json:"icon"`        // A public icon in a data URI format, should be smallish 100x100?

	Role   string `json:"role,omitempty"` // The user type (permissions level)
	Public bool   `json:"public"`         // Whether the user is public or not

	Password string `json:"password,omitempty"` // A hash of the user's password - it is never actually returned - the json params are used internally
}

// String returns a string representation of the configuration
func (c *Configuration) String() string {
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return "ERROR: " + err.Error()
	}
	return string(b)
}

// Save saves the configuration
func (c *Configuration) Save(filename string) error {
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(configHeader))
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

// Load a configuration from the given file, and ensures that it is valid
func Load(filename string) (*Configuration, error) {
	log.Debugf("Loading configuration from %s", filename)

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to load configuration from '%s': %s", filename, err.Error())
	}

	// To allow comments in the json, we minify the file with js minifer before parsing
	m := minify.New()
	m.AddFunc("text/javascript", js.Minify)
	file, err = m.Bytes("text/javascript", file)
	if err != nil {
		return nil, fmt.Errorf("Failed to load configuration from '%s': %s", filename, err.Error())
	}

	// Set up an empty configuration
	c := &Configuration{}
	err = json.Unmarshal(file, c)
	if err != nil {
		return nil, fmt.Errorf("Failed to load configuration from '%s': %s", filename, err.Error())
	}

	// Before doing anything, we need to change the working directory to that of the config file.
	// We switch back to the current working dir once done validating.
	// Validation takes any file names and converts them to absolute paths.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = os.Chdir(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}
	// Change the directory back on exit
	defer os.Chdir(cwd)

	return c, c.Validate()
}
