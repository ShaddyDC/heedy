/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"

	"github.com/gorilla/securecookie"
)

// Validate takes a session and makes sure that all of the keys and fields are set up correctly
func (s *Session) Validate() error {
	if s.AuthKey == "" {
		sessionAuthkey := securecookie.GenerateRandomKey(64)
		s.AuthKey = base64.StdEncoding.EncodeToString(sessionAuthkey)
	}
	if s.EncryptionKey == "" {
		sessionEncKey := securecookie.GenerateRandomKey(32)
		s.EncryptionKey = base64.StdEncoding.EncodeToString(sessionEncKey)
	}

	if s.MaxAge < 0 {
		return errors.New("Max Age for cookie must be >=0")
	}

	return nil
}

// Validate takes a frontend and ensures that all the necessary configuration fields are set up
// correctly.
func (f *Frontend) Validate() (err error) {

	if f.TLSEnabled() {
		// If both key and cert are given, assume that we want to use TLS
		_, err = tls.LoadX509KeyPair(f.TLSCert, f.TLSKey)
		if err != nil {
			return err
		}

		//Set the file paths to be full paths
		f.TLSCert, err = filepath.Abs(f.TLSCert)
		if err != nil {
			return err
		}
		f.TLSKey, err = filepath.Abs(f.TLSKey)
		if err != nil {
			return err
		}
	}

	// Validate the Session
	if err = f.Session.Validate(); err != nil {
		return err
	}

	// Set up the optional configuration parameters

	if f.Hostname == "" {
		f.Hostname, err = os.Hostname()
		if err != nil {
			f.Hostname = "localhost"
		}
	}

	if f.Domain == "" {
		f.Domain = f.Hostname
	}

	return nil
}

// Validate takes a configuration and makes sure that it is set up correctly for use in the ConnectorDB
// database. It returns nil if the configuration is valid, and returns an error if an error was found.
// Validate also sets up any missing values to their defaults if they are not required.
func (c *Configuration) Validate() error {
	// First, make sure that the frontend is valid
	if c.Version != 1 {
		return errors.New("This version of ConnectorDB only accepts configuration version 1")
	}

	if c.BatchSize <= 0 {
		return errors.New("Batch size must be >=0")
	}
	if c.ChunkSize <= 0 {
		return errors.New("Chunk size must be >=0")
	}

	// Now let's validate the frontend
	if err := c.Frontend.Validate(); err != nil {
		return err
	}

	return nil
}
