package middleware

import (
	"bitbucket.org/noon-go/auth"
	"bitbucket.org/noon-go/translator"
	"bitbucket.org/noon-micro/curriculum/pkg/entity"
	loggers "bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"strconv"
)

type middleware func(http.HandlerFunc) http.HandlerFunc

var noonAuthEntity *auth.AuthenticateEntity

// InitializeMiddleware sets http client for middleware
func InitializeMiddleware(authEntity *auth.AuthenticateEntity) {
	noonAuthEntity = authEntity
}

// AuthWrapMiddleware wraps and applies multiple middleware and auth middleware
func AuthWrapMiddleware(next http.HandlerFunc, roles string) http.HandlerFunc {
	applyMiddleware := []middleware{
		recoverHandler,
	}
	wrapped := authMiddleware(next, roles)

	// loop in reverse to preserve middleware order
	for i := len(applyMiddleware) - 1; i >= 0; i-- {
		wrapped = applyMiddleware[i](wrapped)
	}

	return wrapped
}

// UnAuthWrapMiddleware wraps and applies multiple middleware without auth
func UnAuthWrapMiddleware(next http.HandlerFunc) http.HandlerFunc {
	applyMiddleware := []middleware{
		recoverHandler,
	}
	wrapped := next

	// loop in reverse to preserve middleware order
	for i := len(applyMiddleware) - 1; i >= 0; i-- {
		wrapped = applyMiddleware[i](wrapped)
	}

	return wrapped
}

// recoverHandler middleware to recover from
func recoverHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				loggers.Client.Error(loggers.GetErrorStack())
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(500)
				errorResponse := entity.ErrorResponse{
					Status: http.StatusText(500),
					Message: "codePanicked",
				}
				_ = json.NewEncoder(w).Encode(errorResponse)
			}
		}()
		next.ServeHTTP(w, r)
	}
}

// authMiddleware calls auth module and process request
func authMiddleware(next http.HandlerFunc, roles string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := noonAuthEntity.Process(roles, r.Header)
		if err != nil {
			locale := r.Header.Get("locale")
			sendError(w, err, locale)
			return
		}
		userId := strconv.Itoa(response.UserID)
		r.Header.Set("userId", userId)

		next.ServeHTTP(w, r)
	}
}

func sendError(w http.ResponseWriter, err error, locale string) {
	errorResponse := entity.ErrorResponse{
		Status: http.StatusText(500),
		Message: "somethingWentWrong",
	}
	httpStatusCode := http.StatusInternalServerError
	if e, ok := err.(*auth.Error); ok {
		errorResponse.Message = e.Message
		errorResponse.Status = http.StatusText(e.StatusCode)
		httpStatusCode = e.GetHTTPStatusCode()
	} else {
		logrus.Error("authentication returned error : ", err)
	}

	translatedMessage, err := translation.Process("error", errorResponse.Message, locale)
	if err != nil || translatedMessage == "" {
		translatedMessage = "error." + errorResponse.Message
	}
	errorResponse.Message = translatedMessage

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	_ = json.NewEncoder(w).Encode(errorResponse)
}
