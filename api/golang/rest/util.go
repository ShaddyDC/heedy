package rest

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"encoding/json"

	"net/http"

	"github.com/gorilla/schema"
	"github.com/heedy/heedy/backend/assets"
	"github.com/sirupsen/logrus"
)

var QueryDecoder = schema.NewDecoder()
var ErrNotFound = errors.New("not_found: The given endpoint is not available")

// apiHeaders writes headers that need to be present in all API requests
func apiHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // All API requests return json
}

// RequestLogger generates a basic logger that holds relevant request info
func RequestLogger(r *http.Request) *logrus.Entry {
	raddr := r.RemoteAddr
	if fwdFor := r.Header.Get("X-Forwarded-For"); fwdFor != "" {
		raddr = fwdFor
	}
	fields := logrus.Fields{"addr": raddr, "path": r.URL.Path, "method": r.Method}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		fields["realip"] = realIP
	}
	return logrus.WithFields(fields)
}

//UnmarshalRequest unmarshals the input data to the given interface
func UnmarshalRequest(request *http.Request, unmarshalTo interface{}) error {
	defer request.Body.Close()

	//Limit requests to the limit given in configuration
	data, err := ioutil.ReadAll(io.LimitReader(request.Body, *assets.Config().RequestBodyByteLimit))
	if err != nil {
		return err
	}

	return json.Unmarshal(data, unmarshalTo)
}

// ErrorResponse is the response given by the server upon an error
type ErrorResponse struct {
	ErrorName        string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ID               string `json:"id,omitempty"`
}

func (er *ErrorResponse) Error() string {
	return er.ErrorName + ":" + er.ErrorDescription
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSONError(w, r, http.StatusNotFound, ErrNotFound)
}

// WriteJSONError writes an error message as json. It is assumed that the resulting
// status code is not StatusOK, but rather 4xx
func WriteJSONError(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-cache")
	c := CTX(r)

	es := ErrorResponse{
		ErrorName:        "internal_error",
		ErrorDescription: err.Error(),
	}
	myerr := err

	// We can have error types encoded in the error, split with a :
	errs := strings.SplitN(err.Error(), ":", 2)
	if len(errs) > 1 && !strings.Contains(errs[0], " ") {
		es.ErrorName = errs[0]
		es.ErrorDescription = strings.TrimSpace(errs[1])
	}

	if c != nil {
		es.ID = c.RequestID
	}
	jes, err := json.Marshal(&es)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server_error", "error_description": "Failed to create error message"}`))
		if c != nil {
			c.Log.Errorf("Failed to write error message: %s", err)
		} else {
			logrus.Errorf("Failed to write error message: %s", err)
		}
		return
	}

	if c != nil {
		c.Log.Warn(myerr)
	} else {
		logrus.Warn(myerr)
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(jes)))
	w.WriteHeader(status)
	w.Write(jes)
}

// WriteJSON writes response as JSON, or writes the error if such is given
func WriteJSON(w http.ResponseWriter, r *http.Request, data interface{}, err error) {
	if err != nil {
		// By default, an error returns 400
		WriteJSONError(w, r, 400, err)
		return
	}
	jdata, err := json.Marshal(data)
	if err != nil {
		WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}

	// golang json unmarshal encodes empty arrays as null
	// https://github.com/golang/go/issues/27589
	// This detects that, and fixes the problem.
	if bytes.Equal(jdata, []byte("null")) && data != nil {
		if k := reflect.TypeOf(data).Kind(); (k == reflect.Slice || k == reflect.Array) && reflect.ValueOf(data).Len() == 0 {
			jdata = []byte("[]")
		}
	}
	w.Header().Set("Cache-Control", "private, no-cache")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(jdata)))
	w.WriteHeader(http.StatusOK)
	w.Write(jdata)
}

// WriteResult writes empty object if the command succeeded, and outputs an error if it didn't
func WriteResult(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		// By default, an error returns 400
		WriteJSONError(w, r, 400, err)
		return
	}
	// success :)
	w.Header().Set("Cache-Control", "private, no-cache")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", "15")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"result":"ok"}`))

}

// WriteGZIP gzips a response Reader object if gzip is an accepted encoding. While it can be a security risk
// is some cases, it is very useful when the response can be enormous (like timeseries data).
func WriteGZIP(w http.ResponseWriter, r *http.Request, towrite io.Reader, status int) error {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.WriteHeader(status)
		_, err := io.Copy(w, towrite)
		return err
	}
	w.Header().Set("Content-Encoding", "gzip")
	w.WriteHeader(status)
	g := gzip.NewWriter(w)
	_, err := io.Copy(g, towrite)
	if err != nil {
		g.Close()
		return err
	}
	return g.Close()

}
