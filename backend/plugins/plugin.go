package plugins

import (
	"fmt"
	"time"
	"errors"
	"path"
	"encoding/base64"
	"crypto/rand"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/assets"

	"github.com/robfig/cron"
	"github.com/go-chi/chi"

	"github.com/sirupsen/logrus"
)

type Plugin struct {
	// The mux that holds the overlay
	Mux *chi.Mux

	// The assets that are used by the server
	Assets *assets.Assets

	// The database
	DB *database.AdminDB

	// Name of the plugin
	Name string

	// Processes holds all the exec programs being handled by heedy.
	// The map keys are their associated api keys
	Processes map[string]*Exec `json:"exec"`

	// The cron daemon to run in the background for cron processes
	cron *cron.Cron
}

func NewPlugin(db *database.AdminDB,a *assets.Assets, pname string) (*Plugin,error) {
	p := &Plugin{
		DB: db,
		Processes: make(map[string]*Exec),
		Assets:    a,
		Name: pname,
	}
	logrus.Debugf("Loading plugin '%s'",pname)
	
	psettings := a.Config.Plugins[pname]

	if psettings.Routes != nil && len(*psettings.Routes) > 0 {

		mux := chi.NewMux()

		for rname, redirect := range *psettings.Routes {
			revproxy, err := NewReverseProxy(a.DataDir(), redirect)
			if err != nil {
				return nil, err
			}
			logrus.Debugf("%s: Forwarding %s -> %s ", pname, rname, redirect)
			mux.Handle(rname, revproxy)
		}

		p.Mux = mux
	}

	// Initialize the plugin
	return p,nil
}

// Start the backend executables
func (p *Plugin) Start() error {
	if len(p.Processes) > 0 {
		return errors.New("Must first stop running processes to restart Plugin")
	}
	p.cron = cron.New()
	p.cron.Start()
	pname := p.Name
	pv := p.Assets.Config.Plugins[pname]


	for ename, ev := range pv.Exec {
		if ev.Enabled == nil || ev.Enabled != nil && *ev.Enabled {
			keepAlive := false
			if ev.KeepAlive != nil {
				keepAlive = *ev.KeepAlive
			}
			if ev.Cmd == nil || len(*ev.Cmd) == 0 {
				p.Stop()
				return fmt.Errorf("%s/%s has empty command", pname, ename)
			}

			// Create an API key for the exec
			apikeybytes := make([]byte, 64)
			_, err := rand.Read(apikeybytes)

			e := &Exec{
				Plugin:    pname,
				Exec:      ename,
				APIKey:    base64.StdEncoding.EncodeToString(apikeybytes),
				Config:    p.Assets.Config,
				RootDir:   p.Assets.FolderPath,
				DataDir:   p.Assets.DataDir(),
				PluginDir: path.Join(p.Assets.PluginDir(), pname),
				keepAlive: keepAlive,
				cmd:       *ev.Cmd, // TODO: Handle Python
			}

			p.Processes[e.APIKey] = e

			if ev.Cron != nil && len(*ev.Cron) > 0 {
				logrus.Debugf("%s: Enabling cron job %s", pname, ename)
				err = p.cron.AddJob(*ev.Cron, e)
			} else {
				logrus.Debugf("%s: Running %s", pname, ename)
				err = e.Start()
			}
			if err != nil {
				p.Stop()
				return err
			}

		}

	}

	// Now wait until all the endpoints are open
	for ename,ev := range pv.Exec {
		if ev.Enabled == nil || ev.Enabled != nil && *ev.Enabled {
			if ev.Endpoint!=nil {
				logrus.Debugf("%s: Waiting for endpoint %s (%s)",pname,*ev.Endpoint,ename)
				method,host,err := GetEndpoint(p.Assets.DataDir(),*ev.Endpoint)
				if err!=nil {
					p.Stop()
					return err
				}
				if err = WaitForEndpoint(method,host); err!=nil {
					p.Stop()
					return err
				}
				logrus.Debugf("%s: Endpoint %s open",pname,*ev.Endpoint)
				
			}
		}
	}
	return nil
}

func processConnection(pluginKey string,owner string,cv *assets.Connection) *database.Connection {
	c := &database.Connection{
		Details: database.Details{
			Name: &cv.Name,
			Description: cv.Description,
			Avatar: cv.Avatar,
		},
		Enabled: cv.Enabled,
		Plugin: &pluginKey,
		Owner: &owner,
	}
	if cv.Scopes!=nil {
		c.Scopes = &database.ConnectionScopeArray{
			ScopeArray: database.ScopeArray{
				Scopes: *cv.Scopes,
			},
		}
	}
	if cv.AccessToken==nil || !(*cv.AccessToken) {
		empty := ""
		c.AccessToken = &empty
	}
	if cv.SettingSchema!=nil {
		jo := database.JSONObject(*cv.SettingSchema)
		c.SettingSchema = &jo
	}
	if cv.Settings!=nil {
		jo := database.JSONObject(*cv.Settings)
		c.Settings = &jo
	}
	return c
}

// BeforeStart is run before any of the plugin's executables are run.
// This function is used to check if we're to create connections/sources
// for the plugin
func (p *Plugin) BeforeStart() error {
	psettings := p.Assets.Config.Plugins[p.Name]
	for cname,cv := range psettings.Connections {
		// For each connection
		// Check if the connection exists for all users
		var res []string

		pluginKey := p.Name+":"+cname

		err := p.DB.DB.Select(&res,"SELECT username FROM users WHERE username NOT IN ('heedy', 'public', 'users') AND NOT EXISTS (SELECT 1 FROM connections WHERE owner=users.username AND connections.plugin=?);",pluginKey)
		if err!=nil {
			return err
		}
		if len(res) > 0 {
			logrus.Debugf("%s: Creating '%s' connection for all users",p.Name,pluginKey)

			// aaand how exactly do I achieve this?

			for _,uname := range res {
				
				_,_,err = p.DB.CreateConnection(processConnection(pluginKey,uname,cv))
				if err!=nil {
					return err
				}
			}
		}

		for skey,sv := range cv.Sources {
			res = []string{}
			err := p.DB.DB.Select(&res,"SELECT id FROM connections WHERE plugin=? AND NOT EXISTS (SELECT 1 FROM sources WHERE connection=connections.id AND key=?);",pluginKey,skey)
			if err!=nil {
				return err
			}
			if len(res) > 0 {
				logrus.Debugf("%s: Creating '%s/%s' source for all users",p.Name,pluginKey,skey)

				for _,cid := range res {
					logrus.Debug(cid,sv.Name)
				}
			}
		}
	}
	return nil
}

// AfterStart is used for the same purpose as BeforeStart, but it creates deferred sources/connections
func (p *Plugin) AfterStart() error {
	return nil
}


// GetProcessByKey gets the process associated with a given API key
func (p *Plugin) GetProcessByKey(key string) (*Exec,error) {
	v, ok := p.Processes[key]
	if ok {
		return v,nil
	}
	return nil,errors.New("No such key")
}

// Interrupt signals all processes to stop
func (p *Plugin) Interrupt() error {
	p.cron.Stop()
	for _,e := range p.Processes {
		e.Interrupt()
	}
	return nil
}

func (p *Plugin) AnyRunning() bool {
	anyrunning := false
	for _, e := range p.Processes {
		if e.IsRunning() {
			anyrunning = true
		}
	}
	
	return anyrunning
}

func (p *Plugin) HasProcess() bool {
	return len(p.Processes)==0
}

// Kill kills all processes
func (p *Plugin) Kill() {
	for _, e := range p.Processes {
		if e.IsRunning() {
			logrus.Warnf("%s: Killing %s", e.Exec)
			e.Kill()
		}
	}
}

func (p *Plugin) Stop() error {
	p.Interrupt()

	d := assets.Get().Config.GetExecTimeout()

	sleepDuration := 50 * time.Millisecond

	for i := time.Duration(0); i < d; i += sleepDuration {
		if !p.AnyRunning() {
			return nil
		}
		time.Sleep(sleepDuration)
	}

	p.Kill()
	return nil
}

func (p *Plugin) Close() error {
	if p.AnyRunning() {
		p.Stop()
	}
	return nil
}