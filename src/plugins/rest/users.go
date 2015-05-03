package rest

import (
	"net/http"
	"streamdb"

	"github.com/gorilla/mux"
)

//GetUser runs the GET operation routing for REST
func GetUser(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := mux.Vars(request)["user"]

	//there can be certain commands in place of a username - those represent invalid user names
	switch usrname {
	default:
		return ReadUser(o, writer, request)
	case "ls":
		return ListUsers(o, writer, request)
	case "this":
		return GetThis(o, writer, request)
	}

}

//ListUsers lists the users that the given operator can see
func ListUsers(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	u, err := o.ReadAllUsers()
	return JSONWriter(writer, u, err)
}

//CreateUser creates a new user from a REST API request
func CreateUser(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	//usrname := mux.Vars(request)["user"]
	writer.WriteHeader(http.StatusNotImplemented)

	return ErrUnderConstruction
}

//ReadUser reads the given user
func ReadUser(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := mux.Vars(request)["user"]
	u, err := o.ReadUser(usrname)
	return JSONWriter(writer, u, err)
}

//UpdateUser updates the metadata for existing user from a REST API request
func UpdateUser(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusNotImplemented)
	return ErrUnderConstruction
}

//DeleteUser deletes existing user from a REST API request
func DeleteUser(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusNotImplemented)
	return ErrUnderConstruction
}
