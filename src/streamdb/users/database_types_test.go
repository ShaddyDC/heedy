// Package users provides an API for managing user information.
package users

import "testing"

func TestSetNewPassword(t *testing.T) {
	var j User
	var k User
	var l User

	j.SetNewPassword("monkey")
	k.SetNewPassword("password")
	l.PasswordSalt = "tmp"
	l.SetNewPassword("password")

	if j == k {
		t.Errorf("Setting password failed: %v vs %v", j, k)
		return
	}

	if k == l {
		t.Errorf("Salting Failed")
		return
	}

	j.SetNewPassword("password")

	if k != j {
		t.Errorf("Second Set Failed")
		return

	}
}

func TestAdmin(t *testing.T) {
	var j User
	j.Admin = true
	var k User
	k.Admin = false

	if k.IsAdmin() == true {
		t.Errorf("False positive admin")
		return
	}

	if j.IsAdmin() == false {
		t.Errorf("False negative admin")
		return
	}
}

func TestDevicePermissions(t *testing.T) {
	var all Device
	all.IsAdmin = true
	all.Enabled = true
	all.CanWrite = true
	all.CanWriteAnywhere = true

	var none Device

	var onlyEnabled Device
	onlyEnabled.Enabled = true

	var disabledSuper Device
	disabledSuper.IsAdmin = true

	if none.IsActive() {
		t.Errorf("improper active check.")
	}

	if !onlyEnabled.IsActive() {
		t.Errorf("improper active check.")
	}

	if onlyEnabled.IsAdmin {
		t.Errorf("improper elevation of privliges.")
	}

	if !all.IsAdmin {
		t.Errorf("Correct admin was denied")
	}

	// WriteAllowed

	if none.WriteAllowed() {
		t.Errorf("Granted write to unprivliged")
	}

	if !all.WriteAllowed() {
		t.Errorf("Denied write to privliged device")
	}

	// WriteAnywhereAllowed

	if none.WriteAnywhereAllowed() {
		t.Errorf("Granted WriteAnywhereAllowed to unprivliged")
	}

	if !all.WriteAnywhereAllowed() {
		t.Errorf("Denied WriteAnywhereAllowed to privliged device")
	}

	// CanModifyUser

	if none.CanActAsUser {
		t.Errorf("Granted CanModifyUser to unprivliged")
	}
}
