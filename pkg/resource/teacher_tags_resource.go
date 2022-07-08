package resource

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	"bitbucket.org/noon-micro/curriculum/pkg/entity"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/helper"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/middleware"
	"bitbucket.org/noon-micro/curriculum/pkg/resource/entity/request"
	"bitbucket.org/noon-micro/curriculum/pkg/resource/entity/response"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	"github.com/oschwald/maxminddb-golang"
	"log"
	"net"
	"net/http"
	"strconv"
)

type TeacherTagsResource struct {
	tts domain.TeacherTagsService
}

func NewTeacherTagsResource(route *mux.Router, tts domain.TeacherTagsService) {
	resource := &TeacherTagsResource{
		tts: tts,
	}
	route.HandleFunc("/teacher/grades", middleware.AuthWrapMiddleware(resource.getGradeTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/boards", middleware.AuthWrapMiddleware(resource.getBoardTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/degrees", middleware.AuthWrapMiddleware(resource.getDegreeTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/majors", middleware.AuthWrapMiddleware(resource.getMajorTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/courses", middleware.AuthWrapMiddleware(resource.getCourseTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/sections", middleware.AuthWrapMiddleware(resource.getSectionTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/subjects", middleware.AuthWrapMiddleware(resource.getSubjectTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/curriculum", middleware.AuthWrapMiddleware(resource.getCurriculumTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/tests", middleware.AuthWrapMiddleware(resource.getTestTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/skills", middleware.AuthWrapMiddleware(resource.getSkillTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/chapters", middleware.AuthWrapMiddleware(resource.getChapterTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/topics", middleware.AuthWrapMiddleware(resource.getTopicTags, "teacher")).Methods("GET")
	route.HandleFunc("/teacher/countries", middleware.UnAuthWrapMiddleware(resource.getCountriesNew)).Methods("GET")
}

func (t *TeacherTagsResource) getBoardTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)

	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTags(domain.TagTypeEnum.Board, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12, params, rw, req)
}

func (t *TeacherTagsResource) getGradeTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)

	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTags(domain.TagTypeEnum.Grade, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12, params, rw, req)
}

func (t *TeacherTagsResource) getDegreeTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTags(domain.TagTypeEnum.Degree, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.University, params, rw, req)
}

func (t *TeacherTagsResource) getMajorTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTags(domain.TagTypeEnum.Major, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.University, params, rw, req)
}

func (t *TeacherTagsResource) getCourseTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTags(domain.TagTypeEnum.Course, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.University, params, rw, req)
}

func (t *TeacherTagsResource) getSectionTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTags(domain.TagTypeEnum.Section, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.Default, params, rw, req)
}

func (t *TeacherTagsResource) getSubjectTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTags(domain.TagTypeEnum.Subject, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12, params, rw, req)
}

func (t *TeacherTagsResource) getCurriculumTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTags(domain.TagTypeEnum.Curriculum, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12, params, rw, req)
}

func (t *TeacherTagsResource) getTestTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	ct, _ := params["curriculum_type"]
	if len(ct) > 0 {
		t.getTags(domain.TagTypeEnum.Test, domain.TagGroupEnum.Curriculum, ct, params, rw, req)
	} else {
		t.getTestSkillTags(domain.TagTypeEnum.Test, domain.TagGroupEnum.Curriculum, params, rw, req)
	}
}

func (t *TeacherTagsResource) getSkillTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	ct, _ := params["curriculum_type"]
	if len(ct) > 0 {
		t.getTags(domain.TagTypeEnum.Skill, domain.TagGroupEnum.Curriculum, ct, params, rw, req)
	} else {
		t.getTestSkillTags(domain.TagTypeEnum.Skill, domain.TagGroupEnum.Curriculum, params, rw, req)
	}
}

func (t *TeacherTagsResource) getChapterTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTags(domain.TagTypeEnum.Chapter, domain.TagGroupEnum.Content, domain.CurriculumTypeEnum.Default, params, rw, req)
}

func (t *TeacherTagsResource) getTopicTags(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	t.getTags(domain.TagTypeEnum.Topic, domain.TagGroupEnum.Content, domain.CurriculumTypeEnum.Default, params, rw, req)
}

func (t *TeacherTagsResource) getTags(tagType string, tagGroup string, curriculumType string, params map[string]string, rw http.ResponseWriter, req *http.Request) {

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

	hierarchySlice := []string{countryId, boardId, gradeId, subjectId, curriculumId, degreeId, majorId, courseId, sectionId, testId, skillId, chapterId, topicId}
	var hierarchies []*string
	if len(hierarchySlice) > 0 {
		for _, v := range hierarchySlice {
			val := v
			if len(val) > 0 {
				hierarchies = append(hierarchies, &val)
			}
		}
	}
	tag := request.GetTeacherTagsDTO{
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
	var getTeacherTags domain.GetTeacherTags
	if err = copier.Copy(&getTeacherTags, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.tts.GetTeacherTags(&getTeacherTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	var getTeacherTagsResponse response.GetTeacherTagsResponseDTO
	if err = copier.Copy(&getTeacherTagsResponse, &res); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	for _, v := range getTeacherTagsResponse.Tags {
		if v == nil {
			continue
		}
		if v.LocaleName != nil {
			v.Name = v.LocaleName
		}
		v.LocaleName = nil
	}
	err = new(entity.Response).SendResponse(rw, getTeacherTagsResponse.Tags, getTeacherTagsResponse.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *TeacherTagsResource) getTestSkillTags(tagType string, tagGroup string, params map[string]string, rw http.ResponseWriter, req *http.Request) {
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
	tag := request.GetTeacherTagsDTO{
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
	var getTeacherTags domain.GetTeacherTags
	if err = copier.Copy(&getTeacherTags, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.tts.GetTestsSkillsForLibrary(&getTeacherTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	var getTeacherTagsResponse response.GetTeacherTagsResponseDTO
	if err = copier.Copy(&getTeacherTagsResponse, &res); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	for _, v := range getTeacherTagsResponse.Tags {
		if v == nil {
			continue
		}
		if v.LocaleName != nil {
			v.Name = v.LocaleName
		}
		v.LocaleName = nil
	}
	err = new(entity.Response).SendResponse(rw, getTeacherTagsResponse.Tags, getTeacherTagsResponse.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *TeacherTagsResource) getCountriesNew(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}

	countryId := req.Header.Get("country")
	locale := req.Header.Get("locale")

	ipAddress := req.Header.Get("X-FORWARDED-FOR")

	db, err := maxminddb.Open("GeoIP2-Country.mmdb")

	if err != nil {

		log.Fatal(err)

	}

	defer db.Close()


	ip := net.ParseIP(ipAddress)


	var record struct {

		Country struct {

			ISOCode string `maxminddb:"iso_code"`

		} `maxminddb:"country"`

	}


	err = db.Lookup(ip, &record)

	if err != nil {

		log.Fatal(err)

	}

	var isoCode = "SA"

	if record.Country.ISOCode != "" {

		isoCode = record.Country.ISOCode

	}

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
		ISOCode:   isoCode,
		CountryId: &countryId,
		Locale:    &locale,
		Start:     startInt,
		Limit:     limitInt,
	}
	var getQueryParams domain.GetCountriesNew
	if err = copier.Copy(&getQueryParams, &queryParams); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), true)
	}
	res, err := t.tts.GetCountriesTagsNew(&getQueryParams)
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
