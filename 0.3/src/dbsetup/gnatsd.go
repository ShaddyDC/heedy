/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbsetup

import (
	"config"
	"util"
)

//GnatsdService is a service for running Redis
type GnatsdService struct {
	BaseService
}

//Start starts the service
func (s *GnatsdService) Start() error {
	configfile, err := s.start()
	if err != nil {
		return err
	}

	_, err = util.RunDaemon(err, GetExecutablePath("gnatsd"), "-c", configfile)
	err = util.WaitPort(s.S.Hostname, int(s.S.Port), err)

	return err
}

//NewGnatsdService creates a new service for gNatsd
func NewGnatsdService(serviceDirectory string, c *config.Configuration) *GnatsdService {
	return &GnatsdService{BaseService{serviceDirectory, "gnatsd", &c.Nats, c}}
}
