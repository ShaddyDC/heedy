/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbsetup

import (
	"config"
	"runtime"
	"syscall"
	"util"

	log "github.com/Sirupsen/logrus"
)

//Status represents the status of the service
type Status int

const (
	// When there is no status to speak of
	StatusNone = iota
	// The service is not running
	StatusStopped = iota
	// The service is running
	StatusRunning = iota
	// When the service encountered an error
	StatusError = iota
)

type Service interface {
	//Creates the service. This may include:
	// - Creating configuration files
	// - Creating directories, setting up databases
	Create() error

	//Starts the service from existing config
	// - Make sure all necessary files exist, load configs
	// - starts the service
	Start() error

	//Stops the service, closing all connections
	//Running stop on a nonexisting process has undefined behavior
	Stop() error

	//Immediately kills the process
	Kill() error

	//Name returns the name of the service
	Name() string

	//Status of the service
	Status() Status
}

//BaseService allows a few functions of a Service to be automatically implemented
type BaseService struct {
	ServiceDirectory string
	ServiceName      string
	S                *config.Service
	C                *config.Configuration
}

//Name returns the name of the service
func (bs BaseService) Name() string {
	return bs.ServiceName
}

//Status returns the status of the service
func (bs BaseService) Status() Status {
	_, err := util.GetProcess(bs.ServiceDirectory, bs.ServiceName, nil)
	if err != nil {
		return StatusStopped
	}
	return StatusRunning
}

func (bs BaseService) start() (string, error) {
	if bs.Status() == StatusRunning {
		return "", ErrAlreadyRunning
	}
	log.Infof("Staring %s on port %d", bs.Name(), bs.S.Port)

	if bs.S.Hostname == "" {
		bs.S.Hostname = "localhost"
	}

	configReplacements := GenerateConfigReplacements(bs.ServiceDirectory, bs.Name(), bs.C)
	return SetConfig(bs.ServiceDirectory, bs.Name()+".conf", configReplacements, nil)
}

//Create runs the base creation code - it copies the necessary configuration files
func (bs BaseService) Create() error {
	log.Infof("Setting up %s server", bs.Name())

	return CopyConfig(bs.ServiceDirectory, bs.Name()+".conf", nil)
}

//Stop shuts down a service
func (bs BaseService) Stop() error {
	log.Infof("Stopping %s...", bs.Name())

	p, err := util.GetProcess(bs.ServiceDirectory, bs.ServiceName, nil)
	if err != nil {
		return err
	}

	log.Debugf("%s has pid %d", bs.Name(), p.Pid)

	if runtime.GOOS == "windows" {
		// Why does windows not have a nicer signal to stop a process?
		p.Kill()
	} else {
		// On linux, sigterm is used
		if err := p.Signal(syscall.SIGTERM); err != nil {
			return err
		}
	}

	p.Wait()

	return nil
}

// Kill murders a process in cold blood
func (bs *BaseService) Kill() error {
	log.Warnf("Killing %s server", bs.Name())

	if bs.Status() != StatusRunning {
		return nil
	}

	p, err := util.GetProcess(bs.ServiceDirectory, bs.ServiceName, nil)
	if err != nil {
		return err
	}

	if err := p.Kill(); err != nil {
		return err
	}

	return nil
}
