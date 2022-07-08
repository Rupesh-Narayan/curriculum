package service

import (
	"bitbucket.org/noon-micro/curriculum/config"
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	noonerror "bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/flow"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	repository "bitbucket.org/noon-micro/curriculum/pkg/repository/redis"
	"bitbucket.org/noon-micro/curriculum/pkg/service/constant"
	"database/sql"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type TagsServiceStruct struct {
	tr   domain.TagsRepository
	ptmr domain.ParentTagMappingRepository
	tlmr domain.TagLocaleMappingRepository
	ltmr domain.LegacyTagMappingRepository
	gpr  domain.GradeProductRepository
}

func NewTagsService(tr domain.TagsRepository, ptmr domain.ParentTagMappingRepository, tlmr domain.TagLocaleMappingRepository, ltmr domain.LegacyTagMappingRepository, gpr domain.GradeProductRepository) *TagsServiceStruct {
	return &TagsServiceStruct{tr: tr, ptmr: ptmr, tlmr: tlmr, ltmr: ltmr, gpr: gpr}
}

func (t *TagsServiceStruct) FetchTags(id *string) (tag *domain.Tags, err error) {
	var tagData domain.Tags
	redisKey := repository.CurriculumPrefix + *id
	val, err := repository.RedisClient.Get(redisKey).Result()
	err1 := json.Unmarshal([]byte(val), &tagData)
	if err != nil || err1 != nil {
		tag, err := t.tr.FetchTags(id)
		if err != nil {
			return nil, err
		}
		if tag != nil {
			tagByte, err := json.Marshal(*tag)
			if err == nil {
				repository.RedisClient.Set(redisKey, string(tagByte), repository.RedisTtl)
			}
		}
		return tag, nil
	}
	return &tagData, nil
}

func (t *TagsServiceStruct) GetTagsConcurrent(tagIds []*string) (tagData []*domain.Tags, err error) {
	if len(tagIds) == 0 {
		return
	}
	errChan := make(chan error)
	wgDone := make(chan bool)
	tagDataChan := make(chan *domain.Tags, len(tagIds))
	var wg sync.WaitGroup
	for _, tagId := range tagIds {
		wg.Add(1)
		go func(tagId *string) {
			defer func() {
				if err := recover(); err != nil {
					logger.Client.Error("getTagsPanicked:id:"+*tagId, logger.GetErrorStack())
					errChan <- noonerror.New(noonerror.ErrInternalServer, "getTagsPanicked")
				}
			}()
			tagData, err := t.FetchTags(tagId)
			if err != nil {
				errChan <- err
			}
			tagDataChan <- tagData
			wg.Done()
		}(tagId)
	}
	go func() {
		wg.Wait()
		close(wgDone)
	}()
	select {
	case <-wgDone:
		for range tagIds {
			tagData = append(tagData, <-tagDataChan)
		}
	case err = <-errChan:
		return
	}
	return
}

func (t *TagsServiceStruct) FetchFilteredTags(curriculumType *string, tagType *string) (tags []*domain.Tags, err error) {
	return t.tr.FetchFilteredTags(curriculumType, tagType)
}

func (t *TagsServiceStruct) FetchFilteredTagsPaginated(curriculumType *string, tagType *string, start *int, limit *int) (tags []*domain.Tags, err error) {
	if *tagType == domain.TagTypeEnum.Country {
		redisKey := repository.CurriculumCountryPrefix + *curriculumType + ":" + *tagType + ":" + strconv.Itoa(*start) + ":" + strconv.Itoa(*limit)
		val, err := repository.RedisClient.Get(redisKey).Result()
		err1 := json.Unmarshal([]byte(val), &tags)
		if err != nil || err1 != nil {
			tag, err := t.tr.FetchFilteredTagsPaginated(curriculumType, tagType, start, limit)
			if err != nil {
				return nil, err
			}
			if tag != nil {
				tagByte, err := json.Marshal(tag)
				if err == nil {
					repository.RedisClient.Set(redisKey, string(tagByte), repository.RedisTtl)
				}
			}
			return tag, nil
		}
		return tags, nil
	}
	return t.tr.FetchFilteredTagsPaginated(curriculumType, tagType, start, limit)
}

func (t *TagsServiceStruct) FetchFilteredTagsPaginatedForAdmin(curriculumType *string, tagType *string, start *int, limit *int) (tags []*domain.Tags, err error) {
	if *tagType == domain.TagTypeEnum.Country {
		redisKey := repository.CurriculumCountryAdminPrefix + *curriculumType + ":" + *tagType + ":" + strconv.Itoa(*start) + ":" + strconv.Itoa(*limit)
		val, err := repository.RedisClient.Get(redisKey).Result()
		err1 := json.Unmarshal([]byte(val), &tags)
		if err != nil || err1 != nil {
			tag, err := t.tr.FetchFilteredTagsPaginatedForAdmin(curriculumType, tagType, start, limit)
			if err != nil {
				return nil, err
			}
			if tag != nil {
				tagByte, err := json.Marshal(tag)
				if err == nil {
					repository.RedisClient.Set(redisKey, string(tagByte), repository.RedisTtl)
				}
			}
			return tag, nil
		}
		return tags, nil
	}
	return t.tr.FetchFilteredTagsPaginatedForAdmin(curriculumType, tagType, start, limit)
}

func (t *TagsServiceStruct) IsCollegePresent(tagType *string, tagId *string) (hasCollege bool, err error) {
	return t.ptmr.IsCollegePresent(tagType, tagId)
}

func (t *TagsServiceStruct) FetchByTagGroup(tagGroup *string, tagType *string) (tags []*domain.Tags, err error) {
	return t.tr.FetchByTagGroup(tagGroup, tagType)
}

func (t *TagsServiceStruct) FetchByInTags(ids []*string) (tags []*domain.Tags, err error) {
	if len(ids) == 0 {
		return
	}
	redisKeys := make([]string, len(ids))
	var allTagData []*domain.Tags
	for i, id := range ids {
		redisKeys[i] = repository.CurriculumPrefix + *id
	}
	if len(ids) > 0 {
		values, _ := repository.RedisClient.MGet(redisKeys...).Result()
		for i, val := range values {
			var tagData *domain.Tags
			if val != nil {
				err = json.Unmarshal([]byte(val.(string)), &tagData)
			}
			if err != nil || val == nil {
				tag, err := t.tr.FetchTags(ids[i])
				if err != nil || tag == nil {
					continue
				}
				tagData = tag
				tagByte, err := json.Marshal(*tag)
				if err == nil {
					repository.RedisClient.Set(redisKeys[i], string(tagByte), repository.RedisTtl)
				}
			}
			allTagData = append(allTagData, tagData)
		}
		if len(allTagData) > 0 {
			return allTagData, nil
		}
	}
	return t.tr.FetchByInTags(ids)
}

func (t *TagsServiceStruct) CreateTags(tx *sql.Tx, tags *domain.Tags) (id *string, err error) {
	return t.tr.CreateTags(tx, tags)
}

func (t *TagsServiceStruct) CreateParentTagMapping(tx *sql.Tx, parentTagMapping *domain.ParentTagMapping) (err error) {
	redisKey := repository.CurriculumParentTagMappingPrefix + *parentTagMapping.TagID
	if err := repository.RedisClient.Del(redisKey).Err(); err != nil {
		return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
	}
	if *parentTagMapping.Order > 0 {
		redisKey = repository.CurriculumTagOrderPrefix + *parentTagMapping.ParentTagID + ":" + *parentTagMapping.TagType
		if err := repository.RedisClient.Del(redisKey).Err(); err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
		}
	}
	return t.ptmr.CreateParentTagMapping(tx, parentTagMapping)
}

func (t *TagsServiceStruct) FetchParentTagMappings(id *string) (parentTagMappings []*domain.ParentTagMapping, err error) {
	redisKey := repository.CurriculumParentTagMappingPrefix + *id
	val, err := repository.RedisClient.Get(redisKey).Result()
	err1 := json.Unmarshal([]byte(val), &parentTagMappings)
	if err != nil || err1 != nil {
		parentTagMappings, err = t.ptmr.FetchParentTagMappings(id)
		if err != nil {
			return nil, err
		}
		if len(parentTagMappings) > 0 {
			tagByte, err := json.Marshal(parentTagMappings)
			if err == nil {
				repository.RedisClient.Set(redisKey, string(tagByte), repository.RedisTtl)
			}
		}
		return parentTagMappings, nil
	}
	return t.ptmr.FetchParentTagMappings(id)
}

func (t *TagsServiceStruct) ToggleHideParentTagMapping(tx *sql.Tx, hidden bool, tagId *string, id *string) (err error) {
	redisKey := repository.CurriculumParentTagMappingPrefix + *tagId
	if err := repository.RedisClient.Del(redisKey).Err(); err != nil {
		return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
	}
	return t.ptmr.ToggleHideParentTagMapping(tx, hidden, id)
}

func (t *TagsServiceStruct) DeleteParentTagMapping(tx *sql.Tx, id *string) (err error) {
	return t.ptmr.DeleteParentTagMapping(tx, id)
}

func (t *TagsServiceStruct) FetchByInParentTagMappings(ids []*string) (parentTagMappings []*domain.ParentTagMapping, err error) {
	if len(ids) == 0 {
		return
	}
	redisKeys := make([]string, len(ids))
	for i, id := range ids {
		redisKeys[i] = repository.CurriculumParentTagMappingPrefix + *id
	}
	if len(ids) > 0 {
		values, _ := repository.RedisClient.MGet(redisKeys...).Result()
		for i, val := range values {
			var parentTagMappingData []*domain.ParentTagMapping
			if val != nil {
				err = json.Unmarshal([]byte(val.(string)), &parentTagMappingData)
			}
			if err != nil || val == nil {
				parentTagMappingsSlice, err := t.ptmr.FetchParentTagMappings(ids[i])
				if err != nil || parentTagMappingsSlice == nil {
					return nil, err
				}
				if len(parentTagMappingsSlice) > 0 {
					parentTagMappingData = parentTagMappingsSlice
					tagByte, err := json.Marshal(parentTagMappingsSlice)
					if err == nil {
						repository.RedisClient.Set(redisKeys[i], string(tagByte), repository.RedisTtl)
					}
				}
			}
			parentTagMappings = append(parentTagMappings, parentTagMappingData...)
		}
		if len(parentTagMappings) > 0 {
			return parentTagMappings, nil
		}
	}
	return t.ptmr.FetchByInParentTagMappings(ids)
}

func (t *TagsServiceStruct) FetchFilteredParentTagMappings(tagType *string, id *string) (parentTagMappings []*domain.ParentTagMapping, err error) {
	return t.ptmr.FetchFilteredParentTagMappings(tagType, id)
}

func (t *TagsServiceStruct) FetchParentTagMappingByParentTagIdTagId(tagId *string, parentTagId *string) (parentTagMappings *domain.ParentTagMapping, err error) {
	return t.ptmr.FetchParentTagMappingByParentTagIdTagId(tagId, parentTagId)
}

func (t *TagsServiceStruct) FetchByInParentTagMappingsByParentTagIdTagIds(ids []*string, parentTagId *string) (parentTagMappings []*domain.ParentTagMapping, err error) {
	return t.ptmr.FetchByInParentTagMappingsByParentTagIdTagIds(ids, parentTagId)
}

func (t *TagsServiceStruct) DeleteTags(tx *sql.Tx, id *string) (err error) {
	redisKey := repository.CurriculumPrefix + *id
	if err := repository.RedisClient.Del(redisKey).Err(); err != nil {
		return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
	}
	return t.tr.DeleteTags(tx, id)
}

func (t *TagsServiceStruct) UpdateLocale(tx *sql.Tx, localeAvailable bool, id *string) (err error) {
	redisKey := repository.CurriculumPrefix + *id
	if err := repository.RedisClient.Del(redisKey).Err(); err != nil {
		return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
	}
	return t.tr.UpdateLocale(tx, localeAvailable, id)
}

func (t *TagsServiceStruct) UpdateTag(updateTag *domain.UpdateTag) (err error) {
	redisKey := repository.CurriculumPrefix + *updateTag.ID
	if *updateTag.Type == domain.TagTypeEnum.Country {
		redisPattern := repository.CurriculumCountryPrefix + "*"
		keys, err := repository.RedisClient.Keys(redisPattern).Result()
		if err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
		}
		if len(keys) > 0 {
			err = repository.RedisClient.Del(keys...).Err()
		}
		if err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
		}
	}
	if err := repository.RedisClient.Del(redisKey).Err(); err != nil {
		return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
	}
	return t.tr.UpdateTag(updateTag)
}

func (t *TagsServiceStruct) ToggleTags(publish bool, ids []*string) (err error) {
	for _, id := range ids {
		redisKey := repository.CurriculumPrefix + *id
		if err := repository.RedisClient.Del(redisKey).Err(); err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
		}
	}
	return t.tr.ToggleTags(publish, ids)
}

func (t *TagsServiceStruct) CreateTagLocaleMapping(tx *sql.Tx, tagLocaleMapping *domain.TagLocaleMapping) (err error) {
	redisKey := repository.CurriculumTagLocaleMappingPrefix + *tagLocaleMapping.TagID + ":" + *tagLocaleMapping.CountryId + ":" + *tagLocaleMapping.Locale
	if err := repository.RedisClient.Del(redisKey).Err(); err != nil {
		return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
	}
	return t.tlmr.CreateTagLocaleMapping(tx, tagLocaleMapping)
}

func (t *TagsServiceStruct) FetchTagLocaleMappings(id *string) (tagLocaleMappings []*domain.TagLocaleMapping, err error) {
	return t.tlmr.FetchTagLocaleMappings(id)
}

func (t *TagsServiceStruct) FetchTagLocalesByTagIds(ids []*string, locale *string, countryId *string) (tagLocaleMappings []*domain.TagLocaleMapping, err error) {
	return t.tlmr.FetchTagLocalesByTagIds(ids, locale, countryId)
}

func (t *TagsServiceStruct) DeleteTagLocaleMapping(tx *sql.Tx, tagLocaleMapping *domain.TagLocaleMapping) (err error) {
	redisKey := repository.CurriculumTagLocaleMappingPrefix + *tagLocaleMapping.TagID + ":" + *tagLocaleMapping.CountryId + ":" + *tagLocaleMapping.Locale
	if err := repository.RedisClient.Del(redisKey).Err(); err != nil {
		return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
	}
	return t.tlmr.DeleteTagLocaleMapping(tx, tagLocaleMapping.ID)
}

func (t *TagsServiceStruct) FetchTagLocaleMappingsByLocale(tagData []*domain.Tags, countryId *string, locale *string) (tagResults []*domain.Tags, err error) {
	if len(tagData) == 0 || countryId == nil || locale == nil {
		return tagData, nil
	}
	var tagLocaleIds []*string
	for _, v := range tagData {
		if v.LocaleAvailable {
			tagLocaleIds = append(tagLocaleIds, v.ID)
		}
	}
	tagLocaleMap, err := t.fetchTagLocaleMappingsDataByLocale(tagLocaleIds, countryId, locale)
	if err != nil {
		return
	}
	if tagLocaleMap != nil {
		for _, v := range tagData {
			locale, ok := tagLocaleMap[*v.ID]
			if ok {
				v.LocaleName = &locale
			}
		}
	}
	return tagData, nil
}

func (t *TagsServiceStruct) FetchTagLocaleMappingsByLocaleForContext(tagData []*domain.Tags, countryId *string, locale *string) (tagResults []*domain.Tags, err error) {
	if len(tagData) == 0 || countryId == nil || locale == nil {
		return tagData, nil
	}
	var tagLocaleIds []*string
	for _, v := range tagData {
		if v.LocaleAvailable {
			tagLocaleIds = append(tagLocaleIds, v.ID)
		}
	}
	tagLocaleMap, err := t.fetchTagLocaleMappingsDataByLocale(tagLocaleIds, countryId, locale)
	if err != nil {
		return
	}
	if tagLocaleMap != nil {
		for _, v := range tagData {
			locale, ok := tagLocaleMap[*v.ID]
			if ok {
				v.Name = &locale
			}
		}
	}
	return tagData, nil
}

func (t *TagsServiceStruct) fetchTagLocaleMappingsDataByLocale(tagIds []*string, countryId *string, locale *string) (tagLocaleMap map[string]string, err error) {
	tagLocaleMap = make(map[string]string)
	if len(tagIds) == 0 {
		return
	}
	wgDone := make(chan bool)
	tagDataChan := make(chan *domain.TagLocaleMapping, len(tagIds))
	var wg sync.WaitGroup
	for _, tagId := range tagIds {
		wg.Add(1)
		go func(tagId *string) {
			defer func() {
				if err := recover(); err != nil {
					logger.Client.Error("getLocaleTagMappingPanicked:id:"+*tagId, logger.GetErrorStack())
					wg.Done()
				}
			}()
			var tagData domain.TagLocaleMapping
			redisKey := repository.CurriculumTagLocaleMappingPrefix + *tagId + ":" + *countryId + ":" + *locale
			val, err := repository.RedisClient.Get(redisKey).Result()
			err1 := json.Unmarshal([]byte(val), &tagData)
			if err != nil || err1 != nil {
				tag, err := t.tlmr.FetchTagLocaleMappingByLocale(tagId, countryId, locale)
				if err != nil {
					wg.Done()
					return
				}
				if tag != nil {
					tagByte, err := json.Marshal(*tag)
					if err == nil {
						repository.RedisClient.Set(redisKey, string(tagByte), repository.RedisTtl)
					}
					tagData = *tag
				}
			}
			tagDataChan <- &tagData
			wg.Done()
		}(tagId)
	}
	go func() {
		wg.Wait()
		close(tagDataChan)
		close(wgDone)
	}()
	select {
	case <-wgDone:
		for v := range tagDataChan {
			if v.TagID != nil {
				tagLocaleMap[*v.TagID] = *v.Name
			}
		}
		return tagLocaleMap, nil
	}
}

func (t *TagsServiceStruct) FetchTagOrders(parentTagIds *string, tagType *string) (tagOrders []*domain.ParentTagMapping, err error) {
	redisKey := repository.CurriculumTagOrderPrefix + *parentTagIds + ":" + *tagType
	val, err := repository.RedisClient.Get(redisKey).Result()
	err1 := json.Unmarshal([]byte(val), &tagOrders)
	if err != nil || err1 != nil {
		tagOrders, err = t.ptmr.FetchParentTagMappingsByParentTagIds(parentTagIds, tagType)
		if err != nil {
			return nil, err
		}
		if len(tagOrders) > 0 {
			tagByte, err := json.Marshal(tagOrders)
			if err == nil {
				repository.RedisClient.Set(redisKey, string(tagByte), repository.RedisTtl)
			}
		}
		return tagOrders, nil
	}
	return t.ptmr.FetchParentTagMappingsByParentTagIds(parentTagIds, tagType)
}

func (t *TagsServiceStruct) UpdateTagOrders(tx *sql.Tx, orders []*domain.Order, parentTagIds *string, tagType *string) (err error) {
	redisKey := repository.CurriculumTagOrderPrefix + *parentTagIds + ":" + *tagType
	if err := repository.RedisClient.Del(redisKey).Err(); err != nil {
		return noonerror.New(noonerror.ErrInternalServer, "redisDeleteError")
	}
	errChan := make(chan error)
	wgDone := make(chan bool)
	var wg sync.WaitGroup
	for _, order := range orders {
		wg.Add(1)
		go func(order *domain.Order) {
			defer func() {
				if err := recover(); err != nil {
					logger.Client.Error("updateTagOrdersPanicked:id:"+*order.ID, logger.GetErrorStack())
					errChan <- noonerror.New(noonerror.ErrInternalServer, "updateTagOrdersPanicked")
				}
			}()
			err := t.ptmr.UpdateTagOrder(tx, order.Order, order.SqlId)
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}(order)
	}
	go func() {
		wg.Wait()
		close(wgDone)
	}()
	select {
	case <-wgDone:
		break
	case err = <-errChan:
		return
	}
	return
}

func verifyAndFetchParentCurriculumTags(curriculumType *string, tagHierarchySlice []*domain.Tags, tagLevel int) (parentTags *string, err error) {
	curriculumHierarchy, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return
	}
	mappedCurriculumType, err := flow.CurriculumMapper(curriculumType)
	if err != nil {
		return
	}
	allLevels := make([]int, tagLevel)
	parentTagIds := []string{}
	allLevels[tagLevel-1] = tagLevel
	sort.Slice(tagHierarchySlice, func(i, j int) bool {
		return curriculumHierarchy[*tagHierarchySlice[i].Type].Level <curriculumHierarchy[*tagHierarchySlice[j].Type].Level
	})
	for _, parentTag := range tagHierarchySlice {
		if parentTag == nil {
			return nil, noonerror.New(noonerror.ErrBadRequest, "parentTagTypeAbsent")
		}
		parentHierarchyTag, ok := curriculumHierarchy[*parentTag.Type]
		if !ok {
			return nil, noonerror.New(noonerror.ErrBadRequest, "parentTagInvalid")
		}
		parentCurriculumType := parentTag.CurriculumType
		if parentCurriculumType != "root" && *mappedCurriculumType != parentCurriculumType {
			return nil, noonerror.New(noonerror.ErrBadRequest, "parentTagTypeInvalid")
		}
		if parentHierarchyTag.Level >= tagLevel {
			return nil, noonerror.New(noonerror.ErrBadRequest, "parentTagInvalid id: "+*parentTag.ID)
		}
		allLevels[parentHierarchyTag.Level-1] = parentHierarchyTag.Level
		parentTagIds = append(parentTagIds,*parentTag.ID )
	}
	//for _, level := range allLevels {
	//	if level == 0 {
	//		return nil, noonerror.New(noonerror.ErrBadRequest, "parentTagMissing")
	//	}
	//}
	parentTag := strings.Join(parentTagIds, ".")
	return &parentTag, nil
}

func verifyAndFetchParentCurriculumTagsForContent(curriculumType *string, tagType *string, tagHierarchySlice []*domain.Tags, accessType string) (parentTags *string, parentHideOrderTags *string, parentIdentifierTagIds map[string]string, err error) {
	curriculumHierarchy, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return
	}
	if curriculumHierarchy[*tagType].Level == 1 {
		return nil, nil, nil, noonerror.New(noonerror.ErrBadRequest, "notEnoughParentTags")
	}
	_, ok := curriculumHierarchy[*tagType]
	if !ok {
		return nil, nil, nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	mappedCurriculumType, err := flow.CurriculumMapper(curriculumType)
	if err != nil {
		return
	}
	allParentTagIds := make([]string, curriculumHierarchy[*tagType].Level-1)
	parentCurriculumTagsSlice := make([]string, curriculumHierarchy[*tagType].Level-1)
	parentIdentifierTagLevels := make([]int, curriculumHierarchy[*tagType].Level-1)
	parentIdentifierTagIds = make(map[string]string)
	for _, parentTag := range tagHierarchySlice {
		if parentTag == nil {
			return nil, nil, nil, noonerror.New(noonerror.ErrBadRequest, "parentTagTypeAbsent")
		}
		parentHierarchyTag, ok := curriculumHierarchy[*parentTag.Type]
		if !ok {
			return nil, nil, nil, noonerror.New(noonerror.ErrBadRequest, "hierarchyInvalid")
		}
		parentCurriculumType := parentTag.CurriculumType
		if parentCurriculumType != "root" && *mappedCurriculumType != parentCurriculumType {
			return nil, nil, nil, noonerror.New(noonerror.ErrBadRequest, "parentTagTypeInvalid")
		}
		if parentHierarchyTag.Level > len(allParentTagIds) {
			return nil, nil, nil, noonerror.New(noonerror.ErrBadRequest, "hierarchyInvalid")
		}
		allParentTagIds[parentHierarchyTag.Level-1] = *parentTag.ID
		if !parentHierarchyTag.IsIdentifier {
			parentCurriculumTagsSlice[parentHierarchyTag.Level-1] = *parentTag.ID
		} else {
			if parentIdentifierTagLevels[parentHierarchyTag.Level-1] > 0 {
				return nil, nil, nil, noonerror.New(noonerror.ErrBadRequest, "illegalParentIdentifierTags")
			}
			parentIdentifierTagLevels[parentHierarchyTag.Level-1] = parentHierarchyTag.Level
			parentIdentifierTagIds[*parentTag.Type] = *parentTag.ID
		}
	}
	for k, v := range curriculumHierarchy {
		if v.Level <= len(allParentTagIds) && len(allParentTagIds[v.Level-1]) == 0 {
			if v.IsIdentifier && accessType == constant.ReadAccessType {
				continue
			}
			if k == domain.TagTypeEnum.Board {
				continue
			}
			return nil, nil, nil, noonerror.New(noonerror.ErrBadRequest, "parentTagMissing")
		}
	}
	var fullParentTagIds []string
	for _, v := range allParentTagIds {
		if len(v) > 0 {
			fullParentTagIds = append(fullParentTagIds, v)
		}
	}
	parentTag := strings.Join(fullParentTagIds, ".")
	var parentCurriculumTags []string
	for _, v := range parentCurriculumTagsSlice {
		if len(v) > 0 {
			parentCurriculumTags = append(parentCurriculumTags, v)
		}
	}
	parentCurriculumTag := strings.Join(parentCurriculumTags, ".")
	return &parentCurriculumTag, &parentTag, parentIdentifierTagIds, nil
}

func isRootCurriculum(curriculumType *string) (isRoot string) {
	isRoot = constant.DerivedCurriculum
	if *curriculumType == "k12" || *curriculumType == "university" || *curriculumType == "general_test_prep" || *curriculumType == "general_skill" {
		return constant.RootCurriculum
	}
	return
}

func assignDefaults(attributes map[string]interface{}, tagType string) map[string]interface{} {
	if len(tagType) == 0 || attributes == nil {
		return attributes
	}
	configParams := config.GetConfig()
	if tagType == domain.TagTypeEnum.Subject || tagType == domain.TagTypeEnum.Test || tagType == domain.TagTypeEnum.Skill || tagType == domain.TagTypeEnum.Course {
		color, ok := attributes["color"]
		if (!ok || color == nil) && len(configParams.DefaultColor) > 0 {
			attributes["color"] = configParams.DefaultColor
		}
		pic, ok := attributes["pic"]
		if (!ok || pic == nil) && len(configParams.DefaultPic) > 0 {
			attributes["pic"] = configParams.DefaultPic
		}
	}
	return attributes
}

func (t *TagsServiceStruct) FetchTagIdFromLegacyId(legacyType *string, id *string) (legacyTagMappings []*domain.LegacyTagMapping, err error) {
	return t.ltmr.FetchTagIdFromLegacyId(legacyType, id)
}

func (t *TagsServiceStruct) FetchLegacyIdFromTagId(tagId *string) (legacyTagMappings []*domain.LegacyTagMapping, err error) {
	return t.ltmr.FetchLegacyIdFromTagId(tagId)
}

func (t *TagsServiceStruct) FetchLegacyIdFromTagIds(tagIds []*string) (legacyTagMappings []*domain.LegacyTagMapping, err error) {
	return t.ltmr.FetchLegacyIdFromTagIds(tagIds)
}

func (t *TagsServiceStruct) FetchGradesFromProductId(productId *string) (gradeProducts []*domain.GradeProduct, err error) {
	redisKey := repository.CurriculumGradeProductPrefix + *productId
	val, err := repository.RedisClient.Get(redisKey).Result()
	err1 := json.Unmarshal([]byte(val), &gradeProducts)
	if err != nil || err1 != nil {
		gradeProducts, err = t.gpr.FetchGradesFromProductId(productId)
		if err != nil {
			return nil, err
		}
		if len(gradeProducts) > 0 {
			tagByte, err := json.Marshal(gradeProducts)
			if err == nil {
				repository.RedisClient.Set(redisKey, string(tagByte), repository.RedisTtl)
			}
		}
		return gradeProducts, nil
	}
	return t.gpr.FetchGradesFromProductId(productId)
}

func (t *TagsServiceStruct) OrderTags(tags []*domain.Tags, tagType *string, curriculumType *string, hierarchy *string) ([]*domain.Tags, error) {
	var tagOrders []*domain.ParentTagMapping
	curriculum, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "curriculumTypeInvalid")
	}
	curriculumInfo, ok := curriculum[*tagType]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "typeInvalid")
	}
	if curriculumInfo.IsOrdered {
		tagOrders, _ = t.FetchTagOrders(hierarchy, tagType)
		if len(tags) > len(tagOrders) && len(tagOrders) > 0 {
			return nil, noonerror.New(noonerror.ErrInternalServer, "tagLengthMismatch")
		}
	}
	//var tagOrderWithoutZeroOrders []*domain.ParentTagMapping
	//for _, v := range tagOrders {
	//	if v != nil && *v.Order > 0 {
	//		tagOrderWithoutZeroOrders = append(tagOrderWithoutZeroOrders, v)
	//	}
	//}
	//tagOrders = tagOrderWithoutZeroOrders
	if len(tagOrders) == 0 {
		sort.SliceStable(tags, func(i, j int) bool {
			return strings.ToLower(*tags[i].Name) < strings.ToLower(*tags[j].Name)
		})
		return tags, nil}
	//} else {
	//	sort.SliceStable(tagOrders, func(i, j int) bool {
	//		return *tagOrders[i].Order < *tagOrders[j].Order
	//	})
	//}
	sort.SliceStable(tagOrders, func(i, j int) bool {
		return *tagOrders[i].Order < *tagOrders[j].Order
	})
	tagsMap := make(map[string]*domain.Tags)
	for _, v := range tags {
		tagsMap[*v.ID] = v
	}
	var finalTags []*domain.Tags
	for _, v := range tagOrders {
		k, ok := tagsMap[*v.TagID]
		if ok {
			finalTags = append(finalTags, k)
		}
	}
	return finalTags, nil
}

func getCurriculumTypeFromParent(tags []*domain.Tags) (curriculumType *string) {
	if len(tags) == 0 {
		return
	}
	curriculumType = new(string)
	for _, v := range tags {
		if v == nil {
			continue
		}
		if *v.Type == domain.TagTypeEnum.Grade {
			*curriculumType = domain.CurriculumTypeEnum.K12
			return curriculumType
		}
		if *v.Type == domain.TagTypeEnum.Degree {
			*curriculumType = domain.CurriculumTypeEnum.University
			return curriculumType
		}
		if *v.Type == domain.TagTypeEnum.Test {
			if v.CurriculumType == domain.CurriculumTypeEnum.K12 {
				*curriculumType = domain.CurriculumTypeEnum.K12TestPrep
			} else if v.CurriculumType == domain.CurriculumTypeEnum.University {
				*curriculumType = domain.CurriculumTypeEnum.UniversityTestPrep
			} else {
				*curriculumType = domain.CurriculumTypeEnum.GeneralTestPrep
			}
			return curriculumType
		}
		if *v.Type == domain.TagTypeEnum.Skill {
			if v.CurriculumType == domain.CurriculumTypeEnum.K12 {
				*curriculumType = domain.CurriculumTypeEnum.K12Skill
			} else if v.CurriculumType == domain.CurriculumTypeEnum.University {
				*curriculumType = domain.CurriculumTypeEnum.UniversitySkill
			} else {
				*curriculumType = domain.CurriculumTypeEnum.GeneralSkill
			}
			return curriculumType
		}
	}
	return nil
}
