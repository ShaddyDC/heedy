/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package plainoperator

import (
	"connectordb/datastream"
	"connectordb/operator/interfaces"
	"connectordb/operator/messenger"
	"connectordb/users"
	"errors"
)

const (
	PlainOperatorName = " ADMIN "
)

var (
	ErrAdmin = errors.New("An administrative operator has no user or device")
)

/**

The basic database access operator, overload anything here that you need to get
functionality right.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>

All Rights Reserved

**/

func NewPlainOperator(udb users.UserDatabase, ds *datastream.DataStream, msg *messenger.Messenger) PlainOperator {
	return PlainOperator{udb, ds, msg}
}

func NewWrappedPlainOperator(udb users.UserDatabase, ds *datastream.DataStream, msg *messenger.Messenger) interfaces.Operator {
	return &interfaces.PathOperatorMixin{&PlainOperator{udb, ds, msg}}
}

// This operator is very insecure but very fast, good for embedded environments
// where all you care about is speed and can trust all your users
type PlainOperator struct {
	Userdb users.UserDatabase     // SqlUserDatabase holds the methods needed to CRUD users/devices/streams
	ds     *datastream.DataStream // datastream holds methods for inserting datapoints into streams
	msg    *messenger.Messenger   // messenger is a connection to the messaging client
}

//Name here is a special one meaning that it is the database administration operator
// It is not a valid username
func (db *PlainOperator) Name() string {
	return PlainOperatorName
}

//User returns the current user
func (db *PlainOperator) User() (usr *users.User, err error) {
	return nil, ErrAdmin
}

func (db *PlainOperator) Device() (*users.Device, error) {
	return nil, ErrAdmin
}

func (o *PlainOperator) CreateUser(username, email, password string) error {
	return o.Userdb.CreateUser(username, email, password)
}

func (o *PlainOperator) ReadAllUsers() ([]users.User, error) {
	return o.Userdb.ReadAllUsers()
}

func (o *PlainOperator) ReadUser(username string) (*users.User, error) {
	return o.Userdb.ReadUserByName(username)
}

func (o *PlainOperator) ReadUserByID(userID int64) (*users.User, error) {
	return o.Userdb.ReadUserById(userID)
}

func (o *PlainOperator) UpdateUser(modifieduser *users.User) error {
	return o.Userdb.UpdateUser(modifieduser)
}

func (o *PlainOperator) Login(username, password string) (*users.User, *users.Device, error) {
	return o.Userdb.Login(username, password)
}

func (o *PlainOperator) DeleteUserByID(userID int64) error {
	// users are going to be GC'd from redis in the future - but we currently don't have that implemented,
	// so manually delete all the devices from redis if user delete succeeds
	dev, err := o.ReadAllDevicesByUserID(userID)
	if err != nil {
		return err
	}

	err = o.Userdb.DeleteUser(userID)

	if err == nil {
		for i := 0; i < len(dev); i++ {
			o.ds.DeleteDevice(dev[i].DeviceID)
		}
	}
	return err
}

func (o *PlainOperator) ReadAllDevicesByUserID(userID int64) ([]users.Device, error) {
	return o.Userdb.ReadDevicesForUserID(userID)
}

func (o *PlainOperator) CreateDeviceByUserID(userID int64, deviceName string) error {
	return o.Userdb.CreateDevice(deviceName, userID)
}

func (o *PlainOperator) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	return o.Userdb.ReadDeviceByID(deviceID)
}

func (o *PlainOperator) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	return o.Userdb.ReadDeviceForUserByName(userID, devicename)
}

func (o *PlainOperator) ReadDeviceByAPIKey(apikey string) (*users.Device, error) {
	return o.Userdb.ReadDeviceByAPIKey(apikey)
}

func (o *PlainOperator) UpdateDevice(modifieddevice *users.Device) error {
	return o.Userdb.UpdateDevice(modifieddevice)
}

func (o *PlainOperator) DeleteDeviceByID(deviceID int64) error {
	err := o.Userdb.DeleteDevice(deviceID)
	if err == nil {
		err = o.ds.DeleteDevice(deviceID)
	}
	return err
}

//UpdateStream updates the stream. BUG(daniel) the function currently does not give an error
//if someone attempts to update the schema (which is an illegal operation anyways)
func (o *PlainOperator) UpdateStream(modifiedstream *users.Stream) error {
	strm, err := o.ReadStreamByID(modifiedstream.StreamID)
	if err != nil {
		return err
	}

	err = o.Userdb.UpdateStream(modifiedstream)

	if err == nil && strm.Downlink == true && modifiedstream.Downlink == false {
		//There was a downlink here. Since the downlink was removed, we delete the associated
		//downlink substream
		o.DeleteStreamByID(strm.StreamID, "downlink")
	}

	return err
}
