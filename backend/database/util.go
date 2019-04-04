package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// GenerateKey creates a random API key
func GenerateKey(length int) (string, error) {
	// Prepare the plugin API key
	apikey := make([]byte, length)
	_, err := rand.Read(apikey)
	return base64.StdEncoding.EncodeToString(apikey), err
}

// HashPassword generates a bcrypt hash for the given password
func HashPassword(password string) (string, error) {
	passwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(passwd), err
}

// CheckPassword checks if the password is valid
func CheckPassword(password, hashed string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

var (
	nameValidator = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]*$")
)

func ValidName(name string) error {
	if nameValidator.MatchString(name) && len(name) > 0 {
		return nil
	}
	return ErrInvalidName
}

// Ensures that the avatar is in a valid format
func ValidAvatar(avatar string) error {
	if avatar == "" {
		return nil
	}
	// We permit special avatar prefixes to be used. The first one is material:, which represents material icons
	// that are assumed to be bundled with all applications that display heedy data. The second is fa: which
	// will represent fontawesome avatars in the future
	if strings.HasPrefix(avatar, "mi:") || strings.HasPrefix(avatar, "fa:") {
		if len(avatar) > 30 {
			return errors.New("bad_request: avatar icon name can't be more than 30 characters")
		}
		return nil
	}
	if !strings.HasPrefix(avatar, "data:image/") {
		return errors.New("bad_request: Avatar iamges must be data-urls or be prefixed with mi: or fa:")
	}
	return nil
}

// Checks whether the given group access level is OK
func ValidGroupAccessLevel(level int) error {
	if level < 0 || level > 600 {
		return errors.New("malformed_query: Access levels are 0-600")
	}
	return nil
}

// Performs a set of tests on the result and error of a
// call to see what kind of error we should return.
func getExecError(result sql.Result, err error) error {
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
