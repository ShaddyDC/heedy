package authoperator

import (
	"connectordb/authoperator/permissions"
	"connectordb/operator"
	"connectordb/pathwrapper"
	"connectordb/users"

	pconfig "config/permissions"
)

// AuthOperator is the operator which represents actions as
// a particular logged in device
type AuthOperator struct {
	Operator operator.PathOperator
	pathwrapper.Wrapper

	devicePath string // The string name of this operator
	deviceID   int64  // The ID of this device
}

// NewAuthOperator creates a new authentication operator based upon the given DeviceID
func NewAuthOperator(op operator.PathOperator, deviceID int64) (*AuthOperator, error) {
	dev, err := op.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	usr, err := op.ReadUserByID(dev.UserID)
	if err != nil {
		return nil, err
	}

	ao := &AuthOperator{op, pathwrapper.Wrapper{}, usr.Name + "/" + dev.Name, deviceID}
	ao.Wrapper = pathwrapper.Wrap(ao)
	return ao, nil
}

// NewNobody logs in as a "nobody"
func NewNobody(op operator.PathOperator) *AuthOperator {
	ao := &AuthOperator{op, pathwrapper.Wrapper{}, "nobody", -2}
	ao.Wrapper = pathwrapper.Wrap(ao)
	return ao
}

// Name is the path to the device underlying the operator
func (a *AuthOperator) Name() string {
	return a.devicePath
}

// User returns the current user (ie, user that is logged in).
// No permissions checking is done
func (a *AuthOperator) User() (usr *users.User, err error) {
	if a.deviceID == -2 {
		// Nobody has deviceID -2
		return &users.User{
			UserID: -2,
			Name:   "nobody",
			Role:   "nobody",
		}, nil
	}
	dev, err := a.Operator.ReadDeviceByID(a.deviceID)
	if err != nil {
		return nil, err
	}
	return a.Operator.ReadUserByID(dev.UserID)
}

// Device returns the current device. No permissions checking
// is done on the device
func (a *AuthOperator) Device() (*users.Device, error) {
	if a.deviceID == -2 {
		// The nobody operator has deviceID = -2
		return &users.Device{
			DeviceID: -2,
			UserID:   -2,
			Name:     "none",
			Role:     "user",
		}, nil
	}
	return a.Operator.ReadDeviceByID(a.deviceID)
}

// AdminOperator returns the administrative operator
func (a *AuthOperator) AdminOperator() operator.PathOperator {
	return a.Operator.AdminOperator()
}

// getUserAndDevice returns both the current user AND the current device
// it is just there to simplify our work
func (a *AuthOperator) getUserAndDevice() (*users.User, *users.Device, error) {
	dev, err := a.Device()
	if err != nil {
		return nil, nil, err
	}
	u, err := a.User()
	return u, dev, err
}

// getAccessLevels gets the access levels for the current user/device combo
func (a *AuthOperator) getAccessLevels(userID int64, ispublic, issself bool) (*pconfig.Permissions, *users.User, *users.Device, *pconfig.AccessLevel, *pconfig.AccessLevel, error) {
	u, d, err := a.getUserAndDevice()
	if err != nil {
		return nil, nil, nil, nil, nil, permissions.ErrNoAccess
	}
	perm := pconfig.Get()

	up, dp := permissions.GetAccessLevels(perm, u, d, userID, ispublic, issself)
	return perm, u, d, up, dp, nil
}

// getDeviceAccessLevels is same as getAccessLevels, but it is given a deviceID
func (a *AuthOperator) getDeviceAccessLevels(deviceID int64) (*pconfig.Permissions, *users.Device, *users.User, *users.Device, *pconfig.AccessLevel, *pconfig.AccessLevel, error) {
	selfuser, selfdevice, err := a.getUserAndDevice()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	dev, err := a.Operator.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, permissions.ErrNoAccess
	}

	perm := pconfig.Get()
	up, dp := permissions.GetAccessLevels(perm, selfuser, selfdevice, dev.UserID, dev.Public, selfdevice.DeviceID == dev.DeviceID)

	return perm, dev, selfuser, selfdevice, up, dp, nil
}
