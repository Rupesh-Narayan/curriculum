package service

import (
	"bitbucket.org/noon-micro/curriculum/config"
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	noonerror "bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/flow"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	repository "bitbucket.org/noon-micro/curriculum/pkg/repository/mysql"
	redisrepo "bitbucket.org/noon-micro/curriculum/pkg/repository/redis"
	"bitbucket.org/noon-micro/curriculum/pkg/service/constant"
	dtomapper "bitbucket.org/noon-micro/curriculum/pkg/service/mapper"
	"context"
	"encoding/json"
	"github.com/jinzhu/copier"
	"strconv"
	"sync"
	"time"
)

type RpcTagsServiceStruct struct {
	ts domain.TagsService
	es domain.Elastic
}

func NewRpcTagsService(ts domain.TagsService, es domain.Elastic) *RpcTagsServiceStruct {
	return &RpcTagsServiceStruct{ts: ts, es: es}
}

func (t *RpcTagsServiceStruct) CreateTags(tags *domain.CreateMultipleTags) (tagResponses []*domain.TagResponse, err error) {
	if len(tags.Hierarchy) == 0 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "hierarchyAbsent")
	}
	tagHierarchySlice, err := t.ts.FetchByInTags(tags.Hierarchy)
	if err != nil {
		return
	}
	if *tags.CurriculumType == domain.CurriculumTypeEnum.Default {
		tags.CurriculumType = getCurriculumTypeFromParent(tagHierarchySlice)
	}
	if len(tagHierarchySlice) != len(tags.Hierarchy) {
		return nil, noonerror.New(noonerror.ErrBadRequest, "parentTagMissing")
	}
	var identifierSlice []*domain.Tags
	if len(tags.Identifier) > 0 {
		identifierSlice, err = t.ts.FetchByInTags(tags.Identifier)
	}
	curriculumHierarchy, err := flow.GetCurriculum(tags.CurriculumType)
	if err != nil {
		return
	}
	_, ok := curriculumHierarchy[*tags.Type]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "invalidContent")
	}
	parentTags, parentHideOrderTags, parentIdentifierTagIds, err := verifyAndFetchParentCurriculumTagsForContent(tags.CurriculumType, tags.Type, tagHierarchySlice, constant.WriteAccessType)
	if err != nil {
		return
	}
	positionMap := make(map[int]*domain.TagResponse)
	errChan := make(chan error)
	wgDone := make(chan bool)
	type response struct {
		tagResponse *domain.TagResponse
		position    int
	}
	tagDataChan := make(chan *response, len(tags.Tags))
	var wg sync.WaitGroup
	for i, val := range tags.Tags {
		wg.Add(1)
		go func(val *domain.CreateTag, position int) {
			defer func() {
				if err := recover(); err != nil {
					logger.Client.Error("createTagPanicked", logger.GetErrorStack())
					errChan <- noonerror.New(noonerror.ErrInternalServer, "createTagPanicked")
				}
			}()
			tagData, err := t.createTag(val, tags, parentTags, parentHideOrderTags, parentIdentifierTagIds, identifierSlice)
			if err != nil {
				errChan <- err
			}
			tagDataChan <- &response{tagData, position}
			wg.Done()
		}(val, i)
	}
	go func() {
		wg.Wait()
		close(wgDone)
	}()
	select {
	case <-wgDone:
		for range tags.Tags {
			tagData := <-tagDataChan
			positionMap[tagData.position] = tagData.tagResponse
		}
		break
	case err = <-errChan:
		return
	}
	for i := range tags.Tags {
		_, ok := positionMap[i]
		if !ok {
			return nil, noonerror.New(noonerror.ErrInternalServer, "createTagError")
		}
		tagResponses = append(tagResponses, positionMap[i])
	}
	return tagResponses, nil
}

func (t *RpcTagsServiceStruct) GetTagsByIds(tags *domain.GetTagsByIds, locale bool) (tagResponses []*domain.TagResponse, err error) {

	errChan := make(chan error)
	wgDone := make(chan bool)
	tagDataChan := make(chan *domain.TagResponse, len(tags.TagIds))
	var wg sync.WaitGroup

	for _, v := range tags.TagIds {
		wg.Add(1)
		go func(v *string) {
			if v == nil {
				wg.Done()
				return
			}
			defer func() {
				if err := recover(); err != nil {
					logger.Client.Error("getTags:id:"+*v, logger.GetErrorStack())
					errChan <- noonerror.New(noonerror.ErrInternalServer, "getTagsPanicked")
				}
			}()
			tagData, err := t.ts.FetchTags(v)
			if err != nil {
				errChan <- err
				return
			}
			if tagData == nil {
				wg.Done()
				return
			}
			tagDataSlice, err := t.ts.FetchTagLocaleMappingsByLocale([]*domain.Tags{tagData}, tags.CountryId, tags.Locale)
			if err != nil {
				errChan <- err
				return
			}
			tagData = tagDataSlice[0]
			tagResponse := new(domain.TagResponse)
			var tagLocaleData []*domain.TagLocaleMapping
			tagResponse.ID = v
			tagResponse.Type = tagData.Type
			tagResponse.Name = tagData.Name
			if tagData.LocaleName != nil {
				tagResponse.Name = tagData.LocaleName
			}
			if *tagData.Type == domain.TagTypeEnum.Grade {
				for k, v := range constant.GradeTagMap {
					if *tagData.ID == v {
						grade, _ := strconv.Atoi(k)
						tagResponse.Grade = &grade
					}
					if tagResponse.Grade == nil {
						defaultGrade := constant.DefaultGrade
						tagResponse.Grade = &defaultGrade
					}
				}
			}
			//for board tag data type
			if *tagData.Type == domain.TagTypeEnum.Board && *tagData.ID == config.GetConfig().BoardTagId {
				var boardAttributes = make(map[string]interface{})
				boardAttributes["is_default"]=true
				tagData.Attributes=boardAttributes
			}
			if *tagData.Type == domain.TagTypeEnum.Country && tags.Locale != nil && *tags.Locale == constant.DefaultLocale {
				fullNameInterface, ok := tagData.Attributes["full_name"]
				if ok {
					fullName, okAssertion := fullNameInterface.(string)
					if okAssertion {
						*tagData.Name = fullName
					}
				}
			}

			tagResponse.CurriculumType = &tagData.CurriculumType
			tagResponse.Attributes = tagData.Attributes
			if tagData.LocaleAvailable && locale {
				tagLocaleData, err = t.ts.FetchTagLocaleMappings(v)
				if err != nil {
					errChan <- err
					return
				}
				for _, val := range tagLocaleData {
					var locale domain.LocaleResponse
					locale.Locale = val.Locale
					locale.Name = val.Name
					locale.CountryId = val.CountryId
					tagResponse.Locale = append(tagResponse.Locale, &locale)
				}
			}
			tagDataChan <- tagResponse
			wg.Done()
		}(v)
	}
	go func() {
		wg.Wait()
		close(tagDataChan)
		close(wgDone)
	}()
	select {
	case <-wgDone:
		for v := range tagDataChan {
			tagResponses = append(tagResponses, v)
		}
	case err = <-errChan:
		return nil, err
	}
	return tagResponses, nil
}

func (t *RpcTagsServiceStruct) ValidateHierarchy(validateHierarchy *domain.ValidateHierarchy) (err error) {
	tagIdMap := make(map[string]*domain.Tags)
	var allTagIds []*string
	for _, v := range validateHierarchy.Hierarchies {
		if len(v) == 0 {
			return noonerror.New(noonerror.ErrBadRequest, "hierarchyInvalid")
		}
		for _, id := range v {
			_, ok := tagIdMap[*id]
			if !ok {
				tagIdMap[*id] = nil
				allTagIds = append(allTagIds, id)
			}
		}
	}
	allTags, err := t.ts.FetchByInTags(allTagIds)
	if err != nil {
		return err
	}
	for _, v := range allTags {
		if v == nil {
			return noonerror.New(noonerror.ErrBadRequest, "hierarchyInvalid")
		}
		_, ok := tagIdMap[*v.ID]
		if !ok {
			return noonerror.New(noonerror.ErrBadRequest, "hierarchyInvalid")
		}
		tagIdMap[*v.ID] = v
	}
	errChan := make(chan error)
	wgDone := make(chan bool)
	var wg sync.WaitGroup
	for _, v := range validateHierarchy.Hierarchies {
		wg.Add(1)
		go func(v []*string) {
			defer func() {
				if err := recover(); err != nil {
					logger.Client.Error("validateHierarchyPanicked", v, logger.GetErrorStack())
					errChan <- noonerror.New(noonerror.ErrInternalServer, "validateHierarchyPanicked")
				}
			}()
			err := t.validateHierarchySingle(v, validateHierarchy.CurriculumType, tagIdMap)
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}(v)
	}
	go func() {
		wg.Wait()
		close(wgDone)
	}()
	select {
	case <-wgDone:
		return
	case err = <-errChan:
		return err
	}
}

func (t *RpcTagsServiceStruct) validateHierarchySingle(hierarchy []*string, curriculumType *string, tagMap map[string]*domain.Tags) (err error) {
	if curriculumType == nil {
		return noonerror.New(noonerror.ErrBadRequest, "hierarchyInvalid")
	}
	var tagHierarchySlice []*domain.Tags
	for _, v := range hierarchy {
		tag, ok := tagMap[*v]
		if !ok {
			return noonerror.New(noonerror.ErrBadRequest, "hierarchyInvalid")
		}
		tagHierarchySlice = append(tagHierarchySlice, tag)
	}
	if *curriculumType == domain.CurriculumTypeEnum.Default {
		curriculumType = getCurriculumTypeFromParent(tagHierarchySlice)
	}
	curriculumHierarchy, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return err
	}
	level := 0
	var tagType string
	var tagId string
	for _, val := range tagHierarchySlice {
		_, ok := curriculumHierarchy[*val.Type]
		if !ok {
			return noonerror.New(noonerror.ErrBadRequest, "parentTagTypeInvalid")
		}
		if curriculumHierarchy[*val.Type].Level > level {
			level = curriculumHierarchy[*val.Type].Level
			tagType = *val.Type
			tagId = *val.ID
		}
	}
	var filteredTags []*domain.Tags
	for _, val := range tagHierarchySlice {
		if *val.ID != tagId {
			filteredTags = append(filteredTags, val)
		}
	}
	if len(filteredTags) > 0 {
		_, parentHideOrderTags, _, err := verifyAndFetchParentCurriculumTagsForContent(curriculumType, &tagType, filteredTags, constant.ReadAccessType)
		if err != nil {
			return err
		}
		parentTagMapping, err := t.ts.FetchParentTagMappingByParentTagIdTagId(&tagId, parentHideOrderTags)
		if err != nil {
			return err
		}
		if parentTagMapping == nil {
			return noonerror.New(noonerror.ErrBadRequest, "hierarchyInvalid")
		}
		if parentTagMapping.Hidden {
			return noonerror.New(noonerror.ErrBadRequest, "parentHidden")
		}
	}
	return
}

func (t *RpcTagsServiceStruct) createTag(createTag *domain.CreateTag, tags *domain.CreateMultipleTags, parentTags *string, parentHideOrderTags *string, parentIdentifierTagIds map[string]string, identifierSlice []*domain.Tags) (tagResponse *domain.TagResponse, err error) {
	ctx := context.Background()
	tx, err := repository.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "ContextCreationError")
	}
	if createTag == nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagDataMissing")
	}
	var createTags domain.CreateTags
	if err = copier.Copy(&createTags, tags); err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "mapperError")
	}
	if len(createTag.CountryId) == 0 {
		createTag.CountryId = "0"
	}
	createTags.Name = createTag.Name
	createTags.CountryId = createTag.CountryId
	hierarchyType := constant.HierarchyCurriculum
	mappedCurriculumType, err := flow.CurriculumMapper(tags.CurriculumType)
	if err != nil {
		return nil, err
	}
	tagId, err := t.ts.CreateTags(tx, &domain.Tags{Type: tags.Type, Name: createTag.Name, CurriculumType: *mappedCurriculumType,
		CreatorId: tags.CreatorId, CreatorType: *tags.CreatorType, Access: domain.AccessEnum.Teacher, TagGroup: *tags.TagGroup, LocaleAvailable: false, CountryId: createTag.CountryId,
		Attributes: tags.Attributes, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	if err != nil {
		_ = tx.Rollback()
		return
	}
	hierarchyCurriculumType := isRootCurriculum(tags.CurriculumType)
	parentIdMap := map[string]*string{}
	parentIdMap[hierarchyCurriculumType] = parentTags
	parentIdMap[hierarchyType] = parentHideOrderTags
	for k, v := range parentIdentifierTagIds {
		val := v
		parentIdMap[k] = &val
	}
	if len(identifierSlice) > 0 {
		for _, v := range identifierSlice {
			continueInner := false
			for _, vInner := range parentIdentifierTagIds {
				if vInner == *v.ID {
					continueInner = true
					break
				}
			}
			if continueInner {
				continue
			}
			parentTagId := *v.ID
			parentIdMap[*v.Type] = &parentTagId
		}
	}
	var allParentTags []*string
	for _, v := range parentIdMap {
		allParentTags = append(allParentTags, v)
	}
	rollback := func() {
		deleteTag := false
		go func() {
			_ = t.es.UpdateTag(tagId, &deleteTag, nil)
		}()
	}
	createElasticEntity, err := dtomapper.CreateElasticTagEntity(tagId, &createTags, allParentTags, domain.AccessEnum.Teacher)
	if err != nil {
		return
	}
	err = t.es.CreateTag(createElasticEntity)
	if err != nil {
		rollback()
		_ = tx.Rollback()
		return
	}
	for k, v := range parentIdMap {
		key := k
		order := 0
		err = t.ts.CreateParentTagMapping(tx, &domain.ParentTagMapping{TagID: tagId, TagType: tags.Type, ParentTagType: &key, ParentTagID: v, Order: &order, Hidden: false, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()})
		if err != nil {
			rollback()
			_ = tx.Rollback()
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		rollback()
		return
	}
	stringTagId := *tagId
	return &domain.TagResponse{
		ID:   &stringTagId,
		Type: tags.Type,
		Name: createTag.Name,
	}, nil
}

func (t *RpcTagsServiceStruct) GetTags(tags *domain.GetTags) (getTagResponse *domain.GetTagsResponse, err error) {
	switch *tags.TagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.getCurriculumTags(tags)
	}
	return
}

func (t *RpcTagsServiceStruct) GetDefaultTags() (*domain.DefaultTags, error) {
	miscTagId := config.GetConfig().MiscTagId
	resourceTagid := config.GetConfig().ResourceTagId
	defaultTagIds := []*string{&miscTagId, &resourceTagid}

	tagData, err := t.ts.FetchByInTags(defaultTagIds)
	if err != nil {
		return nil, err
	}
	if tagData == nil || len(tagData) != 2 {
		err = noonerror.New(noonerror.ErrInternalServer, "parentTagTypeInvalid")
		return nil, err
	}
	dc, _ := dtomapper.CreateDefaultTagResponse(tagData)
	return dc, nil
}

func (t *RpcTagsServiceStruct) getCurriculumTags(tags *domain.GetTags) (getTagResponse *domain.GetTagsResponse, err error) {
	parents := []*string{tags.Hierarchy}
	createElasticEntity, err := dtomapper.GetElasticTagEntity(tags, parents, nil, domain.AccessEnum.Global, tags.CurriculumType, "admin", tags.TagGroup, 0, 100)
	if err != nil {
		return
	}
	filteredTags, next, err := t.es.GetTags(createElasticEntity)
	tagData, err := t.ts.FetchByInTags(filteredTags)
	if err != nil {
		return
	}
	getTagResponse, _ = dtomapper.CreateGetTagResponse(tagData, tags, map[string]bool{}, next)
	return getTagResponse, nil
}

func (t *RpcTagsServiceStruct) GetSuggestedCurriculum(getSuggestedTags *domain.GetSuggestedTags) ([]*domain.SuggestedTags, error) {

	tagHierarchySlice, err := t.ts.FetchByInTags(getSuggestedTags.TagIds)
	if err != nil {
		return nil, err
	}
	if *getSuggestedTags.CurriculumType == domain.CurriculumTypeEnum.Default {
		getSuggestedTags.CurriculumType = getCurriculumTypeFromParent(tagHierarchySlice)
	}
	_, err = flow.GetCurriculum(getSuggestedTags.CurriculumType)
	if err != nil {
		return nil, err
	}
	tagHierarchyMap := make(map[string]*domain.Tags, len(tagHierarchySlice))
	for _, item := range tagHierarchySlice {
		tagHierarchyMap[*item.ID] = item
	}

	contentType := domain.TagTypeEnum.Chapter

	_, parentHideOrderTags, _, err := verifyAndFetchParentCurriculumTagsForContent(getSuggestedTags.CurriculumType, &contentType, tagHierarchySlice, constant.ReadAccessType)
	if err != nil {
		return nil, err
	}
	chapterTags, err := t.getChildTags(getSuggestedTags, parentHideOrderTags, &contentType)

	if err != nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "invalidChapter")
	}
	if chapterTags == nil || len(chapterTags) == 0 {
		return []*domain.SuggestedTags{}, nil
	}

	chapterIdToTagMap := make(map[string][]*domain.SuggestedTags)
	errChan := make(chan error)
	wgDone := make(chan bool)
	tagDataChan := make(chan map[string][]*domain.SuggestedTags, len(chapterTags))
	var wg sync.WaitGroup

	for _, v := range chapterTags {
		wg.Add(1)
		go func(v *domain.SuggestedTags) {
			defer func() {
				if err := recover(); err != nil {
					logger.Client.Error("getChildTags:id:"+*v.ID, logger.GetErrorStack())
					errChan <- noonerror.New(noonerror.ErrInternalServer, "getTagsPanicked")
				}
			}()
			chapterHideOrderTags := *parentHideOrderTags + "." + *v.ID
			topicContentType := domain.TagTypeEnum.Topic
			topicSuggestedTags, err := t.getChildTags(getSuggestedTags, &chapterHideOrderTags, &topicContentType)
			if err != nil {
				errChan <- err
			}
			idToTagMap := make(map[string][]*domain.SuggestedTags)
			idToTagMap[*v.ID] = topicSuggestedTags
			tagDataChan <- idToTagMap
			wg.Done()
		}(v)
	}
	go func() {
		wg.Wait()
		close(tagDataChan)
		close(wgDone)
	}()
	select {
	case <-wgDone:
		for range chapterTags {
			chanelData := <-tagDataChan
			for k, v := range chanelData {
				chapterIdToTagMap[k] = v
			}
		}
	case err = <-errChan:
		return nil, err
	}

	for _, v := range chapterTags {
		v.Topics = chapterIdToTagMap[*v.ID]
	}

	return chapterTags, nil
}

func (t *RpcTagsServiceStruct) GetLegacyDataFromTagId(tagId *string) (legacyResponse []*domain.LegacyResponse, err error) {
	legacyTagMappings, err := t.ts.FetchLegacyIdFromTagId(tagId)
	if err != nil {
		return
	}
	for _, v := range legacyTagMappings {
		legacyResponse = append(legacyResponse, &domain.LegacyResponse{ID: *v.LegacyId, Type: *v.LegacyIdType})
	}
	return legacyResponse, nil
}

func (t *RpcTagsServiceStruct) GetLegacyDataFromTagIds(tagIds []*string) (legacyResponse []*domain.LegacyResponse, err error) {
	visited := make(map[string]struct{})
	var nonDuplicatedTagIds []*string
	for _, v := range tagIds {
		_, ok := visited[*v]
		if !ok {
			nonDuplicatedTagIds = append(nonDuplicatedTagIds, v)
			visited[*v] = struct{}{}
		}
	}
	legacyTagMappings, err := t.ts.FetchLegacyIdFromTagIds(nonDuplicatedTagIds)
	if err != nil {
		return
	}
	for _, v := range legacyTagMappings {
		legacyResponse = append(legacyResponse, &domain.LegacyResponse{TagId: *v.TagID, ID: *v.LegacyId, Type: *v.LegacyIdType})
	}
	return legacyResponse, nil
}

func (t *RpcTagsServiceStruct) GetTagDataFromLegacyId(legacyType *string, legacyId *string) (legacyResponse []*domain.LegacyResponse, err error) {
	if *legacyType == "product" {
		_, ok := constant.UniversityProductsMap[*legacyId]
		if ok {
			configuration := config.GetConfig()
			return []*domain.LegacyResponse{
				{ID: configuration.DegreeTagId, Type: domain.TagTypeEnum.Degree},
				{ID: configuration.MajorTagId, Type: domain.TagTypeEnum.Major},
				{ID: configuration.CourseTagId, Type: domain.TagTypeEnum.Course},
				{ID: configuration.UniversitySectionTagId, Type: domain.TagTypeEnum.Section},
			}, nil
		}
	}
	legacyTagMappings, err := t.ts.FetchTagIdFromLegacyId(legacyType, legacyId)
	if err != nil {
		return
	}
	for _, v := range legacyTagMappings {
		legacyResponse = append(legacyResponse, &domain.LegacyResponse{ID: *v.TagID, Type: *v.TagIdType})
	}
	return legacyResponse, nil
}

func (t *RpcTagsServiceStruct) GetGradeTags(grade *string, productId *string) (legacyResponse []*domain.LegacyResponse, err error) {
	_, ok := constant.UniversityProductsMap[*productId]
	if *grade == "13" && ok {
		return
	} else if *grade == "0" && ok {
		return
	}
	gradeType := "grade"
	gradeProducts, err := t.ts.FetchGradesFromProductId(productId)
	if err != nil {
		return
	}
	if *grade == "13" && !ok {
		highestGrade := 0
		var folderId *string
		for _, v := range gradeProducts {
			if v.Grade != nil {
				grade, _ := strconv.Atoi(*v.Grade)
				if grade > highestGrade {
					highestGrade = grade
					folderId = v.FolderId
				}
			}
		}
		if folderId != nil {
			legacyResponse, err = t.legacyGradeResponse(folderId, legacyResponse)
			if err != nil {
				return nil, err
			}
		} else {
			gradeType := domain.TagTypeEnum.Grade
			gradeId := constant.GradeTagMap["12"]
			legacyResponse = append(legacyResponse, &domain.LegacyResponse{ID: gradeId, Type: gradeType})
		}
	} else if *grade == "0" {
		for _, v := range gradeProducts {
			if v.FolderId != nil {
				legacyResponse, err = t.legacyGradeResponse(v.FolderId, legacyResponse)
				if err != nil {
					return nil, err
				}
			} else if v.Grade != nil {
				gradeId, ok := constant.GradeTagMap[*v.Grade]
				if ok {
					legacyResponse = append(legacyResponse, &domain.LegacyResponse{ID: gradeId, Type: gradeType})
				}
			}
		}
	} else if len(gradeProducts) > 0 {
		for _, v := range gradeProducts {
			if v.FolderId != nil && v.Grade != nil && *v.Grade == *grade {
				legacyResponse, err = t.legacyGradeResponse(v.FolderId, legacyResponse)
				if err != nil {
					return nil, err
				}
			} else if v.Grade != nil && *v.Grade == *grade {
				gradeId, ok := constant.GradeTagMap[*v.Grade]
				if ok {
					legacyResponse = append(legacyResponse, &domain.LegacyResponse{ID: gradeId, Type: gradeType})
				}
			}
		}
		if len(legacyResponse) == 0 {
			gradeId, ok := constant.GradeTagMap[*grade]
			if ok {
				legacyResponse = append(legacyResponse, &domain.LegacyResponse{ID: gradeId, Type: gradeType})
			}
		}
	} else {
		gradeId, ok := constant.GradeTagMap[*grade]
		if ok {
			legacyResponse = append(legacyResponse, &domain.LegacyResponse{ID: gradeId, Type: gradeType})
		}
	}
	return legacyResponse, nil
}

func (t *RpcTagsServiceStruct) legacyGradeResponse(folderId *string, legacyResponseInput []*domain.LegacyResponse) (legacyResponse []*domain.LegacyResponse, err error) {
	folderType := "folder"
	legacyTagMappings, err := t.ts.FetchTagIdFromLegacyId(&folderType, folderId)
	if err != nil {
		return nil, err
	}
	for _, v := range legacyTagMappings {
		legacyResponseInput = append(legacyResponseInput, &domain.LegacyResponse{ID: *v.TagID, Type: *v.TagIdType})
	}
	return legacyResponseInput, nil
}

func (t *RpcTagsServiceStruct) getChildTags(getSuggestedTags *domain.GetSuggestedTags, parentTags *string, contentType *string) ([]*domain.SuggestedTags, error) {

	tagGroup := domain.TagGroupEnum.Content
	getTag := domain.GetTags{TagGroup: &tagGroup, Type: contentType}

	parents := []*string{parentTags}
	createElasticEntity, err := dtomapper.GetElasticTagEntity(getTag, parents, []*string{}, domain.AccessEnum.Global, getSuggestedTags.CurriculumType, "admin", getTag.TagGroup, 0, 100)
	if err != nil {
		return nil, err
	}
	tagIds, _, err := t.es.GetTags(createElasticEntity)
	if err != nil {
		return nil, err
	}
	tagData, err := t.ts.FetchByInTags(tagIds)
	if err != nil {
		return nil, err
	}

	tagData, err = t.ts.OrderTags(tagData, contentType, getSuggestedTags.CurriculumType, parentTags)
	if err != nil {
		return nil, err
	}
	str, _ := dtomapper.SuggestedTagResponse(tagData)
	return str, nil
}

func (t *RpcTagsServiceStruct) GetRpcTags(tags *domain.GetRpcTags) (getTagResponse *domain.GetTagsResponseForProduct, err error) {

	switch *tags.TagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.getCurriculumTagsForProducts(tags)
	}
	return

}

func (t *RpcTagsServiceStruct) getCurriculumTagsForProducts(gtt *domain.GetRpcTags) (*domain.GetTagsResponseForProduct, error) {

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

	var gradeTag *domain.Tags
	var boardTag *domain.Tags
	var gradeTags []*domain.Tags
	for _, v := range tagHierarchySlice {
		if v != nil && *v.Type == domain.TagTypeEnum.Grade {
			gradeTag = v
		}
		if v != nil && *v.Type == domain.TagTypeEnum.Board {
			boardTag = v
		}
	}
	var parentTags []*string
	//multi grade scenario
	if gradeTag != nil {
		gradeTags, err = t.getMultiGrades(*gtt.CountryId, boardTag,gradeTag)
		if err != nil {
			return nil, err
		}
	}
	for _, v := range gradeTags {
		parent := *gtt.CountryId + "." + *v.ID
		if boardTag!=nil {
			parent = *gtt.CountryId + "." + *boardTag.ID+"."+ *v.ID
		}
		parentTags = append(parentTags, &parent)
	}
	if len(gradeTags) == 0 {
		parent, err := verifyAndFetchParentCurriculumTags(gtt.CurriculumType, tagHierarchySlice, tagHierarchy.Level)
		if err != nil {
			return nil, err
		}
		parentTags = append(parentTags, parent)
	}
	if len(parentTags) == 0 {
		return new(domain.GetTagsResponseForProduct), nil
	}
	var filteredTags []*string
	errChan := make(chan error)
	wgDone := make(chan bool)
	tagDataChan := make(chan []*string, len(parentTags))
	var wg sync.WaitGroup
	for _, parent := range parentTags {
		wg.Add(1)
		go func(parent *string) {
			defer func() {
				if err := recover(); err != nil {
					logger.Client.Error("getProductsOfCountryPanicked:parent:"+*parent, logger.GetErrorStack())
					errChan <- noonerror.New(noonerror.ErrInternalServer, "getProductsOfCountryPanicked")
				}
			}()
			getTag := domain.GetTeacherTags{Text: gtt.Text, TagGroup: gtt.TagGroup, Type: gtt.Type}
			parents := []*string{parent}

			createElasticEntity, err := dtomapper.GetElasticTagEntity(getTag, parents, nil, domain.AccessEnum.Global, gtt.CurriculumType, "admin", gtt.TagGroup, gtt.Start, gtt.Limit)
			if err != nil {
				errChan <- err
			}
			ids, _, err := t.es.GetTags(createElasticEntity)
			if err != nil {
				errChan <- err
			}
			tagDataChan <- ids
			wg.Done()
		}(parent)
	}
	go func() {
		wg.Wait()
		close(wgDone)
	}()
	select {
	case <-wgDone:
		tagIdMap := make(map[string]struct{})
		for range parentTags {
			tagIds := <-tagDataChan
			for _, v := range tagIds {
				_, ok := tagIdMap[*v]
				if !ok {
					filteredTags = append(filteredTags, v)
					tagIdMap[*v] = struct{}{}
				}
			}
		}
	case err = <-errChan:
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
	tagData, err = t.ts.OrderTags(tagData, gtt.Type, gtt.CurriculumType, nil)
	if err != nil {
		return nil, err
	}
	getTagResponse, _ := dtomapper.GetTagResponseForProduct(tagData, gtt.Type, gtt.CurriculumType, make(map[string]bool), nil)
	return getTagResponse, nil
}

func (t *RpcTagsServiceStruct) getMultiGrades(countryId string,boardTag *domain.Tags, gradeTag *domain.Tags) (finalTagData []*domain.Tags, err error) {
	_, ok := constant.MultiGradeMap[countryId]
	if !ok {
		return []*domain.Tags{gradeTag}, nil
	}
	tagGroup := domain.TagGroupEnum.Curriculum
	tagType := domain.TagTypeEnum.Grade
	curriculumType := domain.CurriculumTypeEnum.K12

	redisKey := redisrepo.CurriculumMultiGradePrefix + countryId + ":" + *gradeTag.ID
	if boardTag!=nil{
		redisKey = redisrepo.CurriculumMultiGradePrefix + countryId +":"+*boardTag.ID+ ":" + *gradeTag.ID
	}
	val, err := redisrepo.RedisClient.Get(redisKey).Result()
	err1 := json.Unmarshal([]byte(val), &finalTagData)
	if err == nil && err1 == nil {
		return finalTagData, err
	}
	getTag := domain.GetTeacherTags{Text: nil, TagGroup: &tagGroup, Type: &tagType}
	parents := []*string{&countryId}
	if boardTag!=nil{
		var parent = countryId+"."+*boardTag.ID
		parents = []*string{&parent}	}

	createElasticEntity, err := dtomapper.GetElasticTagEntity(getTag, parents, nil, "", &curriculumType, "admin", &tagGroup, 0, 100)
	if err != nil {
		return nil, err
	}
	filteredTags, _, err := t.es.GetTags(createElasticEntity)
	if err != nil {
		return nil, err
	}
	tagData, err := t.ts.FetchByInTags(filteredTags)
	if err != nil {
		return nil, err
	}
	for _, v := range tagData {
		if *v.ID == *gradeTag.ID {
			finalTagData = append(finalTagData, v)
			continue
		}
		if v.Attributes == nil {
			continue
		}
		multiGradeObject, ok := v.Attributes["multi_grade"]
		if !ok {
			continue
		}
		multiGradeArray, ok := multiGradeObject.([]interface{})
		if !ok {
			continue
		}
		for _, multiGrade := range multiGradeArray {
			countryObject, ok := multiGrade.(map[string]interface{})
			if !ok {
				continue
			}
			countryIdMultiGrade, ok := countryObject["country_id"].(string)
			if !ok || countryIdMultiGrade != countryId {
				continue
			}
			gradeTagIds, ok := countryObject["grade_tag_ids"].([]interface{})
			if !ok || gradeTagIds == nil {
				continue
			}
			for _, gradeObject := range gradeTagIds {
				grade, ok := gradeObject.(string)
				if !ok {
					continue
				}
				if grade == *gradeTag.ID {
					finalTagData = append(finalTagData, v)
				}
			}
		}
	}
	if len(finalTagData) > 0 {
		tagByte, err := json.Marshal(finalTagData)
		if err == nil {
			redisrepo.RedisClient.Set(redisKey, string(tagByte), redisrepo.MultiGradeTtl)
		}
	}
	return finalTagData, nil
}
