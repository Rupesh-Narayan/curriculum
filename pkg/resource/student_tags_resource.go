package resource

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	"bitbucket.org/noon-micro/curriculum/pkg/entity"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/helper"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/middleware"
	"bitbucket.org/noon-micro/curriculum/pkg/resource/entity/request"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/oschwald/maxminddb-golang"
	mapper "gopkg.in/jeevatkm/go-model.v1"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type StudentTagsResource struct {
	sts domain.StudentTagsService
}

func NewStudentTagsResource(route *mux.Router, sts domain.StudentTagsService) {
	resource := &StudentTagsResource{
		sts: sts,
	}
	route.HandleFunc("/student/countries", middleware.UnAuthWrapMiddleware(resource.getCountriesNew)).Methods("GET")
	//route.HandleFunc("/student/countries_new", middleware.RecoverHandler(auth.Authenticate("admin", resource.getCountriesNew))).Methods("GET")
	route.HandleFunc("/student/grades", middleware.UnAuthWrapMiddleware(resource.getGrades)).Methods("GET")
	route.HandleFunc("/student/boards", middleware.UnAuthWrapMiddleware(resource.getBoards)).Methods("GET")
	route.HandleFunc("/student/degrees", middleware.UnAuthWrapMiddleware(resource.getDegrees)).Methods("GET")
	route.HandleFunc("/student/majors", middleware.UnAuthWrapMiddleware(resource.getMajors)).Methods("GET")
}

func (t *StudentTagsResource) getCountries(rw http.ResponseWriter, req *http.Request) {
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
	}
	if hierarchy == "" {
		tag.Hierarchy = nil
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), true)
		return
	}
	var getTags domain.GetCountries
	mapper.Copy(&getTags, tag)
	res, err := t.sts.GetCountries(&getTags)
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

func (t *StudentTagsResource) getCountriesNew(rw http.ResponseWriter, req *http.Request) {
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
	fmt.Print(record.Country.ISOCode)
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
		ISOCode: isoCode,
		CountryId: &countryId,
		Locale:    &locale,
		Start:     startInt,
		Limit:     limitInt,
	}
	err = helper.Validate(queryParams)
	var getQueryParams domain.GetCountriesNew
	mapper.Copy(&getQueryParams, queryParams)
	res, err := t.sts.GetCountriesNew(&getQueryParams)
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

func (t *StudentTagsResource) getBoards(rw http.ResponseWriter, req *http.Request) {
	paramsVals, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	countryIdString, _ := paramsVals["country_id"]

	tagType := "board"
	curriculumType := "k12"
	tagGroup := "curriculum"

	locale := req.Header.Get("locale")
	if locale == "" {
		locale = "ar"
	}

	tag := request.GetTagsDTO{
		Type:           &tagType,
		CurriculumType: &curriculumType,
		TagGroup:       &tagGroup,
		Hierarchy:      &countryIdString,
		CountryId:      &countryIdString,
		Locale:         &locale,
	}

	err1 := helper.Validate(tag)
	if err1 != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err1.Error()), req.Header.Get("locale"), true)
		return
	}
	var getTags domain.GetTags
	mapper.Copy(&getTags, tag)

	res, err := t.sts.GetBoards(&getTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, res.Boards, res.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}


func (t *StudentTagsResource) getGrades(rw http.ResponseWriter, req *http.Request) {
	paramsVals, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	countryIdString, _ := paramsVals["country_id"]
	boardId, _ := paramsVals["board_id"]
	tagType := "grade"
	curriculumType := "k12"
	tagGroup := "curriculum"

	locale := req.Header.Get("locale")
	if locale == "" {
		locale = "ar"
	}
	hierarchy := countryIdString
	if len(boardId)>0{
		hierarchy = countryIdString+"."+boardId
	}
	tag := request.GetTagsDTO{
		Type:           &tagType,
		CurriculumType: &curriculumType,
		TagGroup:       &tagGroup,
		Hierarchy:      &hierarchy,
		CountryId:      &countryIdString,
		Locale:         &locale,
	}

	err1 := helper.Validate(tag)
	if err1 != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err1.Error()), req.Header.Get("locale"), true)
		return
	}
	var getTags domain.GetTags
	mapper.Copy(&getTags, tag)

	res, err := t.sts.GetGrades(&getTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	err = new(entity.Response).SendResponse(rw, res.Grades, res.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *StudentTagsResource) getDegrees(rw http.ResponseWriter, req *http.Request) {
	paramsVals, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}
	countryIdString, _ := paramsVals["country_id"]

	tagType := "degree"
	curriculumType := "university"
	tagGroup := "curriculum"

	text, _ := paramsVals["text"]
	start, _ := paramsVals["start"]
	startInt := 0
	if len(start) > 0 {
		startInt, _ = strconv.Atoi(start)
	}
	limit, _ := paramsVals["limit"]
	limitInt := 20
	if len(limit) > 0 {
		limitInt, _ = strconv.Atoi(limit)
	}

	locale := req.Header.Get("locale")
	if locale == "" {
		locale = "ar"
	}

	tag := request.GetTagsSearchDTO{
		Text:           &text,
		Type:           &tagType,
		CurriculumType: &curriculumType,
		TagGroup:       &tagGroup,
		Hierarchy:      &countryIdString,
		CountryId:      &countryIdString,
		Start:          startInt,
		Limit:          limitInt,
		Locale:         &locale,
	}

	err1 := helper.Validate(tag)
	if err1 != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err1.Error()), req.Header.Get("locale"), true)
		return
	}
	var getTags domain.GetTags
	mapper.Copy(&getTags, tag)

	if *getTags.Text == "" {
		getTags.Text = nil
	}

	res, err := t.sts.GetDegrees(&getTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}

	err = new(entity.Response).SendResponse(rw, res.Degrees, res.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}

func (t *StudentTagsResource) getMajors(rw http.ResponseWriter, req *http.Request) {
	paramsVals, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}

	country_id, _ := paramsVals["country_id"]
	degree, _ := paramsVals["degree_id"]

	tagType := "major"

	var hierarchyString = country_id + "." + degree

	curriculumType := "university"
	tagGroup := "curriculum"

	text, _ := paramsVals["text"]
	start, _ := paramsVals["start"]
	startInt := 0
	if len(start) > 0 {
		startInt, _ = strconv.Atoi(start)
	}
	limit, _ := paramsVals["limit"]
	limitInt := 20
	if len(limit) > 0 {
		limitInt, _ = strconv.Atoi(limit)
	}

	locale := req.Header.Get("locale")
	if locale == "" {
		locale = "ar"
	}

	tag := request.GetTagsSearchDTO{
		Text:           &text,
		Type:           &tagType,
		CurriculumType: &curriculumType,
		TagGroup:       &tagGroup,
		Hierarchy:      &hierarchyString,
		CountryId:      &country_id,
		Start:          startInt,
		Limit:          limitInt,
		Locale:         &locale,
	}

	err1 := helper.Validate(tag)
	if err1 != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err1.Error()), req.Header.Get("locale"), true)
		return
	}
	var getTags domain.GetTags
	mapper.Copy(&getTags, tag)

	if *getTags.Text == "" {
		getTags.Text = nil
	}

	res, err := t.sts.GetMajors(&getTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), true)
		return
	}

	err = new(entity.Response).SendResponse(rw, res.Degrees, res.Meta, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), true)
		return
	}
}
