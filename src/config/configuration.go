/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/nu7hatch/gouuid"

	log "github.com/Sirupsen/logrus"
)

//SqlType is the type of sql database used
const SqlType = "postgres"

// Configuration represents the options which are kept in a config file
type Configuration struct {
	Version int `json:"version"` // The version of the configuration file

	// Options pertaining to the frontend server.
	Frontend

	// Configuration options for a service
	Redis Service `json:"redis"`
	Nats  Service `json:"nats"`
	Sql   Service `json:"sql"`

	// The size of batches and chunks to use with the database
	BatchSize int `json:"batchsize"` // BatchSize is the number of datapoints per database entry
	ChunkSize int `json:"chunksize"` // ChunkSize is number of batches per database insert transaction

	//These are optional - if they are set, an initial user is created on Create()
	//They are used only when passing a Configuration object to Create()
	InitialUsername     string `json:"-"`
	InitialUserPassword string `json:"-"`
	InitialUserEmail    string `json:"-"`

	// The given usernames are forbidden.
	DisallowedNames []string `json:"disallow_names"` //The names that are not permitted

	// The email suffixes that are permitted during user creation
	AllowedEmailSuffixes []string `json:"allowed_email_suffixes"`

	// The following are exported fields that are used internally, and are not available to json.
	// This is honestly just lazy programming on my part - I am using the config struct as a temporary variable
	// placeholder when creating/starting a database... So technically it is part of the configuration, but it is
	// given explicitly as part of the command line args
	DatabaseDirectory string `json:"-"`
}

// NewConfiguration generates a configuration with reasonable defaults for use in ConnectorDB
func NewConfiguration() *Configuration {
	redispassword, _ := uuid.NewV4()
	natspassword, _ := uuid.NewV4()

	sessionAuthKey := securecookie.GenerateRandomKey(64)
	sessionEncKey := securecookie.GenerateRandomKey(32)

	return &Configuration{
		Version: 1,
		Redis: Service{
			Hostname: "localhost",
			Port:     6379,
			Password: redispassword.String(),
			Enabled:  true,
		},
		Nats: Service{
			Hostname: "localhost",
			Port:     4222,
			Username: "connectordb",
			Password: natspassword.String(),
			Enabled:  true,
		},
		Sql: Service{
			Hostname: "localhost",
			Port:     52592,
			//TODO: Have SQL accedd be auth'd
			Enabled: true,
		},

		Frontend: Frontend{
			Hostname: "0.0.0.0", // Host on all interfaces by default
			Port:     8000,

			Enabled: true,

			// Sets up the session cookie keys that are used
			Session: Session{
				AuthKey:       base64.StdEncoding.EncodeToString(sessionAuthKey),
				EncryptionKey: base64.StdEncoding.EncodeToString(sessionEncKey),
				MaxAge:        60 * 60 * 24 * 30 * 4, //About 4 months is the default expiration time of a cookie
			},
		},

		//The defaults to use for the batch and chunks
		BatchSize: 250,
		ChunkSize: 5,

		// Disallowed names are names that would conflict with the ConnectorDB frontend
		DisallowedNames: []string{"support", "www", "api", "app", "favicon.ico", "robots.txt", "sitemap.xml", "join", "login"},
	}
}

// GetSqlConnectionString returns the string used to connect to postgres
func (c *Configuration) GetSqlConnectionString() string {
	return c.Sql.GetSqlConnectionString()
}

// IsAllowedUsername checks if the user name is allowed by the configuration
func (c *Configuration) IsAllowedUsername(name string) bool {
	for i := range c.DisallowedNames {
		if name == c.DisallowedNames[i] {
			return false
		}
	}
	return true
}

// IsAllowedEmail checks in the configuration's AllowedEmailSuffixes to see if
// the given email address is valid allowed to sign up.
func (c *Configuration) IsAllowedEmail(emailAddress string) bool {
	if len(c.AllowedEmailSuffixes) == 0 {
		return true
	}

	for _, suffix := range c.AllowedEmailSuffixes {
		if strings.HasSuffix(emailAddress, suffix) {
			return true
		}
	}

	return false
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
	return ioutil.WriteFile(filename, b, os.FileMode(0755))
}

// Load a configuration from the given file, and ensures that it is valid
func Load(filename string) (*Configuration, error) {
	log.Debugf("Loading configuration from %s", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := NewConfiguration()
	err = json.Unmarshal(file, c)
	if err != nil {
		return nil, err
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
