package plugin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/schema"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
)

var ErrUnimplemented = errors.New("unimplemented")
var queryEncoder = schema.NewEncoder()

type PluginDB struct {
	P       *Plugin
	Entity  string
	Overlay int

	RequestID string

	client http.Client
}

func (db *PluginDB) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	host := db.P.Meta.Config.GetHost()
	if host == "" {
		host = "localhost"
	}
	host = "http://" + host + ":" + strconv.Itoa(int(db.P.Meta.Config.GetPort())) + path

	r, err := http.NewRequest(method, host, body)
	if err == nil {
		r.Header.Add("X-Heedy-Auth", db.Entity)
		r.Header.Add("X-Heedy-Key", db.P.Meta.APIKey)
		r.Header.Add("X-Heedy-Overlay", strconv.Itoa(db.Overlay))
		if db.RequestID != "" {
			r.Header.Add("X-Heedy-ID", db.RequestID)
		}

	}
	return r, err
}

// BasicRequest runs a basic query, and does not return the body unless there was an error
func (db *PluginDB) BasicRequest(method, api string, body io.Reader) error {
	r, err := db.NewRequest(method, api, body)
	if err != nil {
		return err
	}
	resp, err := db.client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 400 {
		return nil
	}

	// Error
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// The response is an error, so unmarshal into the error struct
	var eresp rest.ErrorResponse
	err = json.Unmarshal(b, &eresp)
	if err != nil {
		return err
	}
	return &eresp

}

func (db *PluginDB) UnmarshalRequest(obj interface{}, method, api string, body io.Reader) error {
	r, err := db.NewRequest(method, api, body)
	if err != nil {
		return err
	}
	resp, err := db.client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		// The response is an error, so unmarshal into the error struct
		var eresp rest.ErrorResponse
		err = json.Unmarshal(b, &eresp)
		if err != nil {
			return err
		}
		return &eresp
	}

	// Unmarshal the result
	return json.Unmarshal(b, obj)
}

func (db *PluginDB) StringRequest(method, api string, body io.Reader) (string, error) {
	r, err := db.NewRequest(method, api, body)
	if err != nil {
		return "", err
	}
	resp, err := db.client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		// The response is an error, so unmarshal into the error struct
		var eresp rest.ErrorResponse
		err = json.Unmarshal(b, &eresp)
		if err != nil {
			return "", err
		}
		return "", &eresp
	}
	return string(b), nil
}

// Fire allows PluginDB to conform to the events.Handler interface, which is used to fire events
func (db *PluginDB) Fire(e *events.Event) {
	api := "/api/heedy/v1/events"
	b, err := json.Marshal(e)
	if err != nil {
		db.P.Logger().Warnf("Failed to fire event: %s", err.Error())
		return
	}

	err = db.BasicRequest("POST", api, bytes.NewBuffer(b))
	if err != nil {
		db.P.Logger().Warnf("Failed to fire event: %s", err.Error())
	}
}

func (db *PluginDB) AdminDB() *database.AdminDB {
	adb, err := db.P.AdminDB()
	if err != nil {
		db.P.Logger().Errorf("Could not open AdminDB: %s", err.Error())
		return nil
	}
	return adb
}

func (db *PluginDB) ID() string {
	return db.Entity
}

func (db *PluginDB) Type() database.DBType {
	if db.Entity == "heedy" {
		return database.AdminType
	}
	if db.Entity == "public" {
		return database.PublicType
	}
	i := strings.Index(db.Entity, "/")
	if i > -1 {
		return database.UserType
	}
	return database.AppType
}

func (db *PluginDB) CreateUser(u *database.User) error {
	return database.ErrBadQuery("Can't create users through the REST API")
}

func (db *PluginDB) ReadUser(name string, o *database.ReadUserOptions) (*database.User, error) {
	api := fmt.Sprintf("/api/heedy/v1/users/%s", name)

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	var u database.User

	err := db.UnmarshalRequest(&u, "GET", api, nil)
	return &u, err
}
func (db *PluginDB) UpdateUser(u *database.User) error {
	api := fmt.Sprintf("/api/heedy/v1/users/%s", u.ID)
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return db.BasicRequest("PATCH", api, bytes.NewBuffer(b))
}
func (db *PluginDB) DelUser(name string) error {
	api := fmt.Sprintf("/api/heedy/v1/users/%s", name)
	return db.BasicRequest("DELETE", api, nil)
}

func (db *PluginDB) ListUsers(o *database.ListUsersOptions) ([]*database.User, error) {
	var sl []*database.User
	api := "/api/heedy/v1/users"

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	err := db.UnmarshalRequest(&sl, "GET", api, nil)
	return sl, err
}

func (db *PluginDB) CanCreateSource(s *database.Source) error {
	return ErrUnimplemented
}
func (db *PluginDB) CreateSource(s *database.Source) (string, error) {
	api := "/api/heedy/v1/sources"
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	err = db.UnmarshalRequest(&s, "POST", api, bytes.NewBuffer(b))
	return s.ID, err
}
func (db *PluginDB) ReadSource(id string, o *database.ReadSourceOptions) (*database.Source, error) {
	api := fmt.Sprintf("/api/heedy/v1/sources/%s", id)

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	var s database.Source

	err := db.UnmarshalRequest(&s, "GET", api, nil)
	return &s, err
}
func (db *PluginDB) UpdateSource(s *database.Source) error {
	api := fmt.Sprintf("/api/heedy/v1/sources/%s", s.ID)
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return db.BasicRequest("PATCH", api, bytes.NewBuffer(b))
}
func (db *PluginDB) DelSource(id string) error {
	api := fmt.Sprintf("/api/heedy/v1/sources/%s", id)
	return db.BasicRequest("DELETE", api, nil)
}

func (db *PluginDB) ShareSource(sourceid, userid string, sa *database.ScopeArray) error {
	return ErrUnimplemented
}
func (db *PluginDB) UnshareSourceFromUser(sourceid, userid string) error {
	return ErrUnimplemented
}
func (db *PluginDB) UnshareSource(sourceid string) error {
	return ErrUnimplemented
}
func (db *PluginDB) GetSourceShares(sourceid string) (m map[string]*database.ScopeArray, err error) {
	return nil, ErrUnimplemented
}

// ListSources lists the given sources
func (db *PluginDB) ListSources(o *database.ListSourcesOptions) ([]*database.Source, error) {
	var sl []*database.Source
	api := "/api/heedy/v1/sources"

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	err := db.UnmarshalRequest(&sl, "GET", api, nil)
	return sl, err
}

func (db *PluginDB) CreateApp(c *database.App) (string, string, error) {
	api := "/api/heedy/v1/apps"
	b, err := json.Marshal(c)
	if err != nil {
		return "", "", err
	}

	err = db.UnmarshalRequest(&c, "POST", api, bytes.NewBuffer(b))
	accessToken := ""
	if c.AccessToken != nil {
		accessToken = *c.AccessToken
	}
	return c.ID, accessToken, err
}
func (db *PluginDB) ReadApp(id string, o *database.ReadAppOptions) (*database.App, error) {
	api := fmt.Sprintf("/api/heedy/v1/apps/%s", id)

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	var c database.App

	err := db.UnmarshalRequest(&c, "GET", api, nil)
	return &c, err
}
func (db *PluginDB) UpdateApp(c *database.App) error {
	api := fmt.Sprintf("/api/heedy/v1/apps/%s", c.ID)
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return db.BasicRequest("PATCH", api, bytes.NewBuffer(b))
}
func (db *PluginDB) DelApp(id string) error {
	api := fmt.Sprintf("/api/heedy/v1/apps/%s", id)
	return db.BasicRequest("DELETE", api, nil)
}
func (db *PluginDB) ListApps(o *database.ListAppOptions) ([]*database.App, error) {
	var cl []*database.App
	api := "/api/heedy/v1/apps"

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	err := db.UnmarshalRequest(&cl, "GET", api, nil)
	return cl, err
}
