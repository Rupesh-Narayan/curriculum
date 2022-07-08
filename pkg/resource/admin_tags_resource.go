package resource

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	"bitbucket.org/noon-micro/curriculum/pkg/entity"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/helper"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/middleware"
	"bitbucket.org/noon-micro/curriculum/pkg/resource/entity/request"
	entityresponse "bitbucket.org/noon-micro/curriculum/pkg/resource/entity/response"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type AdminTagsResource struct {
	ats domain.AdminTagsService
}

func NewAdminTagsResource(route *mux.Router, ats domain.AdminTagsService) {
	resource := &AdminTagsResource{
		ats: ats,
	}
	route.HandleFunc("/admin/tags", middleware.AuthWrapMiddleware(resource.createTagsFromAdmin, "admin")).Methods("POST")
	route.HandleFunc("/admin/tags", middleware.AuthWrapMiddleware(resource.addTagsToHierarchy, "admin")).Methods("PUT")
	route.HandleFunc("/admin/tags/multiple", middleware.AuthWrapMiddleware(resource.addMultipleTagsToHierarchy, "admin")).Methods("PUT")
	route.HandleFunc("/admin/tags/update", middleware.AuthWrapMiddleware(resource.updateTag, "admin")).Methods("PUT")
	route.HandleFunc("/admin/tags/order", middleware.AuthWrapMiddleware(resource.updateTagOrder, "admin")).Methods("PUT")
	route.HandleFunc("/admin/tags/locale/{action}", middleware.AuthWrapMiddleware(resource.updateLocaleForTags, "admin")).Methods("PUT")
	route.HandleFunc("/admin/tags/delete/hierarchy", middleware.AuthWrapMiddleware(resource.removeTagsFromHierarchy, "admin")).Methods("PUT")
	route.HandleFunc("/admin/tags/delete/identifier", middleware.AuthWrapMiddleware(resource.removeIdentifier, "admin")).Methods("PUT")
	route.HandleFunc("/admin/tags", middleware.AuthWrapMiddleware(resource.getTags, "admin")).Methods("GET")
	route.HandleFunc("/admin/tags/{id:[0-9]+}", middleware.AuthWrapMiddleware(resource.getTag, "admin")).Methods("GET")
	route.HandleFunc("/admin/tags/search", middleware.AuthWrapMiddleware(resource.getTagsSearch, "admin")).Methods("GET")
	route.HandleFunc("/admin/elastic/migrate", middleware.UnAuthWrapMiddleware(resource.migrateToElastic)).Methods("POST")

	route.HandleFunc("/admin/boards", middleware.AuthWrapMiddleware(resource.getBoardTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/grades", middleware.AuthWrapMiddleware(resource.getGradeTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/degrees", middleware.AuthWrapMiddleware(resource.getDegreeTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/majors", middleware.AuthWrapMiddleware(resource.getMajorTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/courses", middleware.AuthWrapMiddleware(resource.getCourseTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/sections", middleware.AuthWrapMiddleware(resource.getSectionTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/subjects", middleware.AuthWrapMiddleware(resource.getSubjectTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/curriculum", middleware.AuthWrapMiddleware(resource.getCurriculumTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/tests", middleware.AuthWrapMiddleware(resource.getTestTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/skills", middleware.AuthWrapMiddleware(resource.getSkillTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/chapters", middleware.AuthWrapMiddleware(resource.getChapterTags, "admin.supply")).Methods("GET")
	route.HandleFunc("/admin/topics", middleware.AuthWrapMiddleware(resource.getTopicTags, "admin.supply")).Methods("GET")

	route.HandleFunc("/admin/countries", middleware.AuthWrapMiddleware(resource.getCountriesNew, "admin.supply")).Methods("GET")

}

func (t *AdminTagsResource) createTagsFromAdmin(rw http.ResponseWriter, req *http.Request) {
	var tag request.CreateTagsForAdminDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), true)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var createTag domain.CreateTags
	if err = copier.Copy(&createTag, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	userId, err := strconv.ParseInt(req.Header.Get("Userid"), 10, 64)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
	createTag.CreatorId = &userId
	createTag.CreatorType = new(string)
	*createTag.CreatorType = "admin"
	res, err := t.ats.CreateAdminTags(createTag.TagGroup, &createTag)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, res, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) addTagsToHierarchy(rw http.ResponseWriter, req *http.Request) {
	var tag request.UpdateHierarchyDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), true)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var updateTag domain.UpdateTags
	if err = copier.Copy(&updateTag, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.ats.UpdateAdminTags(updateTag.TagGroup, &updateTag)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, res, nil, http.StatusCreated)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) addMultipleTagsToHierarchy(rw http.ResponseWriter, req *http.Request) {
	var tag request.UpdateMultipleHierarchyDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), true)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var updateTag domain.UpdateMultipleTags
	if err = copier.Copy(&updateTag, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.ats.UpdateMultipleAdminTags(&updateTag)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, res, nil, http.StatusCreated)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) updateTag(rw http.ResponseWriter, req *http.Request) {
	var tag request.UpdateTagDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), true)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var updateTag domain.UpdateTag
	if err = copier.Copy(&updateTag, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	err = t.ats.UpdateTag(&updateTag)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, nil, nil, http.StatusCreated)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) updateTagOrder(rw http.ResponseWriter, req *http.Request) {
	var tag request.UpdateTagOrderDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), true)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var updateTag domain.UpdateTagOrder
	if err = copier.Copy(&updateTag, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	err = t.ats.UpdateTagOrder(updateTag.TagGroup, &updateTag)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, nil, nil, http.StatusCreated)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) removeTagsFromHierarchy(rw http.ResponseWriter, req *http.Request) {
	var tag request.RemoveHierarchyDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), true)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var removeHierarchy domain.RemoveHierarchy
	if err = copier.Copy(&removeHierarchy, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.ats.RemoveAdminTagFromHierarchy(removeHierarchy.TagGroup, &removeHierarchy)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, res, nil, http.StatusCreated)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) removeIdentifier(rw http.ResponseWriter, req *http.Request) {
	var tag request.RemoveIdentifierDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), true)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	err = t.ats.RemoveIdentifierTag(tag.ID)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, nil, nil, http.StatusCreated)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) migrateToElastic(rw http.ResponseWriter, req *http.Request) {
	var tag request.MigrateToElastic
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), true)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	err = t.ats.MigrateToElastic(tag.Start, tag.End)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, nil, nil, http.StatusCreated)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) getTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	tagType, _ := params["type"]
	curriculumType, _ := params["curriculum_type"]
	tagGroup, _ := params["tag_group"]
	hierarchy, _ := params["hierarchy"]
	identifier, _ := params["identifier"]
	countryId, _ := params["country_id"]
	locale, _ := params["locale"]
	multiGrade, _ := params["multi_grade"]
	identifiersString := strings.Split(identifier, ",")
	var identifiers []*string
	if len(identifier) > 0 {
		for _, v := range identifiersString {
			val := v
			identifiers = append(identifiers, &val)
		}
	}
	tag := request.GetTagsDTO{
		Type:           &tagType,
		CurriculumType: &curriculumType,
		TagGroup:       &tagGroup,
		Hierarchy:      &hierarchy,
		Identifier:     identifiers,
		CountryId:      &countryId,
		Locale:         &locale,
		MultiGrade:     &multiGrade,
	}
	if hierarchy == "" {
		tag.Hierarchy = nil
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var getTags domain.GetTags
	if err = copier.Copy(&getTags, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.ats.GetTags(&getTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, res.Tags, res.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) getTag(rw http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	tagIdString := params["id"]
	res, err := t.ats.GetTag(&tagIdString)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	response := new(entityresponse.TagLocaleInfoResponseDTO)
	if err = copier.Copy(response, res); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	err = new(entity.Response).SendResponse(rw, response, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) getTagsSearch(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	text, _ := params["text"]
	tagType, _ := params["type"]
	curriculumType, _ := params["curriculum_type"]
	tagGroup, _ := params["tag_group"]
	countryId, _ := params["country_id"]
	locale, _ := params["locale"]
	start, _ := params["start"]
	startInt := 0
	if len(start) > 0 {
		startInt, _ = strconv.Atoi(start)
	}
	limit, _ := params["limit"]
	limitInt := 20
	if len(limit) > 0 {
		limitInt, _ = strconv.Atoi(limit)
	}
	hierarchy, _ := params["hierarchy"]
	identifier, _ := params["identifier"]
	identifiersString := strings.Split(identifier, ",")
	var identifiers []*string
	if len(identifier) > 0 {
		for _, v := range identifiersString {
			val := v
			identifiers = append(identifiers, &val)
		}
	}
	tag := request.GetTagsSearchDTO{
		Text:           &text,
		Type:           &tagType,
		CurriculumType: &curriculumType,
		TagGroup:       &tagGroup,
		CountryId:      &countryId,
		Locale:         &locale,
		Hierarchy:      &hierarchy,
		Identifier:     identifiers,
		Start:          startInt,
		Limit:          limitInt,
	}
	if text == "" {
		tag.Text = nil
	}
	if hierarchy == "" {
		tag.Hierarchy = nil
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var getTags domain.GetTags
	if err = copier.Copy(&getTags, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.ats.GetTagsSearch(&getTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	var responses []*entityresponse.AdminTagResponseSearchDTO
	for _, v := range res.Tags {
		response := new(entityresponse.AdminTagResponseSearchDTO)
		if err = copier.Copy(response, v); err != nil {
			entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
		}
		responses = append(responses, response)
	}
	err = new(entity.Response).SendResponse(rw, responses, res.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func getQueryParams(req *http.Request) (params map[string]string, err error) {
	params = make(map[string]string)
	if len(req.URL.RawQuery) > 0 {
		rawParams := strings.Split(req.URL.RawQuery, "&")
		for _, v := range rawParams {
			keyValue := strings.Split(v, "=")
			if len(keyValue) == 2 {
				val, err := url.QueryUnescape(keyValue[1])
				if err != nil {
					return nil, noonerror.New(noonerror.ErrInternalServer, "paramParsingError")
				}
				params[keyValue[0]] = val
			} else {
				return nil, noonerror.New(noonerror.ErrInvalidRequest, "invalidParams")
			}
		}
	}
	return
}

func (t *AdminTagsResource) updateLocaleForTags(rw http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	action := params["action"]
	if action != "add" && action != "delete" {
		entity.HandleError(rw, "notFound", noonerror.ErrUserNotFound, req.Header.Get("locale"), true)
		return
	}
	var tag request.TagLocaleDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), true)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var tagLocale domain.TagLocale
	if err = copier.Copy(&tagLocale, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	err = t.ats.UpdateTagLocale(&action, &tagLocale)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, nil, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) getBoardTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)

	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTagsForLibrary(domain.TagTypeEnum.Board, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12, params, rw, req)
}

func (t *AdminTagsResource) getGradeTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)

	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTagsForLibrary(domain.TagTypeEnum.Grade, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12, params, rw, req)
}

func (t *AdminTagsResource) getDegreeTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTagsForLibrary(domain.TagTypeEnum.Degree, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.University, params, rw, req)
}

func (t *AdminTagsResource) getMajorTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTagsForLibrary(domain.TagTypeEnum.Major, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.University, params, rw, req)

}

func (t *AdminTagsResource) getCourseTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTagsForLibrary(domain.TagTypeEnum.Course, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.University, params, rw, req)

}

func (t *AdminTagsResource) getSectionTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTagsForLibrary(domain.TagTypeEnum.Section, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.Default, params, rw, req)
}

func (t *AdminTagsResource) getSubjectTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTagsForLibrary(domain.TagTypeEnum.Subject, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12, params, rw, req)
}

func (t *AdminTagsResource) getCurriculumTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTagsForLibrary(domain.TagTypeEnum.Curriculum, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12, params, rw, req)
}

func (t *AdminTagsResource) getTestTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	ct, _ := params["curriculum_type"]
	if len(ct) > 0 {
		t.getTagsForLibrary(domain.TagTypeEnum.Test, domain.TagGroupEnum.Curriculum, ct, params, rw, req)
	} else {
		t.getTestSkillTags(domain.TagTypeEnum.Test, domain.TagGroupEnum.Curriculum, params, rw, req)
	}
}

func (t *AdminTagsResource) getSkillTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	ct, _ := params["curriculum_type"]
	if len(ct) > 0 {
		t.getTagsForLibrary(domain.TagTypeEnum.Skill, domain.TagGroupEnum.Curriculum, ct, params, rw, req)
	} else {
		t.getTestSkillTags(domain.TagTypeEnum.Skill, domain.TagGroupEnum.Curriculum, params, rw, req)
	}
}

func (t *AdminTagsResource) getChapterTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTagsForLibrary(domain.TagTypeEnum.Chapter, domain.TagGroupEnum.Content, domain.CurriculumTypeEnum.Default, params, rw, req)
}

func (t *AdminTagsResource) getTopicTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTagsForLibrary(domain.TagTypeEnum.Topic, domain.TagGroupEnum.Content, domain.CurriculumTypeEnum.Default, params, rw, req)
}

func (t *AdminTagsResource) getTagsForLibrary(tagType string, tagGroup string, curriculumType string, params map[string]string, rw http.ResponseWriter, req *http.Request) {

	ct, _ := params["curriculum_type"]
	if len(ct) > 0 {
		curriculumType = ct
	}
	countryId, _ := params["country_id"]
	userId, err := strconv.ParseInt(req.Header.Get("Userid"), 10, 64)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
	locale := req.Header.Get("locale")
	boardId, _ := params["board_id"]
	gradeId, _ := params["grade_id"]
	subjectId, _ := params["subject_id"]
	curriculumId, _ := params["curriculum_id"]
	degreeId, _ := params["degree_id"]
	majorId, _ := params["major_id"]
	courseId, _ := params["course_id"]
	sectionId, _ := params["section_id"]
	testId, _ := params["test_id"]
	skillId, _ := params["skill_id"]
	chapterId, _ := params["chapter_id"]
	topicId, _ := params["topic_id"]

	text, _ := params["text"]
	var textPtr *string
	start, _ := params["start"]
	startInt := 0
	if len(start) > 0 {
		startInt, _ = strconv.Atoi(start)
	}
	limit, _ := params["limit"]
	limitInt := 100
	if len(text) > 0 {
		limitInt = 20
		textPtr = &text
	}
	if len(limit) > 0 {
		limitInt, _ = strconv.Atoi(limit)
	}

	hierarchySlice := []string{countryId,boardId, gradeId, subjectId, curriculumId, degreeId, majorId, courseId, sectionId, testId, skillId, chapterId, topicId}
	var hierarchies []*string
	if len(hierarchySlice) > 0 {
		for _, v := range hierarchySlice {
			val := v
			if len(val) > 0 {
				hierarchies = append(hierarchies, &val)
			}
		}
	}
	tag := request.GetAdminTagsDTO{
		Text:           textPtr,
		Type:           &tagType,
		CurriculumType: &curriculumType,
		TagGroup:       &tagGroup,
		Hierarchy:      hierarchies,
		Locale:         &locale,
		CountryId:      &countryId,
		CreatorId:      &userId,
		Start:          startInt,
		Limit:          limitInt,
	}

	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var getAdminTags domain.GetAdminTags
	if err = copier.Copy(&getAdminTags, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.ats.GetAdminTags(&getAdminTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	var getAdminTagsResponse entityresponse.GetAdminTagsResponseDTO
	if err = copier.Copy(&getAdminTagsResponse, &res); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	for _, v := range getAdminTagsResponse.Tags {
		if v == nil {
			continue
		}
		if v.LocaleName != nil {
			v.Name = v.LocaleName
		}
		v.LocaleName = nil
	}
	err = new(entity.Response).SendResponse(rw, getAdminTagsResponse.Tags, getAdminTagsResponse.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) getTestSkillTags(tagType string, tagGroup string, params map[string]string, rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	curriculumType := "misc"
	countryId, _ := params["country_id"]
	userId, err := strconv.ParseInt(req.Header.Get("Userid"), 10, 64)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
	locale := req.Header.Get("locale")

	gradeId, _ := params["grade_id"]
	subjectId, _ := params["subject_id"]
	curriculumId, _ := params["curriculum_id"]
	degreeId, _ := params["degree_id"]
	majorId, _ := params["major_id"]
	courseId, _ := params["course_id"]
	sectionId, _ := params["section_id"]
	testId, _ := params["test_id"]
	skillId, _ := params["skill_id"]
	chapterId, _ := params["chapter_id"]
	topicId, _ := params["topic_id"]

	text, _ := params["text"]
	var textPtr *string
	start, _ := params["start"]
	startInt := 0
	if len(start) > 0 {
		startInt, _ = strconv.Atoi(start)
	}
	limit, _ := params["limit"]
	limitInt := 100
	if len(text) > 0 {
		limitInt = 20
		textPtr = &text
	}
	if len(limit) > 0 {
		limitInt, _ = strconv.Atoi(limit)
	}

	hierarchySlice := []string{countryId, gradeId, subjectId, curriculumId, degreeId, majorId, courseId, sectionId, testId, skillId, chapterId, topicId}
	var hierarchies []*string
	if len(hierarchySlice) > 0 {
		for _, v := range hierarchySlice {
			val := v
			if len(val) > 0 {
				hierarchies = append(hierarchies, &val)
			}
		}
	}
	tag := request.GetAdminTagsDTO{
		Text:           textPtr,
		Type:           &tagType,
		CurriculumType: &curriculumType,
		TagGroup:       &tagGroup,
		Hierarchy:      hierarchies,
		Locale:         &locale,
		CountryId:      &countryId,
		CreatorId:      &userId,
		Start:          startInt,
		Limit:          limitInt,
	}

	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var getAdminTags domain.GetAdminTags
	if err = copier.Copy(&getAdminTags, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.ats.GetTestsSkillsForLibrary(&getAdminTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	var getAdminTagsResponse entityresponse.GetAdminTagsResponseDTO
	if err = copier.Copy(&getAdminTagsResponse, &res); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	for _, v := range getAdminTagsResponse.Tags {
		if v == nil {
			continue
		}
		if v.LocaleName != nil {
			v.Name = v.LocaleName
		}
		v.LocaleName = nil
	}
	err = new(entity.Response).SendResponse(rw, getAdminTagsResponse.Tags, getAdminTagsResponse.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *AdminTagsResource) getCountriesNew(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}

	countryId := req.Header.Get("country")
	locale := req.Header.Get("locale")

	ipAddress := req.Header.Get("X-FORWARDED-FOR")
	start, _ := params["start"]
	startInt := 0
	if len(start) > 0 {
		startInt, _ = strconv.Atoi(start)
	}
	limit, _ := params["limit"]
	limitInt := 50
	if len(limit) > 0 {
		limitInt, _ = strconv.Atoi(limit)
	}
	queryParams := request.GetCountriesNewDTO{
		IpAddress: ipAddress,
		CountryId: &countryId,
		Locale:    &locale,
		Start:     startInt,
		Limit:     limitInt,
	}
	var getQueryParams domain.GetCountriesNew
	if err = copier.Copy(&getQueryParams, &queryParams); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.ats.GetCountriesTagsNew(&getQueryParams)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, res.Tags, res.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}
