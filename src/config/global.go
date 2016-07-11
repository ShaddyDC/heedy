/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package config

import (
	"path/filepath"
	"util"

	"config/permissions"

	log "github.com/Sirupsen/logrus"
)

var (

	// globalConfiguration is the configuration used throughout the system.
	// While the configuration can be reloaded during runtime, only certain properties are actually modifiable during runtime
	// and others fail to update silently. Just a warning.
	globalConfiguration *ConfigurationLoader
)

// Get returns the global configuration of the system
func Get() *Configuration {
	if globalConfiguration == nil {
		log.Warn("Global configuration not set - using default")
		return NewConfiguration()
	}
	return globalConfiguration.Get()
}

// LoadConfig just runs SetPath and returns the resulting config
func LoadConfig(location string) (*Configuration, error) {
	if err := SetPath(location); err != nil {
		return nil, err
	}
	return Get(), nil
}

// SetPath sets the global system configuration to the given file name, which will be watched for changes
func SetPath(filename string) error {
	//There are a few different situations that we handle here:
	//1) A database folder is given
	//		In this case we read the connectordb.conf file in the folder
	//2) A config file is given
	//		We read the file

	if util.IsDirectory(filename) {
		filename = filepath.Join(filename, "connectordb.conf")
	}

	cfg, err := NewConfigurationLoader(filename)
	if err != nil {
		return err
	}

	err = permissions.SetPath(cfg.Get().Permissions)
	if err != nil {
		return err
	}

	if globalConfiguration != nil {
		globalConfiguration.Close()
	}
	if globalConfiguration != nil {
		cfg.OnChange = globalConfiguration.OnChange
	}

	globalConfiguration = cfg

	// PipeScript has its own special configuration updater, which modifies the global
	// PipeScript configuration
	Get().PipeScript.Set()
	OnChangeCallback(func(c *Configuration) error {
		return c.PipeScript.Set()
	})

	// Next, we set up global logging. This is pretty convoluted,
	// but it turns out that due to some annoyance in putting together command
	// line logging parameters and these, logging is set up outside of this
	// method - but we still want logging to be set when the configuration changes
	OnChangeCallback(func(c *Configuration) error {
		return SetLoggingFromConfig(c)
	})

	if !Get().Watch {
		// Closing it will keep the config valid
		globalConfiguration.Close()
	}

	return nil
}

// OnChangeCallback adds a calback for modified configuration file
func OnChangeCallback(c ChangeCallback) {
	globalConfiguration.OnChangeCallback(c)
}

// ChangeCallback is a function that takes configuration, and returns an error
type ChangeCallback func(c *Configuration) error

// ConfigurationLoader watches a configuration for changes
type ConfigurationLoader struct {
	Config *Configuration // The currently loaded configuration

	OnChange []ChangeCallback // Callbacks that will be run on configuration file reload

	Watcher *util.FileWatcher // The file watcher that makes sure changes to config are reloaded
}

// Get returns the current configuration - the pointer is the thing that gets exchanged when a
// new configuration is loaded
func (c *ConfigurationLoader) Get() *Configuration {
	c.Watcher.RLock()
	defer c.Watcher.RUnlock()
	return c.Config
}

// NewConfigurationLoader returns a new watcher for the configuration
func NewConfigurationLoader(filename string) (*ConfigurationLoader, error) {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	c, err := Load(filename)
	if err != nil {
		return nil, err
	}
	cf := &ConfigurationLoader{
		Config:   c,
		OnChange: make([]ChangeCallback, 0, 5),
	}
	cf.Watcher, err = util.NewFileWatcher(filename, cf)

	return cf, err
}

// OnChangeCallback adds the given callback to the reload callback list
func (c *ConfigurationLoader) OnChangeCallback(cbk ChangeCallback) {
	c.OnChange = append(c.OnChange, cbk)
}

// Reload attempts to reload the configuration from the config file
func (c *ConfigurationLoader) Reload() error {
	cfg, err := Load(c.Watcher.FileName)
	if err != nil {
		return err
	}

	// Set the global permissions
	err = permissions.SetPath(cfg.Permissions)
	if err != nil {
		return err
	}

	if !cfg.Watch {
		c.Close() // This will still keep the permissions themselves in memory
	}

	c.Watcher.Lock()
	c.Config = cfg
	c.Watcher.Unlock()

	// Now run all callbacks - reload is guaranteed to be sequential, so no need
	// to lock config
	for i := range c.OnChange {
		if err = c.OnChange[i](c.Config); err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

// Close shuts down the configuration loader
func (c *ConfigurationLoader) Close() {
	c.Watcher.Close()
}
