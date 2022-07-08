package entity

import (
	translation "bitbucket.org/noon-go/translator"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"encoding/json"
	"net/http"
)

// ErrorResponse response for error
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// HandleError handles error and send response
func HandleError(rw http.ResponseWriter, msg string, err error, locale string, translate bool) {
	// build default response
	var response *ErrorResponse
	response = &ErrorResponse{Message: "somethingWentWrong",
		Status: http.StatusText(http.StatusInternalServerError)}
	rw.Header().Set("Content-Type", "application/json")

	noonError, ok := err.(*noonerror.NoonError)
	if ok {
		err = noonError.Err
		msg = noonError.Message
	}
	// set header, message and status
	switch err {
	case noonerror.ErrUserNotFound:
		rw.WriteHeader(http.StatusNotFound)
		response.Message = msg
		response.Status = http.StatusText(http.StatusNotFound)
	case noonerror.ErrBadRequest, noonerror.ErrRelationExists, noonerror.ErrParamMissing, noonerror.ErrInvalidRequest:
		rw.WriteHeader(http.StatusBadRequest)
		response.Message = msg
		response.Status = http.StatusText(http.StatusBadRequest)
	case noonerror.ErrInternalServer:
		rw.WriteHeader(http.StatusInternalServerError)
		response.Message = msg
		response.Status = http.StatusText(http.StatusInternalServerError)
	default:
		rw.WriteHeader(http.StatusInternalServerError)
	}

	if translate && len(locale) > 0 {
		translatedMessage, err := translation.Process("error", err.Error(), locale)
		if err != nil || translatedMessage == "" {
			translatedMessage = "error." + response.Message
		}
		response.Message = translatedMessage
	}

	// send response
	e := json.NewEncoder(rw).Encode(response)
	if e != nil {
		return
	}
	return
}
