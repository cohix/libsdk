package resp

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// JSONOk marshals the obj to JSON and returns it to the caller with 200 OK.
func JSONOk(w http.ResponseWriter, obj any) error {
	return JSON(w, obj, http.StatusOK)
}

// JSON marshals the obj to JSON and returns it to the caller with the given status
func JSON(w http.ResponseWriter, obj any, status int) error {
	objJSON, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, "failed to json.Marshal")
	}

	w.WriteHeader(status)
	w.Write(objJSON)

	return nil
}
