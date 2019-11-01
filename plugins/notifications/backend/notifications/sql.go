package notifications

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/heedy/heedy/backend/database"
)

var SQLVersion = 1

const sqlSchema = `
	-- We split up the schema into 3 tables due to issues with UNIQUE when certain values are NULL.
	-- We need apps/sources to be nullable to represent notifications for users/apps
	-- https://stackoverflow.com/questions/22699409/sqlite-null-and-unique

	CREATE TABLE notifications_user (
		user VARCHAR NOT NULL,
		key VARCHAR NOT NULL,

		title VARCHAR NOT NULL,
		description VARCHAR NOT NULL DEFAULT '',
		type VARCHAR NOT NULL DEFAULT 'info',
		timestamp REAL NOT NULL,

		-- User notifications are global=true
		global BOOLEAN NOT NULL DEFAULT true,
		seen BOOLEAN NOT NULL DEFAULT false,

		CONSTRAINT pk PRIMARY KEY (user,key),

		CONSTRAINT user_c
			FOREIGN KEY (user)
			REFERENCES users(username)
			ON UPDATE CASCADE
			ON DELETE CASCADE
	);

	CREATE TABLE notifications_app (
		user VARCHAR NOT NULL,
		app VARCHAR NOT NULL,
		key VARCHAR NOT NULL,

		title VARCHAR NOT NULL,
		description VARCHAR NOT NULL DEFAULT '',
		type VARCHAR NOT NULL DEFAULT 'info',
		timestamp REAL NOT NULL,

		global BOOLEAN NOT NULL DEFAULT false,
		seen BOOLEAN NOT NULL DEFAULT false,

		CONSTRAINT pk PRIMARY KEY (user,app,key),

		CONSTRAINT user_c
			FOREIGN KEY (user)
			REFERENCES users(username)
			ON UPDATE CASCADE
			ON DELETE CASCADE,

		CONSTRAINT app_c
			FOREIGN KEY (app)
			REFERENCES apps(id)
			ON UPDATE CASCADE
			ON DELETE CASCADE
	);

	CREATE TABLE notifications_source (
		user VARCHAR NOT NULL,
		app VARCHAR NOT NULL,
		source VARCHAR NOT NULL,
		key VARCHAR NOT NULL,

		title VARCHAR NOT NULL,
		description VARCHAR NOT NULL DEFAULT '',
		type VARCHAR NOT NULL DEFAULT 'info',
		timestamp REAL NOT NULL,

		global BOOLEAN NOT NULL DEFAULT false,
		seen BOOLEAN NOT NULL DEFAULT false,

		CONSTRAINT pk PRIMARY KEY (user,app,source,key),

		CONSTRAINT user_c
			FOREIGN KEY (user)
			REFERENCES users(username)
			ON UPDATE CASCADE
			ON DELETE CASCADE,

		CONSTRAINT app_c
			FOREIGN KEY (app)
			REFERENCES apps(id)
			ON UPDATE CASCADE
			ON DELETE CASCADE,

		CONSTRAINT source_c
			FOREIGN KEY (source)
			REFERENCES sources(id)
			ON UPDATE CASCADE
			ON DELETE CASCADE
	);
`

// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, curversion int) error {
	if curversion == SQLVersion {
		return nil
	}
	if curversion != 0 {
		return errors.New("Notifications database version too new")
	}
	_, err := db.ExecUncached(sqlSchema)
	return err
}

var ErrAccessDenied = errors.New("access_denied: You don't have necessary permissions for the given query")

type Notification struct {
	Key       string  `json:"key,omitempty"`
	Timestamp float64 `json:"timestamp,omitempty"`

	User       *string `json:"user,omitempty"`
	App *string `json:"app,omitempty"`
	Source     *string `json:"source,omitempty"`

	Type        *string `json:"type,omitempty"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`

	Seen   *bool `json:"seen,omitempty"`
	Global *bool `json:"global,omitempty"`
}

type NotificationsQuery struct {
	User       *string `json:"user,omitempty" schema:"user"`
	App *string `json:"app,omitempty" schema:"app"`
	Source     *string `json:"source,omitempty" schema:"source"`

	Global *bool   `json:"global,omitempty" schema:"global"`
	Seen   *bool   `json:"seen,omitempty" schema:"seen"`
	Key    *string `json:"key,omitempty" schema:"key"`

	Type *string `json:"type,omitempty"`

	// Whether  or not to include self when * present. For example {user="test",app="*"}
	// is unclear whether the user's notifications should be included or not. False by default
	IncludeSelf *bool `json:"include_self,omitempty" schema:"include_self"`
}

func queryAllowed(db database.DB, o *NotificationsQuery) (*NotificationsQuery, error) {
	if o == nil {
		o = &NotificationsQuery{}
	}
	if o.IncludeSelf == nil {
		includeSelfDefault := false
		o.IncludeSelf = &includeSelfDefault
	}

	dbid := db.ID()
	if dbid == "public" {
		return nil, ErrAccessDenied
	}

	// Set up the query's permissions
	if dbid != "heedy" {
		i := strings.Index(dbid, "/")
		if i > -1 {
			usr := dbid[:i]
			conn := dbid[i+1:]
			if o.User != nil || o.App != nil && *o.App != conn {
				return nil, ErrAccessDenied
			}
			o.User = &usr
			o.App = &conn
		} else {
			if o.User != nil && *o.User != dbid {
				return nil, ErrAccessDenied
			}
			o.User = &dbid
		}
	}
	return o, nil
}

func extractQueryBasics(o *NotificationsQuery) ([]string, []interface{}) {
	cNames := []string{}
	cValues := []interface{}{}
	if o.Type != nil {
		cNames = append(cNames, "type")
		cValues = append(cValues, *o.Type)
	}
	if o.Seen != nil {
		cNames = append(cNames, "seen")
		cValues = append(cValues, *o.Seen)
	}
	if o.Global != nil {
		cNames = append(cNames, "global")
		cValues = append(cValues, *o.Global)
	}
	if o.Key != nil {
		cNames = append(cNames, "key")
		cValues = append(cValues, *o.Key)
	}
	return cNames, cValues
}

func includeTable(o *NotificationsQuery) (bool, bool, bool) {
	if o == nil {
		o = &NotificationsQuery{}
	}
	if o.IncludeSelf == nil {
		includeSelfDefault := false
		o.IncludeSelf = &includeSelfDefault
	}

	includeUser := false
	includeApp := false
	includeSource := false
	if o.User == nil && o.App == nil && o.Source == nil {
		includeUser = true
		includeApp = true
		includeSource = true
	} else {
		if o.Source != nil {
			includeSource = true
		}
		if o.App != nil && (*o.App == "*" || o.Source == nil || *o.IncludeSelf || *o.Source != "*") {
			includeApp = true
		}
		if o.User != nil && (*o.User == "*" || (o.Source == nil && o.App == nil) || *o.IncludeSelf || o.Source != nil && *o.Source != "*" || o.App != nil && *o.App != "*") {
			includeUser = true
		}
	}
	return includeUser, includeApp, includeSource
}

// ReadNotifications reads the notifications associated with the given user/app/source
func ReadNotifications(db database.DB, o *NotificationsQuery) ([]Notification, error) {
	// Figure out which tables to query for the results
	includeUser, includeApp, includeSource := includeTable(o)

	o, err := queryAllowed(db, o)
	if err != nil {
		return nil, err
	}

	res := []Notification{}

	// Set up the query that will be used to filter results
	cNames, cValues := extractQueryBasics(o)

	if o.User != nil && *o.User != "*" {
		cNames = append(cNames, "user")
		cValues = append(cValues, *o.User)
	}

	if includeUser {
		queryWhere := strings.Join(cNames, "=? AND ") + "=?"
		var r []Notification
		err := db.AdminDB().Select(&r, fmt.Sprintf("SELECT * FROM notifications_user WHERE %s;", queryWhere), cValues...)
		if err != nil {
			return nil, err
		}
		res = append(res, r...)
	}

	if o.App != nil && *o.App != "*" {
		cNames = append(cNames, "app")
		cValues = append(cValues, *o.App)
	}

	if includeApp {
		queryWhere := strings.Join(cNames, "=? AND ") + "=?"
		var r []Notification
		err := db.AdminDB().Select(&r, fmt.Sprintf("SELECT * FROM notifications_app WHERE %s;", queryWhere), cValues...)
		if err != nil {
			return nil, err
		}
		res = append(res, r...)
	}

	if o.Source != nil && *o.Source != "*" {
		cNames = append(cNames, "source")
		cValues = append(cValues, *o.Source)
	}

	if includeSource {
		queryWhere := strings.Join(cNames, "=? AND ") + "=?"
		var r []Notification
		err := db.AdminDB().Select(&r, fmt.Sprintf("SELECT * FROM notifications_source WHERE %s;", queryWhere), cValues...)
		if err != nil {
			return nil, err
		}
		res = append(res, r...)
	}

	return res, nil
}

func extractNotificationBasics(n *Notification) ([]string, []interface{}) {
	cNames := []string{}
	cValues := []interface{}{}

	if n.Key != "" {
		cNames = append(cNames, "key")
		cValues = append(cValues, n.Key)
	}

	if n.Title != nil {
		cNames = append(cNames, "title")
		cValues = append(cValues, n.Title)
	}
	if n.Description != nil {
		cNames = append(cNames, "description")
		cValues = append(cValues, *n.Description)
	}
	if n.Seen != nil {
		cNames = append(cNames, "seen")
		cValues = append(cValues, *n.Seen)
	}
	if n.Global != nil {
		cNames = append(cNames, "global")
		cValues = append(cValues, *n.Global)
	}
	if n.Type != nil {
		cNames = append(cNames, "type")
		cValues = append(cValues, *n.Type)
	}
	return cNames, cValues
}

func excludeStmt(cNames []string) string {
	narr := make([]string, len(cNames))
	for i := range cNames {
		narr[i] = fmt.Sprintf("%s=excluded.%s", cNames[i], cNames[i])
	}
	return strings.Join(narr, ", ")
}

// WriteNotification writes the given notification. If a notification with the given key and target exists, it updates the existing notification with the new
// values. The notification will only update those values that are specifically set in the new notification
func WriteNotification(db database.DB, n *Notification) error {
	dbid := db.ID()
	if n.Key == "" || n.Title == nil || *n.Title == "" {
		return errors.New("bad_request: Notifications must have a valid key and title")
	}
	if n.Timestamp != 0 {
		return errors.New("bad_request: timestamps are set automatically")
	}
	if n.User == nil && n.App == nil && n.Source == nil && dbid != "heedy" {
		// The notification is to be inserted to itself
		i := strings.Index(dbid, "/")
		if i > -1 {
			conn := dbid[i+1:]
			n.App = &conn
		} else {
			n.User = &dbid
		}
	}

	// Set up the columns that will be set on the notification
	cNames, cValues := extractNotificationBasics(n)
	cNames = append(cNames, "timestamp")
	cValues = append(cValues, float64(time.Now().UnixNano())*1e-9)
	eS := excludeStmt(cNames)

	if n.Source != nil {
		// The notification is for a source
		s, err := db.ReadSource(*n.Source, nil)
		if err != nil {
			return err
		}
		if dbid == "heedy" || dbid == *s.Owner && s.App == nil || s.App != nil && dbid == *s.Owner+"/"+*s.App {
			// Allow writing the notification
			n.User = s.Owner
			n.App = s.App
			cNames = append(cNames, "user", "app", "source")
			cValues = append(cValues, *s.Owner, *s.App, s.ID)
			_, err := db.AdminDB().Exec(fmt.Sprintf("INSERT INTO notifications_source(%s) VALUES (%s) ON CONFLICT(user,key) DO UPDATE SET %s;", strings.Join(cNames, ","), database.QQ(len(cNames)), eS),
				cValues...)
			return err
		}
		return database.ErrAccessDenied("Can't set notifications for this source")
	}
	if n.App != nil {
		c, err := db.ReadApp(*n.App, nil)
		if err != nil {
			return err
		}

		if dbid == "heedy" || dbid == *c.Owner+"/"+c.ID {
			// Allow writing the notification
			n.User = c.Owner
			n.App = &c.ID
			cNames = append(cNames, "user", "app")
			cValues = append(cValues, *c.Owner, c.ID)
			_, err := db.AdminDB().Exec(fmt.Sprintf("INSERT INTO notifications_app(%s) VALUES (%s) ON CONFLICT(user,app,key) DO UPDATE SET %s;", strings.Join(cNames, ","), database.QQ(len(cNames)), eS),
				cValues...)
			return err
		}
		return database.ErrAccessDenied("Can't set notifications for this app")
	}
	if n.User == nil {
		return errors.New("Must specify a target for the notification")
	}
	u, err := db.ReadUser(*n.User, nil)
	if err != nil {
		return err
	}
	if dbid == "heedy" || *u.UserName == dbid {
		cNames = append(cNames, "user")
		cValues = append(cValues, *u.UserName)
		_, err := db.AdminDB().Exec(fmt.Sprintf("INSERT INTO notifications_user(%s) VALUES (%s) ON CONFLICT(user,app,source,key) DO UPDATE SET %s;", strings.Join(cNames, ","), database.QQ(len(cNames)), eS),
			cValues...)
		return err
	}
	return database.ErrAccessDenied("Can't set notifications for this user")
}

// UpdateNotification is a special version that modifies all notifications satisfying the constraints given in NotificationsQuery
func UpdateNotification(db database.DB, n *Notification, o *NotificationsQuery) error {
	includeUser, includeApp, includeSource := includeTable(o)
	if n.Timestamp != 0 {
		return errors.New("bad_request: timestamps are set automatically")
	}

	o, err := queryAllowed(db, o)
	if err != nil {
		return err
	}

	ncNames, ncValues := extractNotificationBasics(n)
	ocNames, ocValues := extractQueryBasics(o)

	queryUpdate := strings.Join(ncNames, "=?, ") + "=?"

	if o.User != nil && *o.User != "*" {
		ocNames = append(ocNames, "user")
		ocValues = append(ocValues, *o.User)
	}

	if includeUser {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("UPDATE notifications_user SET %s WHERE %s", queryUpdate, queryWhere)
		vals := append(append([]interface{}{}, ncValues...), ocValues...)
		_, err := db.AdminDB().Exec(qstring, vals...)
		if err != nil {
			return err
		}
	}

	if o.App != nil && *o.App != "*" {
		ocNames = append(ocNames, "app")
		ocValues = append(ocValues, *o.App)
	}

	if includeApp {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("UPDATE notifications_app SET %s WHERE %s", queryUpdate, queryWhere)
		vals := append(append([]interface{}{}, ncValues...), ocValues...)
		_, err := db.AdminDB().Exec(qstring, vals...)
		if err != nil {
			return err
		}
	}

	if o.Source != nil && *o.Source != "*" {
		ocNames = append(ocNames, "source")
		ocValues = append(ocValues, *o.Source)
	}

	if includeSource {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("UPDATE notifications_source SET %s WHERE %s", queryUpdate, queryWhere)
		vals := append(append([]interface{}{}, ncValues...), ocValues...)
		_, err := db.AdminDB().Exec(qstring, vals...)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteNotification takes a queryer for notifications
func DeleteNotification(db database.DB, o *NotificationsQuery) error {
	includeUser, includeApp, includeSource := includeTable(o)
	o, err := queryAllowed(db, o)
	if err != nil {
		return err
	}

	ocNames, ocValues := extractQueryBasics(o)

	if o.User != nil && *o.User != "*" {
		ocNames = append(ocNames, "user")
		ocValues = append(ocValues, *o.User)
	}

	if includeUser {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("DELETE FROM notifications_user WHERE %s", queryWhere)
		_, err := db.AdminDB().Exec(qstring, ocValues...)
		if err != nil {
			return err
		}
	}

	if o.App != nil && *o.App != "*" {
		ocNames = append(ocNames, "app")
		ocValues = append(ocValues, *o.App)
	}

	if includeApp {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("DELETE FROM notifications_app WHERE %s", queryWhere)
		_, err := db.AdminDB().Exec(qstring, ocValues...)
		if err != nil {
			return err
		}
	}

	if o.Source != nil && *o.Source != "*" {
		ocNames = append(ocNames, "source")
		ocValues = append(ocValues, *o.Source)
	}

	if includeSource {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("DELETE FROM notifications_source WHERE %s", queryWhere)
		_, err := db.AdminDB().Exec(qstring, ocValues...)
		if err != nil {
			return err
		}
	}

	return nil

}
