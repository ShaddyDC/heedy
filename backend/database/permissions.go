package database

import (
	"database/sql"
	"fmt"
)

func createUser(adb *AdminDB, u *User, sqlStatement string, args ...interface{}) error {
	// Only create the user if I have the users:create scope
	rows, err := adb.DB.Query(sqlStatement, args...)

	if err != nil {
		return err
	}
	canCreate := rows.Next()
	rows.Close()
	if !canCreate {
		return ErrAccessDenied("You do not have sufficient permissions to create users")
	}
	return adb.CreateUser(u)
}

func readUser(adb *AdminDB, name string, o *ReadUserOptions, selectStatement string, args ...interface{}) (*User, error) {
	u := &User{}
	err := adb.Get(u, selectStatement, args...)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if o == nil || !o.Avatar {
		u.Avatar = nil
	}
	return u, err
}

// given a user, performs a user update. It is given the sql that returns the distinct set of "scope LIKE 'user:edit%'" scopes
func updateUser(adb *AdminDB, u *User, scopeSQL string, args ...interface{}) error {
	userColumns, userValues, err := userUpdateQuery(u)
	if err != nil {
		return err
	}

	tx, err := adb.DB.Beginx()
	if err != nil {
		return err
	}

	// Make sure that the public has the necessary permissions
	var scopes []string
	err = tx.Select(&scopes, scopeSQL, args...)

	if err != nil {
		return err
	}

	hasEdit := false
	hasEditPassword := false
	hasEditName := false
	for _, scope := range scopes {
		switch scope {
		case "users:edit", "user:edit":
			hasEdit = true
		case "users:edit:password", "user:edit:password":
			hasEditPassword = true
		case "users:edit:name", "user:edit:name":
			hasEditName = true
		}
	}
	if u.Name != nil && !hasEditName || u.Password != nil && !hasEditPassword {
		tx.Rollback()
		return ErrAccessDenied("Editing names or passwords requires the user:edit:name or user:edit:password scopes, which you don't have.")
	}
	if !hasEdit {
		// There is the possibility that we *only* changed the password or username
		totalColumns := 1
		if u.Name != nil {
			totalColumns++
		}
		if u.Password != nil {
			totalColumns++
		}
		if totalColumns < len(userValues) {
			tx.Rollback()
			return ErrAccessDenied("You do not have permission to edit the user '%s'", u.ID)
		}

		// Otherwise, we let the update continue

	}

	// This needs to be first, in case user name is modified - the query will use old name here, and the ID will be cascaded to group owners

	// This uses a join to make sure that the group is in fact an existing user
	result, err := tx.Exec(fmt.Sprintf("UPDATE users SET %s WHERE name=?;", userColumns), userValues...)
	err = getExecError(result, err)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func delUser(adb *AdminDB, name string, sqlStatement string, args ...interface{}) error {
	result, err := adb.DB.Exec(sqlStatement, args...)
	return getExecError(result, err)
}
