package connectordb

import (
	pconfig "config/permissions"
	"connectordb/users"
	"errors"
	"fmt"
	"net/mail"
)

// CountUsers returns the total number of users of the entire database
func (db *Database) CountUsers() (int64, error) {
	return db.Userdb.CountUsers()
}

// ReadAllUsers returns all users of the database. This one will be a pretty expensive operation,
// since there is a virtually unlimited number of users possible, so this should have some type of
// search possibility in the future
func (db *Database) ReadAllUsers() ([]*users.User, error) {
	return db.Userdb.ReadAllUsers()
}

// CreateUser creates a user with the given information. It checks some basic validity
// before creating, and ensures that roles exist/max user amounts are upheld
func (db *Database) CreateUser(name, email, password, role string, public bool) error {
	perm := pconfig.Get()

	if password == "" {
		return errors.New("Password must have at least one character")
	}

	// Make sure that the given email is valid
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("An invalid email address was given")
	}

	// Make sure that the given role exists
	r, ok := perm.UserRoles[role]
	if !ok {
		return fmt.Errorf("The given role '%s' does not exist", role)
	}

	// Make sure that users with this role are allowed to be private if private is set
	if !public && !r.CanBePrivate {
		return fmt.Errorf("Users with role '%s' can't be private.", role)
	}

	return db.Userdb.CreateUser(name, email, password, role, public, perm.MaxUsers)
}

// ReadUserByID reads the user object by ID given
func (db *Database) ReadUserByID(userID int64) (*users.User, error) {
	return db.Userdb.ReadUserById(userID)
}

// ReadUser reads the user object by user name
func (db *Database) ReadUser(username string) (*users.User, error) {
	return db.Userdb.ReadUserByName(username)
}

// UpdateUserByID updates the user with the given UserID with the given data map
func (db *Database) UpdateUserByID(userID int64, update map[string]interface{}) error {
	return errors.New("UNIMPLEMENTED")
}

// DeleteUserByID removes the user with the given UserID. It propagates deletion to add devices
// and streams that the user owns
func (db *Database) DeleteUserByID(userID int64) error {

	// First take all the devices that the user owns, and delete them
	devices, err := db.ReadAllDevicesByUserID(userID)
	if err != nil {
		return err
	}

	for i := range devices {
		err := db.DeleteDeviceByID(devices[i].DeviceID)
		if err != nil {
			return err
		}
	}

	// Lastly, delete the user
	return db.Userdb.DeleteUser(userID)
}
