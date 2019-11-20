package database


import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppObject(t *testing.T) {
	adb, cleanup := newDBWithUser(t)
	defer cleanup()

	udb := NewUserDB(adb, "testy")

	cname := "conn"
	cid,_, err := udb.CreateApp(&App{
		Details: Details{
			ID: "conn",
			Name: &cname,
		},
		Scopes: &AppScopeArray{
			ScopeArray: ScopeArray{
				Scopes: []string{"self.objects.stream","owner:read"},
			},
		},
	})
	require.NoError(t,err)
	c,err := udb.ReadApp(cid,nil)
	require.NoError(t,err)
	cdb := NewAppDB(adb,c)

	name := "tree"
	stype := "stream"
	sid, err := cdb.CreateObject(&Object{
		Details: Details{
			Name: &name,
		},
		Type: &stype,
	})
	require.NoError(t, err)

	name2 := "derpy"
	require.NoError(t, cdb.UpdateObject(&Object{
		Details: Details{
			ID:       sid,
			Name: &name2,
		},
		Meta: &JSONObject{
			"schema": 4,
		},
	}))

	s, err := cdb.ReadObject(sid, nil)
	require.NoError(t, err)
	require.Equal(t, *s.Name, name2)
	require.NotNil(t, s.Scopes)
	require.NotNil(t, s.Meta)
	require.True(t, s.Access.HasScope("*"))

	require.NoError(t, cdb.DelObject(sid))
	require.Error(t, cdb.DelObject(sid))
}