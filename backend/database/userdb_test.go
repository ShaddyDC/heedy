package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserUser(t *testing.T) {
	adb, cleanup := newDB(t)
	defer cleanup()

	// Create
	name := "testy"
	passwd := "testpass"
	require.NoError(t, adb.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	db := NewUserDB(adb, "testy")

	// Can't create the user
	require.Error(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	// Add user creation permission
	adb.AddUserScopeSet("testy", "testy")
	adb.AddScope("testy", "users:create")

	// Create
	name2 := "testy2"
	require.NoError(t, db.CreateUser(&User{
		Details: Details{
			Name: &name2,
		},
		Password: &passwd,
	}))

	_, err := db.ReadUser("notauser", nil)
	require.Error(t, err)

	u, err := db.ReadUser("testy", nil)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")

	_, err = db.ReadUser("testy2", nil)
	require.Error(t, err)

	// Make sure we can no longer read ourselves if we remove the wrong permission
	require.NoError(t, adb.RemScope("testy", "user:read"))
	require.NoError(t, adb.RemScope("users", "user:read"))

	_, err = db.ReadUser("testy", nil)
	require.Error(t, err)

	adb.AddScope("users", "users:read")

	u, err = db.ReadUser("testy", nil)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")

	u, err = db.ReadUser("testy2", nil)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy2")

	passwd = "mypass2"
	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy",
		},
		Password: &passwd,
	}))

	// Shouldn't be allowed to change another user's password without the scope present
	require.Error(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy2",
		},
		Password: &passwd,
	}))

	adb.AddScope("testy", "users:edit:password")

	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy2",
		},
		Password: &passwd,
	}))

	// And now try deleting the user
	require.Error(t, db.DelUser("testy2"))

	adb.AddScope("testy", "user:delete")

	require.Error(t, db.DelUser("testy2"))

	adb.AddScope("users", "users:delete")

	require.NoError(t, db.DelUser("testy2"))

	_, err = adb.ReadUser("testy2", nil)
	require.Error(t, err)

	require.NoError(t, db.DelUser("testy"))

	// And now comes the question of ensuring that the db object is no longer valid...
	// but a user only logs in from browser, so maybe can just manually check eveny couple minutes?

}

func TestUserUpdateAvatar(t *testing.T) {
	adb, cleanup := newDBWithUser(t)
	defer cleanup()

	db := NewUserDB(adb, "testy")
	avatar := "mi:lol"
	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID:     "testy",
			Avatar: &avatar,
		},
	}))
}

func TestUserScopes(t *testing.T) {
	adb, cleanup := newDB(t)
	defer cleanup()

	// Create
	name := "testy"
	passwd := "testpass"
	require.NoError(t, adb.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	name2 := "testy2"
	require.NoError(t, adb.CreateUser(&User{
		Details: Details{
			Name: &name2,
		},
		Password: &passwd,
	}))

	db := NewUserDB(adb, "testy")

	s, err := db.ReadUserScopes("testy")
	require.NoError(t, err)

	require.NoError(t, adb.AddUserScopeSet("testy", "testy"))
	require.NoError(t, adb.AddScope("users", "trunk"))
	require.NoError(t, adb.AddScope("testy", "retfdg"))

	s2, err := db.ReadUserScopes("testy")
	require.NoError(t, err)
	require.Equal(t, len(s)+2, len(s2))

	_, err = db.ReadUserScopes("testy2")
	require.Error(t, err)

	require.NoError(t, adb.AddScope("users", "users:scopes"))
	s, err = db.ReadUserScopes("testy2")
	require.NoError(t, err)

	require.NoError(t, adb.AddScope("users", "trueertwertnk"))
	require.NoError(t, adb.AddScope("testy2", "retfgshfdgaerdg"))
	require.NoError(t, adb.AddUserScopeSet("testy2", "testy2"))

	s2, err = db.ReadUserScopes("testy2")
	require.NoError(t, err)
	require.Equal(t, len(s)+2, len(s2))
}
