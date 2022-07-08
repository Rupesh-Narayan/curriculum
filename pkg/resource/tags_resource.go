package resource

//
//import (
//	"bitbucket.org/noon-micro/auth"
//	"bitbucket.org/noon-micro/curriculum/pkg/domain"
//	"bitbucket.org/noon-micro/curriculum/pkg/entity"
//	"bitbucket.org/noon-micro/curriculum/pkg/lib/error"
//	"bitbucket.org/noon-micro/curriculum/pkg/lib/helper"
//	"bitbucket.org/noon-micro/curriculum/pkg/lib/middleware"
//	"encoding/json"
//	"github.com/gorilla/mux"
//	"net/http"
//	"strconv"
//)
//
//type TagsResource struct {
//	ts domain.TagsService
//}
//
//func NewTagsResource(route *mux.Router, ts domain.TagsService) {
//	resource := &TagsResource{
//		ts: ts,
//	}
//	route.HandleFunc("/tags/{id:[0-9]+}", middleware.RecoverHandler(auth.Authenticate("student.teacher", resource.fetchTags))).Methods("GET")
//	route.HandleFunc("/tags", middleware.RecoverHandler(auth.Authenticate("student.teacher", resource.createTags))).Methods("POST")
//}
//
//func (t *TagsResource) fetchTags(rw http.ResponseWriter, req *http.Request) {
//	vars := mux.Vars(req)
//	id := vars["id"]
//	idInt, err := strconv.ParseInt(id, 10, 64)
//	if err != nil {
//		entity.HandleError(rw, "tagIdInvalid", noonerror.ErrInvalidRequest)
//		return
//
//	}
//	tags, err := t.ts.FetchTags(&idInt)
//	err = new(entity.Response).SendResponse(rw, tags, http.StatusOK)
//	if err != nil {
//		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer)
//		return
//	}
//}
//
//func (t *TagsResource) createTags(rw http.ResponseWriter, req *http.Request) {
//	var tag domain.Tags
//	err := json.NewDecoder(req.Body).Decode(&tag)
//	if err != nil {
//		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest)
//		return
//	}
//	err = helper.Validate(tag)
//	if err != nil {
//		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()))
//		return
//	}
//	_,err = t.ts.CreateTags(&tag)
//	if err != nil {
//		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer)
//		return
//	}
//	err = new(entity.Response).SendResponse(rw, nil, http.StatusCreated)
//	if err != nil {
//		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer)
//		return
//	}
//}
