package service

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	dtomapper "bitbucket.org/noon-micro/curriculum/pkg/service/mapper"
	"fmt"
)

type StudentTagsServiceStruct struct {
	ts  domain.TagsService
	es  domain.Elastic
	geo domain.GeoIp
}

func NewStudentTagsService(ts domain.TagsService, es domain.Elastic, geo domain.GeoIp) *StudentTagsServiceStruct {
	return &StudentTagsServiceStruct{ts: ts, es: es, geo: geo}
}

func (t *StudentTagsServiceStruct) GetCountries(tags *domain.GetCountries) (getTagResponse *domain.GetCountriesResponse, err error) {
	switch *tags.TagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.getCurriculumTags(tags)
	}
	return
}

func (t *StudentTagsServiceStruct) GetCountriesNew(tags *domain.GetCountriesNew) (getTagResponse *domain.GetCountriesNewResponse, err error) {
	return t.getCountriesTagsNew(tags)
}

func (t *StudentTagsServiceStruct) GetBoards(request *domain.GetTags) (getBoardsResponse *domain.GetBoardsResponse, err error) {
	var tagData []*domain.Tags
	tagData, next, err := t.getTagsHandler(request)
	fmt.Println(next)
	if err != nil {
		return nil, err
	}

	tagType := "degree"
	hasCollege, err := t.ts.IsCollegePresent(&tagType, request.Hierarchy)
	getBoardsResponse, _ = dtomapper.CreateGetBoardsResponse(tagData, &hasCollege)
	return getBoardsResponse, nil
}
func (t *StudentTagsServiceStruct) GetGrades(request *domain.GetTags) (getGradesResponse *domain.GetGradesResponse, err error) {
	var tagData []*domain.Tags
	tagData, next, err := t.getTagsHandler(request)
	fmt.Println(next)
	if err != nil {
		return nil, err
	}

	tagType := "degree"
	hasCollege, err := t.ts.IsCollegePresent(&tagType, request.Hierarchy)
	getGradesResponse, _ = dtomapper.CreateGetGradesResponse(tagData,&hasCollege)
	return getGradesResponse, nil
}

func (t *StudentTagsServiceStruct) GetDegrees(request *domain.GetTags) (getGradesResponse *domain.GetDegreesResponse, err error) {
	var tagData []*domain.Tags
	tagData, next, err := t.getTagsHandler(request)
	if err != nil {
		return nil, err
	}
	getGradesResponse, _ = dtomapper.CreateGetDegreesResponse(tagData, next)
	return getGradesResponse, nil
}

func (t *StudentTagsServiceStruct) GetMajors(request *domain.GetTags) (getMajorsResponse *domain.GetMajorsResponse, err error) {
	var tagData []*domain.Tags
	tagData, next, err := t.getTagsHandler(request)
	if err != nil {
		return nil, err
	}
	getMajorsResponse, _ = dtomapper.CreateGetMajorsResponse(tagData, next)
	return getMajorsResponse, nil
}

func (t *StudentTagsServiceStruct) getCurriculumTags(tags *domain.GetCountries) (getTagResponse *domain.GetCountriesResponse, err error) {
	var filteredTags []*string
	var parents []*string
	if tags.Hierarchy != nil {
		parents = append(parents, tags.Hierarchy)
	}
	createElasticEntity, err := dtomapper.GetElasticTagEntity(tags, nil, parents, domain.AccessEnum.Global, tags.CurriculumType, "admin", tags.TagGroup, 0, 100)
	if err != nil {
		return
	}
	tagIds, next, err := t.es.GetTags(createElasticEntity)
	if err != nil {
		parentTagMappingData, err := t.ts.FetchFilteredParentTagMappings(tags.Type, tags.Hierarchy)
		if err != nil {
			return nil, err
		}
		set := make(map[string]interface{})
		for _, v := range parentTagMappingData {
			set[*v.TagID] = nil
		}
		for k := range set {
			key := k
			filteredTags = append(filteredTags, &key)
		}
	} else {
		filteredTags = tagIds
	}
	hiddenSet := make(map[string]bool)
	parentTagMappingData, err := t.ts.FetchByInParentTagMappings(filteredTags)
	if err != nil {
		return
	}
	for _, v := range parentTagMappingData {
		if *v.ParentTagID == *tags.Hierarchy {
			hiddenSet[*v.TagID] = v.Hidden
		}
	}
	tagData, err := t.ts.FetchByInTags(filteredTags)
	if err != nil {
		return
	}
	getTagResponse, _ = dtomapper.CreateGetCountriesResponse(tagData, tags, hiddenSet, next)
	return getTagResponse, nil
}

func (t *StudentTagsServiceStruct) getCountriesTagsNew(tags *domain.GetCountriesNew) (getTagResponse *domain.GetCountriesNewResponse, err error) {
	rootCurriculumType := "root"
	countryType := "country"
	filteredTagsResponse, err := t.ts.FetchFilteredTagsPaginated(&rootCurriculumType, &countryType, &tags.Start, &tags.Limit)
	if err != nil {
		return
	}
	next := tags.Start + tags.Limit
	getTagResponse, _ = dtomapper.CreateGetCountriesNewResponse(filteredTagsResponse, tags.Locale, &tags.ISOCode, &next, false)
	return getTagResponse, nil
}

func (t *StudentTagsServiceStruct) getTagsHandler(tags *domain.GetTags) (tagsResponse []*domain.Tags, next *int, err error) {
	var filteredTags []*string
	var parents []*string
	if tags.Hierarchy != nil {
		parents = append(parents, tags.Hierarchy)
	}
	var nextVal *int
	fmt.Println(nextVal)
	if tags.Text == nil || *tags.Text == "" {
		createElasticEntity, err1 := dtomapper.GetElasticTagEntity(tags, parents, nil, domain.AccessEnum.Global, tags.CurriculumType, "admin", tags.TagGroup, tags.Start, tags.Limit)
		if err1 != nil {
			return
		}
		filteredTags, next, err = t.es.GetTags(createElasticEntity)
	} else {
		createElasticEntity, err1 := dtomapper.GetElasticTagEntity(tags, parents, nil, domain.AccessEnum.Global, tags.CurriculumType, "admin", tags.TagGroup, tags.Start, tags.Limit)
		if err1 != nil {
			return
		}
		filteredTags, next, err = t.es.GetTagsSearch(createElasticEntity)
	}
	hiddenSet := make(map[string]bool)
	parentTagMappingData, err := t.ts.FetchByInParentTagMappingsByParentTagIdTagIds(filteredTags, tags.Hierarchy)
	if err != nil {
		return
	}
	for _, v := range parentTagMappingData {
		if *v.ParentTagID == *tags.Hierarchy {
			hiddenSet[*v.TagID] = v.Hidden
		}
	}
	tagData, err := t.ts.FetchByInTags(filteredTags)
	if err != nil {
		return
	}
	tagData, err = t.ts.FetchTagLocaleMappingsByLocaleForContext(tagData, tags.CountryId, tags.Locale)
	if err != nil {
		return
	}
	tagData, err = t.ts.OrderTags(tagData, tags.Type, tags.CurriculumType, tags.Hierarchy)
	if err != nil {
		return
	}
	return tagData, next, nil
}
