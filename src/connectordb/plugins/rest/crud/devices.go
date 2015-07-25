package crud

import (
	"connectordb/streamdb/operator"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"

	"connectordb/plugins/rest/restcore"
)

func getDevicePath(request *http.Request) (username string, devicename string, devicepath string) {
	username = mux.Vars(request)["user"]
	devicename = mux.Vars(request)["device"]
	devicepath = username + "/" + devicename
	return username, devicename, devicepath
}

//GetThis is a command to return the "username/devicename" of the currently authenticated thing
func GetThis(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(o.Name()))
	return nil
}

//ListDevices lists the devices that the given user has
func ListDevices(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	usrname := mux.Vars(request)["user"]
	d, err := o.ReadAllDevices(usrname)
	return restcore.JSONWriter(writer, d, logger, err)
}

//CreateDevice creates a new user from a REST API request
func CreateDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, devname, devpath := getDevicePath(request)
	err := restcore.ValidName(devname, nil)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return err
	}

	if err = o.CreateDevice(devpath); err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}

	return ReadDevice(o, writer, request, logger)
}

//ReadDevice gets an existing device from a REST API request
func ReadDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, devpath := getDevicePath(request)

	if err := restcore.BadQ(o, writer, request, logger); err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return err
	}
	d, err := o.ReadDevice(devpath)
	return restcore.JSONWriter(writer, d, logger, err)
}

//UpdateDevice updates the metadata for existing device from a REST API request
func UpdateDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, devpath := getDevicePath(request)

	d, err := o.ReadDevice(devpath)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}

	err = restcore.UnmarshalRequest(request, d)
	err = restcore.ValidName(d.Name, err)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return err
	}

	if d.ApiKey == "" {
		//The user wants to reset the API key
		newkey, err := uuid.NewV4()
		if err != nil {
			restcore.WriteError(writer, logger, http.StatusInternalServerError, err, false)
			return err
		}
		d.ApiKey = newkey.String()
	}

	if err = o.UpdateDevice(d); err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}
	return restcore.JSONWriter(writer, d, logger, err)
}

//DeleteDevice deletes existing device from a REST API request
func DeleteDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, devpath := getDevicePath(request)
	err := o.DeleteDevice(devpath)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}
	return restcore.OK(writer)
}
