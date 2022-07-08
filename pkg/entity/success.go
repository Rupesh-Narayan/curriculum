package entity

import (
	"encoding/json"
	"net/http"
)

// Response struct
type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Meta   interface{} `json:"meta,omitempty"`
}

// SendResponse send http response
func (entity Response) SendResponse(rw http.ResponseWriter, data interface{}, meta interface{}, statusCode int) (err error) {
	rw.Header().Set("Content-Type", "application/json")
	entity.Data = data
	entity.Meta = meta
	switch statusCode {
	case http.StatusCreated:
		rw.WriteHeader(http.StatusCreated)
		entity.Status = http.StatusText(http.StatusCreated)
	case http.StatusAccepted:
		rw.WriteHeader(http.StatusAccepted)
		entity.Status = http.StatusText(http.StatusAccepted)
	default:
		rw.WriteHeader(http.StatusOK)
		entity.Status = http.StatusText(http.StatusOK)
	}

	// send response
	err = json.NewEncoder(rw).Encode(entity)
	if err != nil {
		return
	}
	return
}
