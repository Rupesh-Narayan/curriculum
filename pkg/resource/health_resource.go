package resource

import (
	"bitbucket.org/noon-micro/curriculum/pkg/lib/middleware"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type HealthResource struct {
	db *sql.DB
}

func NewHealthResource(route *mux.Router, db *sql.DB) {
	resource := &HealthResource{
		db: db,
	}
	route.HandleFunc("/curriculum/health", middleware.UnAuthWrapMiddleware(resource.healthCheck)).Methods("GET")
}

func (h *HealthResource) healthCheck(rw http.ResponseWriter, _ *http.Request) {
	err := h.db.Ping()
	if err != nil {
		data := healthCheckResponse{Status: "Database unreachable"}
		writeJsonResponse(rw, http.StatusServiceUnavailable, data)
	} else {
		data := healthCheckResponse{Status: "UP"}
		writeJsonResponse(rw, http.StatusOK, data)
	}
}

func writeJsonResponse(w http.ResponseWriter, status int, data healthCheckResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return
	}
	return
}

type healthCheckResponse struct {
	Status string `json:"status"`
}
