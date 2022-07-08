package resource

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	"bitbucket.org/noon-micro/curriculum/pkg/entity"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/helper"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/middleware"
	"bitbucket.org/noon-micro/curriculum/pkg/resource/entity/request"
	"bitbucket.org/noon-micro/curriculum/pkg/resource/entity/response"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	"net/http"
)

type RpcTagsResource struct {
	rts domain.RpcTagsService
}

func NewRpcTagsResource(route *mux.Router, rts domain.RpcTagsService) {
	resource := &RpcTagsResource{
		rts: rts,
	}
	route.HandleFunc("/rpc/getTags", middleware.UnAuthWrapMiddleware(resource.getTagsList)).Methods("POST")
	route.HandleFunc("/rpc/getTagsByHierarchy", middleware.UnAuthWrapMiddleware(resource.getTagsByHierarchy)).Methods("POST")
	route.HandleFunc("/rpc/validateHierarchy", middleware.UnAuthWrapMiddleware(resource.validateHierarchy)).Methods("POST")
	route.HandleFunc("/rpc/createTags", middleware.UnAuthWrapMiddleware(resource.createTags)).Methods("POST")
	route.HandleFunc("/rpc/getDefaultTags", middleware.UnAuthWrapMiddleware(resource.getDefaultTags)).Methods("POST")
	route.HandleFunc("/rpc/getSuggestedTags", middleware.UnAuthWrapMiddleware(resource.getSuggestedTags)).Methods("POST")
	route.HandleFunc("/rpc/getLegacyDataFromTagId", middleware.UnAuthWrapMiddleware(resource.getLegacyDataFromTagId)).Methods("POST")
	route.HandleFunc("/rpc/getLegacyDataFromTagIds", middleware.UnAuthWrapMiddleware(resource.getLegacyDataFromTagIds)).Methods("POST")
	route.HandleFunc("/rpc/getTagDataFromLegacyId", middleware.UnAuthWrapMiddleware(resource.getTagDataFromLegacyId)).Methods("POST")
	route.HandleFunc("/rpc/getGradeTags", middleware.UnAuthWrapMiddleware(resource.getGradeTags)).Methods("POST")

	route.HandleFunc("/rpc/getK12Products", middleware.UnAuthWrapMiddleware(resource.getK12Products)).Methods("POST")
	route.HandleFunc("/rpc/getUniversityProducts", middleware.UnAuthWrapMiddleware(resource.getUniversityProducts)).Methods("POST")
}

func (t *RpcTagsResource) createTags(rw http.ResponseWriter, req *http.Request) {
	var tag request.CreateTagsRPCDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	curriculumType := domain.CurriculumTypeEnum.Default
	tag.CurriculumType = &curriculumType
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	var createTags domain.CreateMultipleTags
	if err = copier.Copy(&createTags, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), false)
	}
	createTags.CreatorType = new(string)
	*createTags.CreatorType = "teacher"
	res, err := t.rts.CreateTags(&createTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	err = new(entity.Response).SendResponse(rw, res, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) getTagsList(rw http.ResponseWriter, req *http.Request) {
	params, err := getQueryParams(req)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	localeString, _ := params["locale"]
	var locale bool
	if len(localeString) > 0 && (localeString == "true" || localeString == "1") {
		locale = true
	}
	var tag request.GetTagsRPCDTO
	err = json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	var getTags domain.GetTagsByIds
	if err = copier.Copy(&getTags, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), false)
	}
	res, err := t.rts.GetTagsByIds(&getTags, locale)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	var responses []*response.TagLocaleInfoResponseDTO
	for _, v := range res {
		resp := new(response.TagLocaleInfoResponseDTO)
		if err := copier.Copy(resp, v); err != nil {
			entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), false)
			return
		}
		responses = append(responses, resp)
	}
	err = new(entity.Response).SendResponse(rw, responses, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) getTagsByHierarchy(rw http.ResponseWriter, req *http.Request) {
	var tag request.GetTagsByHierarchyRPCDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	var getTags domain.GetTags
	if err = copier.Copy(&getTags, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), false)
	}
	res, err := t.rts.GetTags(&getTags)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	err = new(entity.Response).SendResponse(rw, res, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) validateHierarchy(rw http.ResponseWriter, req *http.Request) {
	var tag request.ValidateHierarchyDTO
	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	curriculumType := domain.CurriculumTypeEnum.Default
	tag.CurriculumType = &curriculumType
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	var validateHierarchy domain.ValidateHierarchy
	if err = copier.Copy(&validateHierarchy, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), false)
	}
	err = t.rts.ValidateHierarchy(&validateHierarchy)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	err = new(entity.Response).SendResponse(rw, nil, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) getDefaultTags(rw http.ResponseWriter, _ *http.Request) {

	res, err := t.rts.GetDefaultTags()
	if err != nil {
		entity.HandleError(rw, "", err, "", false)
		return
	}
	err = new(entity.Response).SendResponse(rw, []*domain.DefaultTags{res}, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, "", false)
		return
	}
}

func (t *RpcTagsResource) getSuggestedTags(rw http.ResponseWriter, req *http.Request) {
	var tag request.GetSuggestedTagsDTO

	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	curriculumType := domain.CurriculumTypeEnum.Default
	tag.CurriculumType = &curriculumType
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	var gst domain.GetSuggestedTags
	if err = copier.Copy(&gst, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), false)
	}
	res, err := t.rts.GetSuggestedCurriculum(&gst)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	err = new(entity.Response).SendResponse(rw, res, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) getLegacyDataFromTagId(rw http.ResponseWriter, req *http.Request) {
	var tag request.GetLegacyDataFromTagIdDTO

	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	res, err := t.rts.GetLegacyDataFromTagId(tag.ID)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	err = new(entity.Response).SendResponse(rw, res, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) getLegacyDataFromTagIds(rw http.ResponseWriter, req *http.Request) {
	var tag request.GetLegacyDataFromTagIdsDTO

	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	resp, err := t.rts.GetLegacyDataFromTagIds(tag.IDs)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	err = new(entity.Response).SendResponse(rw, resp, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) getTagDataFromLegacyId(rw http.ResponseWriter, req *http.Request) {
	var tag request.GetTagDataFromLegacyIdDTO

	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	res, err := t.rts.GetTagDataFromLegacyId(tag.Type, tag.ID)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	err = new(entity.Response).SendResponse(rw, res, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) getGradeTags(rw http.ResponseWriter, req *http.Request) {
	var tag request.GetGradeTagsDTO

	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	zero := "0"
	if tag.ProductId == nil {
		tag.ProductId = &zero
	}
	res, err := t.rts.GetGradeTags(tag.Grade, tag.ProductId)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	err = new(entity.Response).SendResponse(rw, res, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) getK12Products(rw http.ResponseWriter, req *http.Request) {
	var tag request.GetK12ProductsDTO

	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	var genericProductsDTO request.GenericProductsDTO
	if err = copier.Copy(&genericProductsDTO, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), false)
	}
	var responses []*response.RpcTagResponseDTO
	res, err := t.getTagsByRpc(domain.TagTypeEnum.Subject, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12, &genericProductsDTO)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	if len(res.Tags) > 0 {
		responses = append(responses, res.Tags...)
	}
	if genericProductsDTO.Test != nil && *genericProductsDTO.Test == true {
		testProductDTO := request.GenericProductsDTO{CountryId: genericProductsDTO.CountryId, Locale: genericProductsDTO.Locale}
		resTest, err := t.getTagsByRpc(domain.TagTypeEnum.Test, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12TestPrep, &testProductDTO)
		if err != nil {
			entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
			return
		}
		if len(resTest.Tags) > 0 {
			responses = append(responses, resTest.Tags...)
		}
	}
	if genericProductsDTO.Skill != nil && *genericProductsDTO.Skill == true {
		testProductDTO := request.GenericProductsDTO{CountryId: genericProductsDTO.CountryId, Locale: genericProductsDTO.Locale}
		resSkill, err := t.getTagsByRpc(domain.TagTypeEnum.Skill, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.K12Skill, &testProductDTO)
		if err != nil {
			entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
			return
		}
		if len(resSkill.Tags) > 0 {
			responses = append(responses, resSkill.Tags...)
		}
	}
	err = new(entity.Response).SendResponse(rw, responses, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) getUniversityProducts(rw http.ResponseWriter, req *http.Request) {
	var tag request.GetUniversityProductsDTO

	err := json.NewDecoder(req.Body).Decode(&tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.ErrInvalidRequest, req.Header.Get("locale"), false)
		return
	}
	err = helper.Validate(tag)
	if err != nil {
		entity.HandleError(rw, "badRequest", noonerror.New(noonerror.ErrInvalidRequest, err.Error()), req.Header.Get("locale"), false)
		return
	}
	var genericProductsDTO request.GenericProductsDTO
		if err = copier.Copy(&genericProductsDTO, &tag); err != nil {
		entity.HandleError(rw, "", noonerror.New(noonerror.ErrInternalServer, "mapperError"), req.Header.Get("locale"), false)
	}
	var responses []*response.RpcTagResponseDTO
	res, err := t.getTagsByRpc(domain.TagTypeEnum.Course, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.University, &genericProductsDTO)
	if err != nil {
		entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
		return
	}
	if len(res.Tags) > 0 {
		responses = append(responses, res.Tags...)
	}
	if genericProductsDTO.Test != nil && *genericProductsDTO.Test == true {
		testProductDTO := request.GenericProductsDTO{CountryId: genericProductsDTO.CountryId, Locale: genericProductsDTO.Locale}
		resTest, err := t.getTagsByRpc(domain.TagTypeEnum.Test, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.UniversityTestPrep, &testProductDTO)
		if err != nil {
			entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
			return
		}
		if len(resTest.Tags) > 0 {
			responses = append(responses, resTest.Tags...)
		}
	}
	if genericProductsDTO.Skill != nil && *genericProductsDTO.Skill == true {
		testProductDTO := request.GenericProductsDTO{CountryId: genericProductsDTO.CountryId, Locale: genericProductsDTO.Locale}
		resSkill, err := t.getTagsByRpc(domain.TagTypeEnum.Skill, domain.TagGroupEnum.Curriculum, domain.CurriculumTypeEnum.UniversitySkill, &testProductDTO)
		if err != nil {
			entity.HandleError(rw, "", err, req.Header.Get("locale"), false)
			return
		}
		if len(resSkill.Tags) > 0 {
			responses = append(responses, resSkill.Tags...)
		}
	}
	err = new(entity.Response).SendResponse(rw, responses, nil, http.StatusOK)
	if err != nil {
		entity.HandleError(rw, "internalServerError", noonerror.ErrInternalServer, req.Header.Get("locale"), false)
		return
	}
}

func (t *RpcTagsResource) getTagsByRpc(tagType string, tagGroup string, curriculumType string, genericDTO *request.GenericProductsDTO) (*response.GetRpcTagsResponseDTO, error) {

	startInt := 0
	limitInt := 100

	hierarchySlice := []*string{genericDTO.CountryId, genericDTO.BoardId, genericDTO.GradeId, genericDTO.DegreeId, genericDTO.MajorId}
	var hierarchies []*string
	if len(hierarchySlice) > 0 {
		for _, v := range hierarchySlice {
			if v != nil {
				hierarchies = append(hierarchies, v)
			}
		}
	}
	tag := request.GetTagsByHierarchyDTO{
		Type:           &tagType,
		CurriculumType: &curriculumType,
		TagGroup:       &tagGroup,
		Hierarchy:      hierarchies,
		Locale:         genericDTO.Locale,
		CountryId:      genericDTO.CountryId,
		Start:          startInt,
		Limit:          limitInt,
	}

	err := helper.Validate(tag)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, err.Error())
	}
	var getRpcTags domain.GetRpcTags
	if err = copier.Copy(&getRpcTags, &tag); err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "mapperError")
	}
	res, err := t.rts.GetRpcTags(&getRpcTags)
	if err != nil {
		return nil, err
	}
	var getRpcTagsResponse response.GetRpcTagsResponseDTO
	if err = copier.Copy(&getRpcTagsResponse, &res); err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "mapperError")
	}
	for _, v := range getRpcTagsResponse.Tags {
		if v == nil {
			continue
		}
		if v.LocaleName != nil {
			v.Name = v.LocaleName
		}
		v.LocaleName = nil
	}
	return &getRpcTagsResponse, nil
}
