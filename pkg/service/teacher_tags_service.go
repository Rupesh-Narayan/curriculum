package service

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	noonerror "bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/flow"
	"bitbucket.org/noon-micro/curriculum/pkg/service/constant"
	dtomapper "bitbucket.org/noon-micro/curriculum/pkg/service/mapper"
	"sort"
	"strings"
)

type TeacherTagsServiceStruct struct {
	ts  domain.TagsService
	es  domain.Elastic
	geo domain.GeoIp
}

func NewTeacherTagsService(ts domain.TagsService, es domain.Elastic, geo domain.GeoIp) *TeacherTagsServiceStruct {
	return &TeacherTagsServiceStruct{ts: ts, es: es, geo: geo}
}

func (t *TeacherTagsServiceStruct) GetTeacherTags(tags *domain.GetTeacherTags) (getTagResponse *domain.GetTagsResponse, err error) {

	switch *tags.TagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.getCurriculumTags(tags)
	case domain.TagGroupEnum.Content:
		return t.getContentTags(tags)
	}
	return

}
func (t *TeacherTagsServiceStruct) SearchTeacherTags(_ *domain.GetTeacherTags) ([]*domain.TagResponse, error) {
	return nil, nil
}

func (t *TeacherTagsServiceStruct) getCurriculumTags(gtt *domain.GetTeacherTags) (*domain.GetTagsResponse, error) {

	tagHierarchySlice, err := t.ts.GetTagsConcurrent(gtt.Hierarchy)
	if err != nil {
		return nil, err
	}
	if *gtt.CurriculumType == domain.CurriculumTypeEnum.Default {
		gtt.CurriculumType = getCurriculumTypeFromParent(tagHierarchySlice)
	}
	curriculumHierarchy, err := flow.GetCurriculum(gtt.CurriculumType)
	if err != nil {
		return nil, err
	}

	tagHierarchy, ok := curriculumHierarchy[*gtt.Type]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	if tagHierarchy.Level == 1 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	parentTags, err := verifyAndFetchParentCurriculumTags(gtt.CurriculumType, tagHierarchySlice, tagHierarchy.Level)
	if err != nil {
		return nil, err
	}

	getTag := domain.GetTeacherTags{Text: gtt.Text, TagGroup: gtt.TagGroup, Type: gtt.Type}
	parents := []*string{parentTags}

	createElasticEntity, err := dtomapper.GetElasticTagEntity(getTag, parents, nil, "", gtt.CurriculumType, "admin", gtt.TagGroup, gtt.Start, gtt.Limit)
	if err != nil {
		return nil, err
	}
	filteredTags, next, err := t.es.GetTags(createElasticEntity)
	if err != nil {
		return nil, err
	}
	tagData, err := t.ts.FetchByInTags(filteredTags)
	if err != nil {
		return nil, err
	}
	tagData, err = t.ts.FetchTagLocaleMappingsByLocale(tagData, gtt.CountryId, gtt.Locale)
	if err != nil {
		return nil, err
	}
	tagData, err = t.ts.OrderTags(tagData, gtt.Type, gtt.CurriculumType, parentTags)
	if err != nil {
		return nil, err
	}
	getTagResponse, _ := dtomapper.GetTagResponse(tagData, gtt.Type, gtt.CurriculumType, make(map[string]bool), next)
	return getTagResponse, nil
}

func (t *TeacherTagsServiceStruct) getContentTags(gtt *domain.GetTeacherTags) (*domain.GetTagsResponse, error) {
	tagHierarchySlice, err := t.ts.FetchByInTags(gtt.Hierarchy)
	if err != nil {
		return nil, err
	}
	if *gtt.CurriculumType == domain.CurriculumTypeEnum.Default {
		gtt.CurriculumType = getCurriculumTypeFromParent(tagHierarchySlice)
	}
	_, err = flow.GetCurriculum(gtt.CurriculumType)
	if err != nil {
		return nil, err
	}
	_, parentHideOrderTags, _, err := verifyAndFetchParentCurriculumTagsForContent(gtt.CurriculumType, gtt.Type, tagHierarchySlice, constant.ReadAccessType)
	if err != nil {
		return nil, err
	}

	getTag := domain.GetTeacherTags{Text: gtt.Text, TagGroup: gtt.TagGroup, Type: gtt.Type}
	parents := []*string{parentHideOrderTags}
	adminElasticEntity, err := dtomapper.GetElasticTagEntity(getTag, parents, []*string{}, domain.AccessEnum.Global, gtt.CurriculumType, "admin", getTag.TagGroup, getTag.Start, getTag.Limit)
	if err != nil {
		return nil, err
	}

	userElasticEntity, err := dtomapper.GetElasticTagEntity(gtt, parents, []*string{}, domain.AccessEnum.Teacher, gtt.CurriculumType, "teacher", getTag.TagGroup, getTag.Start, getTag.Limit)
	if err != nil {
		return nil, err
	}
	tagIds, next, err := t.es.GetTags(adminElasticEntity)
	if err != nil {
		return nil, err
	}
	tagData, err := t.ts.FetchByInTags(tagIds)
	if err != nil {
		return nil, err
	}
	tagData, err = t.ts.OrderTags(tagData, gtt.Type, gtt.CurriculumType, parentHideOrderTags)
	if err != nil {
		return nil, err
	}
	userTagIds, _, err := t.es.GetTags(userElasticEntity)
	if err != nil {
		return nil, err
	}

	userTagData, err := t.ts.FetchByInTags(userTagIds)
	if err != nil {
		return nil, err
	}
	if len(userTagData) > 0 {
		tagData = append(tagData, userTagData...)
		sort.SliceStable(tagData, func(i, j int) bool {
			return strings.ToLower(*tagData[i].Name) < strings.ToLower(*tagData[j].Name)
		})
	}

	getTagResponse, _ := dtomapper.GetTagResponse(tagData, gtt.Type, gtt.CurriculumType, make(map[string]bool), next)
	return getTagResponse, nil
}

func (t *TeacherTagsServiceStruct) GetTestsSkillsForLibrary(gtt *domain.GetTeacherTags) (*domain.GetTagsResponse, error) {

	tagHierarchySlice, err := t.ts.GetTagsConcurrent(gtt.Hierarchy)
	if err != nil {
		return nil, err
	}
	var curriculumType *string
	if *gtt.Type == domain.TagTypeEnum.Test {
		curriculumType = new(string)
		*curriculumType = domain.CurriculumTypeEnum.GeneralTestPrep
	}
	if *gtt.Type == domain.TagTypeEnum.Skill {
		curriculumType = new(string)
		*curriculumType = domain.CurriculumTypeEnum.GeneralSkill
	}
	if curriculumType == nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	curriculumHierarchy, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return nil, err
	}

	tagHierarchy, ok := curriculumHierarchy[*gtt.Type]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	if tagHierarchy.Level == 1 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	parentTags, err := verifyAndFetchParentCurriculumTags(curriculumType, tagHierarchySlice, tagHierarchy.Level)
	if err != nil {
		return nil, err
	}

	getTag := domain.GetTeacherTags{Text: gtt.Text, TagGroup: gtt.TagGroup, Type: gtt.Type}
	parents := []*string{parentTags}

	createElasticEntity, err := dtomapper.GetElasticTagEntityWithoutCurriculumType(getTag, parents, nil, "", "admin", gtt.TagGroup, gtt.Start, gtt.Limit)
	if err != nil {
		return nil, err
	}
	filteredTags, next, err := t.es.GetTags(createElasticEntity)
	if err != nil {
		return nil, err
	}
	tagData, err := t.ts.FetchByInTags(filteredTags)
	if err != nil {
		return nil, err
	}
	tagData, err = t.ts.FetchTagLocaleMappingsByLocale(tagData, gtt.CountryId, gtt.Locale)
	if err != nil {
		return nil, err
	}
	tagData, err = t.ts.OrderTags(tagData, gtt.Type, curriculumType, parentTags)
	if err != nil {
		return nil, err
	}
	getTagResponse, _ := dtomapper.GetTagResponse(tagData, gtt.Type, nil, make(map[string]bool), next)
	return getTagResponse, nil
}

func (t *TeacherTagsServiceStruct) GetCountriesTagsNew(tags *domain.GetCountriesNew) (getTagResponse *domain.GetCountriesNewResponse, err error) {
	rootCurriculumType := domain.CurriculumTypeEnum.Root
	countryType := domain.TagTypeEnum.Country
	filteredTagsResponse, err := t.ts.FetchFilteredTagsPaginated(&rootCurriculumType, &countryType, &tags.Start, &tags.Limit)
	if err != nil {
		return
	}
	next := -1
	if len(filteredTagsResponse) >= tags.Limit {
		next = tags.Start + tags.Limit
	}
	getTagResponse, _ = dtomapper.CreateGetCountriesNewResponse(filteredTagsResponse, tags.Locale, &tags.ISOCode, &next, false)
	return getTagResponse, nil
}
