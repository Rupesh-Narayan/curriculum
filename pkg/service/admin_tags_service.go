package service

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	noonerror "bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/flow"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	repository "bitbucket.org/noon-micro/curriculum/pkg/repository/mysql"
	"bitbucket.org/noon-micro/curriculum/pkg/service/constant"
	dtomapper "bitbucket.org/noon-micro/curriculum/pkg/service/mapper"
	"context"
	"github.com/jinzhu/copier"
	"strconv"
	"strings"
	"sync"
	"time"
)

type AdminTagsServiceStruct struct {
	ts domain.TagsService
	es domain.Elastic
}

func NewAdminTagsService(ts domain.TagsService, es domain.Elastic) *AdminTagsServiceStruct {
	return &AdminTagsServiceStruct{ts: ts, es: es}
}

func (t *AdminTagsServiceStruct) GetTags(tags *domain.GetTags) (getTagResponse *domain.GetTagsResponse, err error) {
	switch *tags.TagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.getCurriculumTags(tags)
	case domain.TagGroupEnum.Content:
		return t.getContentTags(tags)
	}
	return
}

func (t *AdminTagsServiceStruct) GetAdminTags(tags *domain.GetAdminTags) (getTagResponse *domain.GetTagsResponse, err error) {

	switch *tags.TagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.getCurriculumTagsForLibrary(tags)
	case domain.TagGroupEnum.Content:
		return t.getContentTagsForLibrary(tags)
	}
	return

}

func (t *AdminTagsServiceStruct) GetTagsSearch(tags *domain.GetTags) (getTagResponse *domain.GetTagsResponse, err error) {
	switch *tags.TagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.getCurriculumTagsSearch(tags)
	case domain.TagGroupEnum.Content:
		return t.getContentTagsSearch(tags)
	case domain.TagGroupEnum.Identifier:
		return t.getIdentifierTagsSearch(tags)
	}
	return
}

func (t *AdminTagsServiceStruct) CreateAdminTags(tagGroup *string, tags *domain.CreateTags) (tagResponse *domain.TagResponse, err error) {
	switch *tagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.createCurriculumTag(tags)
	case domain.TagGroupEnum.Content:
		return t.createContentTag(tags)
	case domain.TagGroupEnum.Identifier:
		return t.createIdentifierTag(tags)
	}
	return
}

func (t *AdminTagsServiceStruct) UpdateAdminTags(tagGroup *string, tags *domain.UpdateTags) (tagResponse *domain.TagResponse, err error) {
	switch *tagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.updateCurriculumTag(tags)
	case domain.TagGroupEnum.Content:
		return t.updateContentTag(tags)
	}
	return
}

func (t *AdminTagsServiceStruct) UpdateTagOrder(tagGroup *string, tags *domain.UpdateTagOrder) (err error) {
	switch *tagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.updateCurriculumTagOrder(tags)
	case domain.TagGroupEnum.Content:
		return t.updateContentTagOrder(tags)
	}
	return
}

func (t *AdminTagsServiceStruct) UpdateTag(updateTag *domain.UpdateTag) (err error) {
	tagData, err := t.ts.FetchTags(updateTag.ID)
	if err != nil {
		return
	}
	for k, v := range tagData.Attributes {
		_, ok := updateTag.Attributes[k]
		if !ok && updateTag.Attributes != nil {
			updateTag.Attributes[k] = v
		}
	}
	updateTag.Type = tagData.Type
	return t.ts.UpdateTag(updateTag)
}

func (t *AdminTagsServiceStruct) RemoveAdminTagFromHierarchy(tagGroup *string, tags *domain.RemoveHierarchy) (tagResponse *domain.TagResponse, err error) {
	switch *tagGroup {
	case domain.TagGroupEnum.Curriculum:
		return t.removeCurriculumTagHierarchy(tags)
	case domain.TagGroupEnum.Content:
		return t.removeContentTagHierarchy(tags)
	}
	return
}

func (t *AdminTagsServiceStruct) UpdateMultipleAdminTags(tags *domain.UpdateMultipleTags) (ids []*string, err error) {
	if len(tags.Hierarchy) == 0 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "hierarchyAbsent")
	}
	tagHierarchySlice, err := t.ts.GetTagsConcurrent(tags.Hierarchy)
	if err != nil {
		return
	}
	curriculumType := tags.CurriculumType
	curriculumHierarchy, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return
	}
	tagHierarchy, ok := curriculumHierarchy[*tags.Type]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "invalidContent")
	}
	curriculumHierarchyType := isRootCurriculum(tags.CurriculumType)
	parentTags, parentHideOrderTags, parentIdentifierTagIds, err := verifyAndFetchParentCurriculumTagsForContent(tags.CurriculumType, tags.Type, tagHierarchySlice, constant.WriteAccessType)
	if err != nil {
		return
	}
	type tagStruct struct {
		Id   *string
		Type *string
	}
	var allParentTags []tagStruct
	hierarchyType := constant.HierarchyCurriculum
	allParentTags = append(allParentTags, tagStruct{parentTags, &curriculumHierarchyType})
	allParentTags = append(allParentTags, tagStruct{parentHideOrderTags, &hierarchyType})
	for k, v := range parentIdentifierTagIds {
		key := k
		val := v
		allParentTags = append(allParentTags, tagStruct{&val, &key})
	}
	wgDone := make(chan bool)
	errChan := make(chan error, len(tags.IDs))
	successIdsChan := make(chan *string, len(tags.IDs))
	failureIds := make(chan *string, len(tags.IDs))
	wg := new(sync.WaitGroup)
	for _, val := range tags.IDs {
		wg.Add(1)
		go func(id *string) {
			defer func() {
				if err := recover(); err != nil {
					logger.Client.Error("updateMultipleAdminTagsPanicked", logger.GetErrorStack())
					failureIds <- id
					errChan <- noonerror.New(noonerror.ErrInternalServer, "updateMultipleAdminTagsPanicked")
					wg.Done()
					return
				}
			}()
			ctx := context.Background()
			tx, err := repository.Db.BeginTx(ctx, nil)
			if err != nil {
				failureIds <- id
				errChan <- noonerror.New(noonerror.ErrInternalServer, "contextCreationError")
				wg.Done()
				return
			}
			parentTagMappings, err := t.ts.FetchParentTagMappings(id)
			if err != nil {
				failureIds <- id
				errChan <- err
				wg.Done()
				return
			}
			addTagsMap := make(map[string]string)
			hiddenTagsMap := make(map[string]string)
			var addTags []*string
			var hiddenTags []*string
			var allParents []*string
			for _, v := range allParentTags {
				isPresent := false
				for _, parentTagMapping := range parentTagMappings {
					if *parentTagMapping.ParentTagID == *v.Id && parentTagMapping.Hidden {
						isPresent = true
						hiddenTagsMap[*parentTagMapping.ID] = *v.Id
						hiddenTags = append(hiddenTags, v.Id)
						allParents = append(allParents, v.Id)
						break
					} else if *parentTagMapping.ParentTagID == *v.Id && !parentTagMapping.Hidden {
						isPresent = true
						break
					}
				}
				if !isPresent {
					addTagsMap[*v.Type] = *v.Id
					addTags = append(addTags, v.Id)
					allParents = append(allParents, v.Id)
				}
			}
			rollback := func() {
				go func() {
					_ = t.es.RemoveParentTags(id, addTags)
					_ = t.es.HideParentTags(id, hiddenTags)
				}()
				failureIds <- id
				errChan <- err
				wg.Done()
			}
			err = t.es.AddParentTags(id, allParents)
			if err != nil {
				rollback()
				return
			}
			for k, v := range addTagsMap {
				key := k
				val := v
				order := 0
				if val == *parentHideOrderTags {
					order = constant.OrderMax
				}
				if err = t.ts.CreateParentTagMapping(tx, &domain.ParentTagMapping{TagID: id, TagType: tags.Type, ParentTagType: &key, ParentTagID: &val, Order: &order, Hidden: false, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
					_ = tx.Rollback()
					rollback()
					return
				}
			}
			for k := range hiddenTagsMap {
				key := k
				if err = t.ts.ToggleHideParentTagMapping(tx, false, id, &key); err != nil {
					_ = tx.Rollback()
					rollback()
					return
				}
			}
			if err = tx.Commit(); err != nil {
				rollback()
				return
			}
			successIdsChan <- id
			wg.Done()
			return
		}(val)
	}
	go func() {
		wg.Wait()
		close(successIdsChan)
		close(wgDone)
	}()
	successIdMap := make(map[string]string)
	var successIdsString []*string
	select {
	case <-wgDone:
		for v := range successIdsChan {
			successIdMap[*v] = ""
			successIdsString = append(successIdsString, v)
		}
	}
	if tagHierarchy.IsOrdered && len(tagHierarchySlice) > 0 {
		tagOrders, _ := t.ts.FetchTagOrders(parentHideOrderTags, tags.Type)
		maxOrder := 0
		for _, v := range tagOrders {
			_, ok := successIdMap[*v.TagID]
			if ok && *v.Order == constant.OrderMax {
				successIdMap[*v.TagID] = *v.ID
			}
			if *v.Order != constant.OrderMax && *v.Order > maxOrder {
				maxOrder = *v.Order
			}
		}
		var orders []*domain.Order
		for _, v := range tags.IDs {
			sqlId, ok := successIdMap[*v]
			if ok && len(sqlId) > 0 {
				maxOrder++
				order := maxOrder
				orders = append(orders, &domain.Order{Order: &order, SqlId: &sqlId})
			}
		}
		ctx := context.Background()
		tx, err := repository.Db.BeginTx(ctx, nil)
		if err != nil {
			return successIdsString, nil
		}
		if err = t.ts.UpdateTagOrders(tx, orders, parentHideOrderTags, tags.Type); err != nil {
			return successIdsString, nil
		}
		if err = tx.Commit(); err != nil {
			return successIdsString, nil
		}
	}
	return successIdsString, nil
}

func (t *AdminTagsServiceStruct) updateCurriculumTagOrder(tags *domain.UpdateTagOrder) (err error) {
	curriculum, err := flow.GetCurriculum(tags.CurriculumType)
	if err != nil {
		return
	}
	curriculumObject, ok := curriculum[*tags.Type]
	if !ok {
		return noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	if !curriculumObject.IsOrdered {
		return noonerror.New(noonerror.ErrBadRequest, "orderingNotAllowed")
	}
	tagHierarchySlice, err := t.ts.GetTagsConcurrent(tags.Hierarchy)
	if err != nil {
		return
	}
	parentTags, err := verifyAndFetchParentCurriculumTags(tags.CurriculumType, tagHierarchySlice, curriculumObject.Level)
	if err != nil {
		return
	}
	tagOrderData, err := t.ts.FetchTagOrders(parentTags, tags.Type)
	if err != nil {
		return
	}
	for _, v := range tags.Orders {
		orderFound := false
		for _, data := range tagOrderData {
			if *data.TagID == *v.ID {
				orderFound = true
				v.SqlId = data.ID
				break
			}
		}
		if !orderFound {
			return noonerror.New(noonerror.ErrBadRequest, "tagIdMissing")
		}
	}
	return t.ts.UpdateTagOrders(nil, tags.Orders, parentTags, tags.Type)
}

func (t *AdminTagsServiceStruct) updateContentTagOrder(tags *domain.UpdateTagOrder) (err error) {
	curriculum, err := flow.GetCurriculum(tags.CurriculumType)
	if err != nil {
		return
	}
	curriculumObject, ok := curriculum[*tags.Type]
	if !ok {
		return noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	if !curriculumObject.IsOrdered {
		return noonerror.New(noonerror.ErrBadRequest, "orderingNotAllowed")
	}
	tagHierarchySlice, err := t.ts.GetTagsConcurrent(tags.Hierarchy)
	if err != nil {
		return
	}
	_, parentHideOrderTags, _, err := verifyAndFetchParentCurriculumTagsForContent(tags.CurriculumType, tags.Type, tagHierarchySlice, constant.WriteAccessType)
	if err != nil {
		return err
	}
	tagOrderData, err := t.ts.FetchTagOrders(parentHideOrderTags, tags.Type)
	if err != nil {
		return
	}
	for _, v := range tags.Orders {
		orderFound := false
		for _, data := range tagOrderData {
			if *data.TagID == *v.ID {
				orderFound = true
				v.SqlId = data.ID
				break
			}
		}
		if !orderFound {
			return noonerror.New(noonerror.ErrBadRequest, "tagIdMissing")
		}
	}
	return t.ts.UpdateTagOrders(nil, tags.Orders, parentHideOrderTags, tags.Type)
}

func (t *AdminTagsServiceStruct) RemoveIdentifierTag(id *string) (err error) {
	tagData, err := t.ts.FetchTags(id)
	if err != nil || tagData == nil {
		return noonerror.New(noonerror.ErrBadRequest, "tagFetchError")
	}
	if tagData.TagGroup != domain.TagGroupEnum.Identifier {
		return noonerror.New(noonerror.ErrBadRequest, "notIdentifier")
	}
	return t.ts.DeleteTags(nil, id)
}

func (t *AdminTagsServiceStruct) MigrateToElastic(start *string, end *string) (err error) {
	startInt, _ := strconv.Atoi(*start)
	endInt, _ := strconv.Atoi(*end)
	if startInt > endInt {
		return noonerror.New(noonerror.ErrBadRequest, "startGreaterThanEnd")
	}
	for i := startInt; i <= endInt; i++ {
		id := new(string)
		*id = strconv.Itoa(i)
		tagData, err := t.ts.FetchTags(id)
		if err != nil || tagData == nil {
			return noonerror.New(noonerror.ErrBadRequest, "tagFetchError")
		}
		parentTagMappings, err := t.ts.FetchParentTagMappings(id)
		if err != nil {
			return noonerror.New(noonerror.ErrBadRequest, "parentTagsFetchError")
		}
		var parentTags []*string
		for _, v := range parentTagMappings {
			parentTags = append(parentTags, v.ParentTagID)
		}
		var createTags domain.CreateTags
		if err = copier.Copy(&createTags, tagData); err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "mapperError")
		}
		createTags.CurriculumType = &tagData.CurriculumType
		createTags.CreatorType = &tagData.CreatorType
		createTags.TagGroup = &tagData.TagGroup
		createElasticEntity, err := dtomapper.CreateElasticTagEntity(id, &createTags, parentTags, tagData.Access)
		if err != nil {
			return err
		}
		err = t.es.CreateTag(createElasticEntity)
		if err != nil {
			return err
		}
		for _, v := range parentTagMappings {
			var hiddenParents []*string
			if v.Hidden {
				hiddenParents = append(hiddenParents, v.ParentTagID)
			}
			if err := t.es.HideParentTags(id, hiddenParents); err != nil {
				return err
			}
		}
		var tagNames []*domain.TagName
		defaultLocale := constant.DefaultLocale
		tagNames = append(tagNames, &domain.TagName{Locale: &defaultLocale, Value: tagData.Name})
		if tagData.LocaleAvailable {
			tagLocaleMappings, err := t.ts.FetchTagLocaleMappings(id)
			if err != nil {
				return err
			}
			for _, val := range tagLocaleMappings {
				tagNames = append(tagNames, &domain.TagName{Locale: val.Locale, Value: val.Name})
			}
			if err := t.es.UpdateTag(id, nil, tagNames); err != nil {
				return err
			}
		}
	}
	return
}

func (t *AdminTagsServiceStruct) GetTag(id *string) (tagResponse *domain.TagResponse, err error) {
	tagResponse = new(domain.TagResponse)
	var tagLocaleData []*domain.TagLocaleMapping
	tagData, err := t.ts.FetchTags(id)
	if err != nil {
		return
	}
	tagResponse.ID = id
	tagResponse.Type = tagData.Type
	tagResponse.Name = tagData.Name
	tagResponse.Attributes = tagData.Attributes
	if tagData.LocaleAvailable {
		tagLocaleData, err = t.ts.FetchTagLocaleMappings(id)
		if err != nil {
			return tagResponse, nil
		}
		for _, val := range tagLocaleData {
			var locale domain.LocaleResponse
			locale.Locale = val.Locale
			locale.Name = val.Name
			locale.CountryId = val.CountryId
			tagResponse.Locale = append(tagResponse.Locale, &locale)
		}
	}
	return tagResponse, nil
}

func (t *AdminTagsServiceStruct) UpdateTagLocale(action *string, tagLocale *domain.TagLocale) (err error) {
	ctx := context.Background()
	tx, err := repository.Db.BeginTx(ctx, nil)
	if err != nil {
		return noonerror.New(noonerror.ErrInternalServer, "ContextCreationError")
	}
	tagData, err := t.ts.FetchTags(tagLocale.ID)
	if err != nil || tagData == nil {
		return noonerror.New(noonerror.ErrBadRequest, "tagFetchError")
	}
	var tagLocaleData []*domain.TagLocaleMapping
	if tagData.LocaleAvailable {
		tagLocaleData, _ = t.ts.FetchTagLocaleMappings(tagLocale.ID)
	}
	totalLocales := len(tagLocaleData)
	tagLocaleMap := make(map[string][]*domain.TagLocaleMapping)
	tagLocaleFinalMap := make(map[string][]*domain.TagName)
	var localDefaultLocale = constant.DefaultLocale
	key := constant.DefaultLocale + ":" + "0"
	tagLocaleFinalMap[key] = append(tagLocaleFinalMap[key], &domain.TagName{Value: tagData.Name, Locale: &localDefaultLocale})
	updated := false
	for _, val := range tagLocaleData {
		key := strings.ToLower(*val.Locale) + ":" + *val.CountryId
		tagLocaleMap[key] = append(tagLocaleMap[key], val)
		locale := strings.ToLower(*val.Locale)
		if key != constant.DefaultLocale+":"+"0" {
			tagLocaleFinalMap[key] = append(tagLocaleFinalMap[key], &domain.TagName{Value: val.Name, Locale: &locale})
		}
	}
	for _, val := range tagLocale.Locale {
		key := strings.ToLower(*val.Locale) + ":" + *val.CountryId
		_, ok := tagLocaleMap[key]
		if ok {
			for _, locale := range tagLocaleMap[key] {
				if err = t.ts.DeleteTagLocaleMapping(tx, locale); err != nil {
					_ = tx.Rollback()
					return err
				}
				updated = true
				tagLocaleFinalMap[key] = nil
				totalLocales--
			}
			tagLocaleFinalMap[key] = nil
		}
		if *action == "add" {
			locale := strings.ToLower(*val.Locale)
			if err = t.ts.CreateTagLocaleMapping(tx, &domain.TagLocaleMapping{Locale: &locale, CountryId: val.CountryId, TagID: tagLocale.ID,
				Name: tagLocale.Name, TagType: tagData.Type, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
				_ = tx.Rollback()
				return err
			}
			updated = true
			tagLocaleFinalMap[key] = append(tagLocaleFinalMap[key], &domain.TagName{Value: tagLocale.Name, Locale: &locale})
			totalLocales++
		}
	}
	if !tagData.LocaleAvailable && totalLocales > 0 {
		if err = t.ts.UpdateLocale(tx, true, tagLocale.ID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if tagData.LocaleAvailable && totalLocales == 0 {
		if err = t.ts.UpdateLocale(tx, false, tagLocale.ID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if updated {
		var tagLocales []*domain.TagName
		for _, v := range tagLocaleFinalMap {
			if v != nil {
				tagLocales = append(tagLocales, v...)
			}
		}
		if err = t.es.UpdateTag(tagLocale.ID, nil, tagLocales); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return noonerror.New(noonerror.ErrInternalServer, "dbCommitError")
	}
	return
}

func (t *AdminTagsServiceStruct) createCurriculumTag(tags *domain.CreateTags) (tagResponse *domain.TagResponse, err error) {
	ctx := context.Background()
	tx, err := repository.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "ContextCreationError")
	}
	if len(tags.Hierarchy) == 0 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "hierarchyAbsent")
	}
	tagHierarchySlice, err := t.ts.GetTagsConcurrent(tags.Hierarchy)
	if err != nil {
		return
	}
	curriculumType := tags.CurriculumType
	curriculumHierarchy, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return
	}
	mappedCurriculumType, err := flow.CurriculumMapper(curriculumType)
	if err != nil {
		return
	}
	tagHierarchy, ok := curriculumHierarchy[*tags.Type]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	if tagHierarchy.Level == 1 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	parentTags, err := verifyAndFetchParentCurriculumTags(curriculumType, tagHierarchySlice, tagHierarchy.Level)
	if err != nil {
		return
	}
	hierarchyType := constant.HierarchyCurriculum
	tags.CountryId = "0"
	tags.Attributes = assignDefaults(tags.Attributes, *tags.Type)
	if len(tags.Access) == 0 {
		tags.Access = domain.AccessEnum.Global
	}
	tagId, err := t.ts.CreateTags(tx, &domain.Tags{Type: tags.Type, Name: tags.Name, CurriculumType: *mappedCurriculumType,
		CreatorId: tags.CreatorId, CreatorType: *tags.CreatorType, Access: tags.Access, TagGroup: *tags.TagGroup, LocaleAvailable: false, CountryId: tags.CountryId,
		Attributes: tags.Attributes, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	if err != nil {
		_ = tx.Rollback()
		return
	}
	createElasticEntity, err := dtomapper.CreateElasticTagEntity(tagId, tags, []*string{parentTags}, tags.Access)
	if err != nil {
		return
	}
	rollback := func() {
		deleteTag := true
		go func() {
			_ = t.es.UpdateTag(tagId, &deleteTag, nil)
		}()
		_ = tx.Rollback()
		return
	}
	if err = t.es.CreateTag(createElasticEntity); err != nil {
		rollback()
	}
	if err = t.es.HideParentTags(tagId, []*string{parentTags}); err != nil {
		rollback()
	}
	order := 0
	if order, err = t.fetchTagOrders(tagHierarchy.IsOrdered, parentTags, tags.Type, rollback); err != nil {
		return
	}
	if err = t.ts.CreateParentTagMapping(tx, &domain.ParentTagMapping{TagID: tagId, TagType: tags.Type, ParentTagType: &hierarchyType, ParentTagID: parentTags, Order: &order, Hidden: true, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()});
		err != nil {
		rollback()
		return
	}
	if err = tx.Commit(); err != nil {
		deleteTag := true
		go func() {
			_ = t.es.UpdateTag(tagId, &deleteTag, nil)
		}()
		return nil, noonerror.New(noonerror.ErrInternalServer, "dbCommitError")
	}
	tagResponse = &domain.TagResponse{
		ID:     tagId,
		Type:   tags.Type,
		Name:   tags.Name,
		Hidden: true,
	}
	return tagResponse, nil
}

func (t *AdminTagsServiceStruct) fetchTagOrders(ordered bool, parentTags *string, tagType *string, rollback func()) (order int, err error) {
	if !ordered {
		return
	}
	tagOrders, _ := t.ts.FetchTagOrders(parentTags, tagType)
	if len(tagOrders) == constant.TagLimit {
		rollback()
		return 0, noonerror.New(noonerror.ErrBadRequest, "tagLimitReached")
	}
	maxOrder := 0
	for _, v := range tagOrders {
		if *v.Order > maxOrder {
			maxOrder = *v.Order
		}
	}
	order = maxOrder + 1
	return order, nil
}

func (t *AdminTagsServiceStruct) createContentTag(tags *domain.CreateTags) (tagResponse *domain.TagResponse, err error) {
	ctx := context.Background()
	tx, err := repository.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "ContextCreationError")
	}
	if len(tags.Hierarchy) == 0 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "hierarchyAbsent")
	}
	tagHierarchySlice, err := t.ts.GetTagsConcurrent(tags.Hierarchy)
	if err != nil {
		return
	}
	var identifierSlice []*domain.Tags
	if tags.Identifier != nil && len(tags.Identifier) > 0 {
		identifierSlice, err = t.ts.GetTagsConcurrent(tags.Identifier)
		if err != nil {
			return
		}
	}
	curriculumType := tags.CurriculumType
	curriculumHierarchy, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return
	}
	mappedCurriculumType, err := flow.CurriculumMapper(curriculumType)
	if err != nil {
		return
	}
	tagHierarchy, ok := curriculumHierarchy[*tags.Type]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "invalidContent")
	}
	parentCurriculumTags, parentHideOrderTags, parentIdentifierTagIds, err := verifyAndFetchParentCurriculumTagsForContent(curriculumType, tags.Type, tagHierarchySlice, constant.WriteAccessType)
	if err != nil {
		return
	}
	tags.CountryId = "0"
	tagId, err := t.ts.CreateTags(tx, &domain.Tags{Type: tags.Type, Name: tags.Name, CurriculumType: *mappedCurriculumType,
		CreatorId: tags.CreatorId, CreatorType: *tags.CreatorType, Access: domain.AccessEnum.Global, TagGroup: *tags.TagGroup, LocaleAvailable: false, CountryId: tags.CountryId,
		Attributes: tags.Attributes, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	if err != nil {
		_ = tx.Rollback()
		return
	}
	parentIdMap := map[string]*string{}
	curriculumHierarchyType := isRootCurriculum(tags.CurriculumType)
	parentIdMap[curriculumHierarchyType] = parentCurriculumTags
	parentIdMap[constant.HierarchyCurriculum] = parentHideOrderTags
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
			parentIdMap[*v.Type] = v.ID
		}
	}
	var allParentTags []*string
	for _, v := range parentIdMap {
		allParentTags = append(allParentTags, v)
	}
	rollback := func() {
		deleteTag := true
		go func() {
			_ = t.es.UpdateTag(tagId, &deleteTag, nil)
		}()
		_ = tx.Rollback()
	}
	createElasticEntity, err := dtomapper.CreateElasticTagEntity(tagId, tags, allParentTags, domain.AccessEnum.Global)
	if err != nil {
		return
	}
	if err = t.es.CreateTag(createElasticEntity); err != nil {
		rollback()
		return
	}
	err = t.es.HideParentTags(tagId, []*string{parentHideOrderTags})
	if err != nil {
		rollback()
		return
	}
	order := 0
	if order, err = t.fetchTagOrders(tagHierarchy.IsOrdered, parentHideOrderTags, tags.Type, rollback); err != nil {
		return
	}
	for k, v := range parentIdMap {
		key := k
		hidden := false
		orderToApply := 0
		if *v == *parentHideOrderTags {
			hidden = true
			orderToApply = order
		}
		if err = t.ts.CreateParentTagMapping(tx, &domain.ParentTagMapping{TagID: tagId, TagType: tags.Type, ParentTagType: &key, ParentTagID: v, Order: &orderToApply, Hidden: hidden, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
			rollback()
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		deleteTag := true
		go func() {
			_ = t.es.UpdateTag(tagId, &deleteTag, nil)
		}()
		return nil, noonerror.New(noonerror.ErrInternalServer, "dbCommitError")
	}
	tagResponse = &domain.TagResponse{
		ID:     tagId,
		Type:   tags.Type,
		Name:   tags.Name,
		Hidden: true,
	}
	return tagResponse, nil
}

func (t *AdminTagsServiceStruct) createIdentifierTag(tags *domain.CreateTags) (tagResponse *domain.TagResponse, err error) {
	ctx := context.Background()
	tx, err := repository.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "ContextCreationError")
	}
	curriculumType := tags.CurriculumType
	mappedCurriculumType, err := flow.CurriculumMapper(curriculumType)
	if err != nil {
		return
	}
	tags.CountryId = "0"
	tagId, err := t.ts.CreateTags(tx, &domain.Tags{Type: tags.Type, Name: tags.Name, CurriculumType: *mappedCurriculumType,
		CreatorId: tags.CreatorId, CreatorType: *tags.CreatorType, Access: domain.AccessEnum.Global, TagGroup: *tags.TagGroup, LocaleAvailable: false, CountryId: tags.CountryId,
		Attributes: tags.Attributes, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	if err != nil {
		_ = tx.Rollback()
		return
	}
	createElasticEntity, err := dtomapper.CreateElasticTagEntity(tagId, tags, []*string{}, domain.AccessEnum.Global)
	if err != nil {
		return
	}
	if err = t.es.CreateTag(createElasticEntity); err != nil {
		deleteTag := true
		go func() {
			_ = t.es.UpdateTag(tagId, &deleteTag, nil)
		}()
		_ = tx.Rollback()
		return
	}
	if err = tx.Commit(); err != nil {
		deleteTag := true
		go func() {
			_ = t.es.UpdateTag(tagId, &deleteTag, nil)
		}()
		return nil, noonerror.New(noonerror.ErrInternalServer, "dbCommitError")
	}
	tagResponse = &domain.TagResponse{
		ID:   tagId,
		Type: tags.Type,
		Name: tags.Name,
	}
	return tagResponse, nil
}

func (t *AdminTagsServiceStruct) updateCurriculumTag(tags *domain.UpdateTags) (tagResponse *domain.TagResponse, err error) {
	ctx := context.Background()
	tx, err := repository.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "ContextCreationError")
	}
	if len(tags.Hierarchy) == 0 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "hierarchyAbsent")
	}
	tagData, err := t.ts.FetchTags(tags.ID)
	if err != nil || tagData == nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagFetchError")
	}
	if *tags.TagGroup != tagData.TagGroup {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagGroupMismatch")
	}
	tagHierarchySlice, err := t.ts.GetTagsConcurrent(tags.Hierarchy)
	if err != nil {
		return
	}
	curriculumHierarchy, err := flow.GetCurriculum(tags.CurriculumType)
	if err != nil {
		return
	}
	tagHierarchy, ok := curriculumHierarchy[*tagData.Type]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	if tagHierarchy.Level == 1 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	parentTags, err := verifyAndFetchParentCurriculumTags(tags.CurriculumType, tagHierarchySlice, tagHierarchy.Level)
	if err != nil {
		return
	}
	hierarchyType := constant.HierarchyCurriculum
	parentTagMapping, err := t.ts.FetchParentTagMappingByParentTagIdTagId(tags.ID, parentTags)
	if err != nil {
		return
	}
	var hierarchyHiddenId string
	if parentTagMapping != nil && !parentTagMapping.Hidden {
		return nil, noonerror.New(noonerror.ErrBadRequest, "parentMappingExist")
	} else if parentTagMapping != nil && parentTagMapping.Hidden {
		hierarchyHiddenId = *parentTagMapping.ID
	}
	rollback := func() {
		if len(hierarchyHiddenId) == 0 {
			go func() {
				_ = t.es.RemoveParentTags(tags.ID, []*string{parentTags})
			}()
		} else {
			go func() {
				_ = t.es.HideParentTags(tags.ID, []*string{parentTags})
			}()
		}
	}
	if err = t.es.AddParentTags(tags.ID, []*string{parentTags}); err != nil {
		rollback()
		return
	}
	order := 0
	if len(hierarchyHiddenId) == 0 {
		if order, err = t.fetchTagOrders(tagHierarchy.IsOrdered, parentTags, tagData.Type, rollback); err != nil {
			return
		}
		if err = t.es.HideParentTags(tags.ID, []*string{parentTags}); err != nil {
			rollback()
			return
		}
		if err = t.ts.CreateParentTagMapping(tx, &domain.ParentTagMapping{TagID: tags.ID, TagType: tagData.Type, ParentTagType: &hierarchyType, ParentTagID: parentTags, Order: &order, Hidden: true, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
			rollback()
			return
		}
	} else {
		if err = t.ts.ToggleHideParentTagMapping(tx, false, tags.ID, &hierarchyHiddenId); err != nil {
			rollback()
			return
		}
	}
	if err = tx.Commit(); err != nil {
		rollback()
		return nil, noonerror.New(noonerror.ErrInternalServer, "dbCommitError")
	}
	hidden := true
	if len(hierarchyHiddenId) > 0 {
		hidden = false
	}
	tagResponse = &domain.TagResponse{
		ID:     tagData.ID,
		Type:   tagData.Type,
		Name:   tagData.Name,
		Hidden: hidden,
	}
	return tagResponse, nil
}

func (t *AdminTagsServiceStruct) updateContentTag(tags *domain.UpdateTags) (tagResponse *domain.TagResponse, err error) {
	ctx := context.Background()
	tx, err := repository.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "ContextCreationError")
	}
	tagData, err := t.ts.FetchTags(tags.ID)
	if err != nil || tagData == nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagFetchError")
	}
	if *tags.TagGroup != tagData.TagGroup {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagGroupMismatch")
	}
	tagHierarchySlice, err := t.ts.GetTagsConcurrent(tags.Hierarchy)
	if err != nil {
		return
	}
	var identifierSlice []*domain.Tags
	if tags.Identifier != nil && len(tags.Identifier) > 0 {
		identifierSlice, err = t.ts.GetTagsConcurrent(tags.Identifier)
		if err != nil {
			return
		}
	}
	curriculumType := tags.CurriculumType
	curriculumHierarchy, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return
	}
	parentTagMappings, err := t.ts.FetchParentTagMappings(tags.ID)
	if err != nil {
		return
	}
	tagHierarchy, ok := curriculumHierarchy[*tagData.Type]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "invalidContent")
	}
	parentIdMap := map[string]*string{}
	parentIdHiddenMap := map[string]*string{}
	type parentStruct struct {
		Id   *string
		Type *string
	}
	var allParentTags []parentStruct
	parentHideOrderTags := new(string)
	if len(tagHierarchySlice) > 0 {
		parentTags, localParentHideOrderTags, parentIdentifierTagIds, err := verifyAndFetchParentCurriculumTagsForContent(curriculumType, tagData.Type, tagHierarchySlice, constant.WriteAccessType)
		if err != nil {
			return nil, err
		}
		curriculumHierarchyType := isRootCurriculum(tags.CurriculumType)
		parentHideOrderTags = localParentHideOrderTags
		hierarchyType := constant.HierarchyCurriculum
		allParentTags = append(allParentTags, parentStruct{parentTags, &curriculumHierarchyType})
		allParentTags = append(allParentTags, parentStruct{parentHideOrderTags, &hierarchyType})
		for k, v := range parentIdentifierTagIds {
			key := k
			val := v
			allParentTags = append(allParentTags, parentStruct{&val, &key})
		}
	}
	if len(identifierSlice) > 0 {
		for _, v := range identifierSlice {
			allParentTags = append(allParentTags, parentStruct{v.ID, v.Type})
		}
	}
	orderHierarchyPresent := false
	for _, v := range allParentTags {
		var continueInner = false
		for _, parentTagMapping := range parentTagMappings {
			if *parentTagMapping.ParentTagType == *v.Type && *parentTagMapping.ParentTagID == *v.Id {
				continueInner = true
				if parentTagMapping.Hidden {
					parentIdHiddenMap[*parentTagMapping.ID] = parentTagMapping.ParentTagID
				}
				if *v.Id == *parentHideOrderTags {
					orderHierarchyPresent = true
				}
				if *v.Id == *parentHideOrderTags && !parentTagMapping.Hidden {
					return nil, noonerror.New(noonerror.ErrBadRequest, "parentMappingExist")
				}
				break
			}
		}
		if continueInner {
			continue
		}
		parentIdMap[*v.Type] = v.Id
	}
	var allParentTagIds []*string
	var allNonHiddenParentTags []*string
	var allHiddenParentTags []*string
	for _, v := range parentIdMap {
		allParentTagIds = append(allParentTagIds, v)
		allNonHiddenParentTags = append(allNonHiddenParentTags, v)
	}
	for _, v := range parentIdHiddenMap {
		allParentTagIds = append(allParentTagIds, v)
		allHiddenParentTags = append(allHiddenParentTags, v)
	}
	rollback := func() {
		go func() {
			_ = t.es.RemoveParentTags(tags.ID, allNonHiddenParentTags)
			_ = t.es.HideParentTags(tags.ID, allHiddenParentTags)
		}()
	}
	if err = t.es.AddParentTags(tags.ID, allParentTagIds); err != nil {
		rollback()
		return
	}
	order := 0
	if !orderHierarchyPresent {
		if order, err = t.fetchTagOrders(tagHierarchy.IsOrdered, parentHideOrderTags, tagData.Type, rollback); err != nil {
			return
		}
	}
	for k, v := range parentIdMap {
		key := k
		parentOrder := 0
		parentHidden := false
		if *v == *parentHideOrderTags {
			parentOrder = order
			parentHidden = true
		}
		if err = t.ts.CreateParentTagMapping(tx, &domain.ParentTagMapping{TagID: tags.ID, TagType: tagData.Type, ParentTagType: &key, Order: &parentOrder, ParentTagID: v, Hidden: parentHidden, Publish: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
			rollback()
			_ = tx.Rollback()
			return nil, err
		}
	}
	for k := range parentIdHiddenMap {
		key := k
		if err = t.ts.ToggleHideParentTagMapping(tx, false, tags.ID, &key); err != nil {
			rollback()
			_ = tx.Rollback()
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		rollback()
		return nil, noonerror.New(noonerror.ErrInternalServer, "dbCommitError")
	}
	hidden := true
	if orderHierarchyPresent {
		hidden = false
	}
	tagResponse = &domain.TagResponse{
		ID:     tagData.ID,
		Type:   tagData.Type,
		Name:   tagData.Name,
		Hidden: hidden,
	}
	return tagResponse, nil
}

func (t *AdminTagsServiceStruct) removeCurriculumTagHierarchy(tags *domain.RemoveHierarchy) (tagResponse *domain.TagResponse, err error) {
	ctx := context.Background()
	tx, err := repository.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "ContextCreationError")
	}
	if len(tags.Hierarchy) == 0 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "hierarchyAbsent")
	}
	tagData, err := t.ts.FetchTags(tags.ID)
	if err != nil || tagData == nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagFetchError")
	}
	if *tags.TagGroup != tagData.TagGroup {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagGroupMismatch")
	}
	tagHierarchySlice, err := t.ts.GetTagsConcurrent(tags.Hierarchy)
	if err != nil {
		return
	}
	curriculumType := tags.CurriculumType
	curriculumHierarchy, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return
	}
	tagHierarchy, ok := curriculumHierarchy[*tagData.Type]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	if tagHierarchy.Level == 1 {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeInvalid")
	}
	parentTags, err := verifyAndFetchParentCurriculumTags(curriculumType, tagHierarchySlice, tagHierarchy.Level)
	if err != nil || parentTags == nil {
		return
	}
	var allParents []*string
	parentTagMapping, err := t.ts.FetchParentTagMappingByParentTagIdTagId(tags.ID, parentTags)
	if err != nil {
		return
	}
	if parentTagMapping != nil {
		if err = t.ts.ToggleHideParentTagMapping(tx, true, tags.ID, parentTagMapping.ID); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		allParents = append(allParents, parentTags)
	}
	if err = t.es.HideParentTags(tags.ID, allParents); err != nil {
		go func() {
			_ = t.es.AddParentTags(tags.ID, allParents)
		}()
		_ = tx.Rollback()
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		go func() {
			_ = t.es.AddParentTags(tags.ID, allParents)
		}()
		return nil, noonerror.New(noonerror.ErrInternalServer, "dbCommitError")
	}
	tagResponse = &domain.TagResponse{
		ID:     tagData.ID,
		Type:   tagData.Type,
		Name:   tagData.Name,
		Hidden: true,
	}
	return tagResponse, nil
}

func (t *AdminTagsServiceStruct) removeContentTagHierarchy(tags *domain.RemoveHierarchy) (tagResponse *domain.TagResponse, err error) {
	ctx := context.Background()
	tx, err := repository.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "ContextCreationError")
	}
	tagData, err := t.ts.FetchTags(tags.ID)
	if err != nil || tagData == nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagFetchError")
	}
	if *tags.TagGroup != tagData.TagGroup {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagGroupMismatch")
	}
	tagHierarchySlice, err := t.ts.GetTagsConcurrent(tags.Hierarchy)
	if err != nil {
		return
	}
	var identifierSlice []*domain.Tags
	if len(tags.Identifier) > 0 {
		identifierSlice, err = t.ts.GetTagsConcurrent(tags.Identifier)
		if err != nil {
			return
		}
	}
	curriculumType := tags.CurriculumType
	curriculumHierarchy, err := flow.GetCurriculum(curriculumType)
	if err != nil {
		return
	}
	parentTagMappings, err := t.ts.FetchParentTagMappings(tags.ID)
	if err != nil {
		return
	}
	_, ok := curriculumHierarchy[*tagData.Type]
	if !ok {
		return nil, noonerror.New(noonerror.ErrBadRequest, "invalidContent")
	}
	var allParents []*string
	if len(tagHierarchySlice) > 0 {
		_, parentHideOrderTags, _, err := verifyAndFetchParentCurriculumTagsForContent(curriculumType, tagData.Type, tagHierarchySlice, constant.WriteAccessType)
		if err != nil {
			return nil, err
		}
		for _, parentTagMapping := range parentTagMappings {
			if *parentTagMapping.ParentTagID == *parentHideOrderTags {
				allParents = append(allParents, parentTagMapping.ParentTagID)
				err = t.ts.ToggleHideParentTagMapping(tx, true, tags.ID, parentTagMapping.ID)
				if err != nil {
					_ = tx.Rollback()
					return nil, err
				}
			}
		}
	}
	if len(identifierSlice) > 0 {
		for _, v := range identifierSlice {
			for _, parentTagMapping := range parentTagMappings {
				if *parentTagMapping.ParentTagType == *v.Type && *parentTagMapping.ParentTagID == *v.ID {
					allParents = append(allParents, parentTagMapping.ParentTagID)
					err = t.ts.ToggleHideParentTagMapping(tx, true, tags.ID, parentTagMapping.ID)
					if err != nil {
						_ = tx.Rollback()
						return
					}
				}
			}
		}
	}
	err = t.es.HideParentTags(tags.ID, allParents)
	if err != nil {
		go func() {
			_ = t.es.AddParentTags(tags.ID, allParents)
		}()
		_ = tx.Rollback()
		return
	}
	if err = tx.Commit(); err != nil {
		go func() {
			_ = t.es.AddParentTags(tags.ID, allParents)
		}()
		return nil, noonerror.New(noonerror.ErrInternalServer, "dbCommitError")
	}
	tagResponse = &domain.TagResponse{
		ID:     tagData.ID,
		Type:   tagData.Type,
		Name:   tagData.Name,
		Hidden: true,
	}
	return tagResponse, nil
}

func (t *AdminTagsServiceStruct) getCurriculumTags(tags *domain.GetTags) (getTagResponse *domain.GetTagsResponse, err error) {
	var parents []*string
	if tags.Hierarchy != nil {
		parents = append(parents, tags.Hierarchy)
	}
	createElasticEntity, err := dtomapper.GetElasticTagEntity(tags, nil, parents, "", tags.CurriculumType, "admin", tags.TagGroup, 0, 100)
	if err != nil {
		return
	}
	filteredTags, next, err := t.es.GetTags(createElasticEntity)
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
	if *tags.MultiGrade == "false" {
		tagData = filterMultiGradeTags(tagData)
	}
	tagData, err = t.ts.FetchTagLocaleMappingsByLocale(tagData, tags.CountryId, tags.Locale)
	if err != nil {
		return
	}
	tagData, err = t.ts.OrderTags(tagData, tags.Type, tags.CurriculumType, tags.Hierarchy)
	if err != nil {
		return
	}
	getTagResponse, _ = dtomapper.CreateGetTagResponse(tagData, tags, hiddenSet, next)
	return getTagResponse, nil
}

func (t *AdminTagsServiceStruct) getContentTags(tags *domain.GetTags) (getTagResponse *domain.GetTagsResponse, err error) {
	set := make(map[string][]*string)
	var identifierTags []*string
	var parents []*string
	var parentsHidden []*string
	setIdentifiers := make(map[string]*domain.Tags)
	if tags.Hierarchy != nil {
		parentsHidden = append(parentsHidden, tags.Hierarchy)
	}
	var identifierSlice []*domain.Tags
	if tags.Identifier != nil && len(tags.Identifier) > 0 {
		identifierSlice, err = t.ts.GetTagsConcurrent(tags.Identifier)
		if err != nil {
			return
		}
	}
	for _, v := range identifierSlice {
		if v != nil {
			parents = append(parents, v.ID)
		}
	}
	createElasticEntity, err := dtomapper.GetElasticTagEntity(tags, parents, parentsHidden, "", tags.CurriculumType, "admin", tags.TagGroup, 0, 100)
	if err != nil {
		return
	}
	tagIds, next, err := t.es.GetTags(createElasticEntity)
	if err != nil {
		return
	}
	parentTagMappingData, err := t.ts.FetchByInParentTagMappings(tagIds)
	if err != nil {
		return
	}
	for _, v := range parentTagMappingData {
		if *v.ParentTagType != constant.DerivedCurriculum && *v.ParentTagType != constant.HierarchyCurriculum {
			set[*v.TagID] = append(set[*v.TagID], v.ParentTagID)
		}
		if *v.ParentTagID == *tags.Hierarchy {
			hidden := "false"
			if v.Hidden {
				hidden = "true"
			}
			set[*v.TagID] = append(set[*v.TagID], &hidden)
		}
		if !strings.Contains(*v.ParentTagID, ".") {
			setIdentifiers[*v.ParentTagID] = new(domain.Tags)
		}
	}
	for k := range setIdentifiers {
		key := k
		identifierTags = append(identifierTags, &key)
	}
	identifierData, err := t.ts.FetchByInTags(identifierTags)
	for _, v := range identifierData {
		setIdentifiers[*v.ID] = v
	}
	if err != nil {
		return
	}
	tagData, err := t.ts.FetchByInTags(tagIds)
	if err != nil {
		return
	}
	tagData, err = t.ts.OrderTags(tagData, tags.Type, tags.CurriculumType, tags.Hierarchy)
	if err != nil {
		return
	}
	getTagResponse, _ = dtomapper.CreateGetTagResponseWithIdentifiers(tagData, tags, setIdentifiers, set, next)
	return getTagResponse, nil
}

func (t *AdminTagsServiceStruct) getCurriculumTagsSearch(tags *domain.GetTags) (getTagResponse *domain.GetTagsResponse, err error) {
	if tags.Type == nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeAbsent")
	}
	var tagData []*domain.Tags
	var parents []*string
	if tags.Hierarchy != nil {
		parents = append(parents, tags.Hierarchy)
	}
	var tagIds []*string
	var next *int
	if tags.Text == nil {
		createElasticEntity, err := dtomapper.GetElasticTagEntity(tags, parents, nil, "", tags.CurriculumType, "admin", tags.TagGroup, tags.Start, tags.Limit)
		if err != nil {
			return nil, err
		}
		tagIds, next, err = t.es.GetTags(createElasticEntity)
		if err != nil {
			return nil, err
		}
	} else {
		createElasticEntity, err := dtomapper.GetElasticTagEntity(tags, parents, nil, "", tags.CurriculumType, "admin", tags.TagGroup, tags.Start, tags.Limit)
		if err != nil {
			return nil, err
		}
		tagIds, next, err = t.es.GetTagsSearch(createElasticEntity)
		if err != nil {
			return nil, err
		}
	}
	tagData, err = t.ts.FetchByInTags(tagIds)
	if err != nil {
		return nil, err
	}
	tagData, err = t.ts.FetchTagLocaleMappingsByLocale(tagData, tags.CountryId, tags.Locale)
	if err != nil {
		return
	}
	getTagResponse, _ = dtomapper.CreateGetTagResponse(tagData, tags, map[string]bool{}, next)
	return getTagResponse, nil
}

func (t *AdminTagsServiceStruct) getContentTagsSearch(tags *domain.GetTags) (getTagResponse *domain.GetTagsResponse, err error) {
	if tags.Type == nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "tagTypeAbsent")
	}
	if tags.Hierarchy == nil {
		return nil, noonerror.New(noonerror.ErrBadRequest, "hierarchyAbsent")
	}
	var parents []*string
	if tags.Hierarchy != nil {
		parents = append(parents, tags.Hierarchy)
	}
	if len(tags.Identifier) > 0 {
		for _, v := range tags.Identifier {
			parents = append(parents, v)
		}
	}
	var tagIds []*string
	var next *int
	if tags.Text == nil {
		createElasticEntity, err := dtomapper.GetElasticTagEntity(tags, parents, nil, "", tags.CurriculumType, "admin", tags.TagGroup, tags.Start, tags.Limit)
		if err != nil {
			return nil, err
		}
		tagIds, next, err = t.es.GetTags(createElasticEntity)
		if err != nil {
			return nil, err
		}
	} else {
		createElasticEntity, err := dtomapper.GetElasticTagEntity(tags, parents, nil, "", tags.CurriculumType, "admin", tags.TagGroup, tags.Start, tags.Limit)
		if err != nil {
			return nil, err
		}
		tagIds, next, err = t.es.GetTagsSearch(createElasticEntity)
		if err != nil {
			return nil, err
		}
	}
	tagData, err := t.ts.FetchByInTags(tagIds)
	if err != nil {
		return
	}
	getTagResponse, _ = dtomapper.CreateGetTagResponse(tagData, tags, map[string]bool{}, next)
	return getTagResponse, nil
}

func (t *AdminTagsServiceStruct) getIdentifierTagsSearch(tags *domain.GetTags) (getTagResponse *domain.GetTagsResponse, err error) {
	identifierTagGroup := domain.TagGroupEnum.Identifier
	var tagData []*domain.Tags
	createElasticEntity, err := dtomapper.GetElasticTagEntity(tags, nil, nil, "", tags.CurriculumType, "admin", &identifierTagGroup, tags.Start, tags.Limit)
	if err != nil {
		return
	}
	tagIds, next, err := t.es.GetTagsSearch(createElasticEntity)
	if err != nil {
		tagData, err = t.ts.FetchByTagGroup(&identifierTagGroup, tags.Type)
		if err != nil {
			return nil, err
		}
	} else {
		tagData, err = t.ts.FetchByInTags(tagIds)
		if err != nil {
			return nil, err
		}
	}
	getTagResponse, _ = dtomapper.CreateGetTagResponse(tagData, tags, map[string]bool{}, next)
	return getTagResponse, nil
}

func (t *AdminTagsServiceStruct) getCurriculumTagsForLibrary(gtt *domain.GetAdminTags) (*domain.GetTagsResponse, error) {

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

	getTag := domain.GetAdminTags{Text: gtt.Text, TagGroup: gtt.TagGroup, Type: gtt.Type}
	parents := []*string{parentTags}

	createElasticEntity, err := dtomapper.GetElasticTagEntity(getTag, []*string{}, parents, "", gtt.CurriculumType, "admin", gtt.TagGroup, gtt.Start, gtt.Limit)
	if err != nil {
		return nil, err
	}
	filteredTags, next, err := t.es.GetTags(createElasticEntity)
	if err != nil {
		return nil, err
	}
	hiddenSet := make(map[string]bool)
	parentTagMappingData, err := t.ts.FetchByInParentTagMappingsByParentTagIdTagIds(filteredTags, parentTags)
	if err != nil {
		return nil, err
	}
	for _, v := range parentTagMappingData {
		if *v.ParentTagID == *parentTags {
			hiddenSet[*v.TagID] = v.Hidden
		}
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
	getTagResponse, _ := dtomapper.GetTagResponse(tagData, gtt.Type, gtt.CurriculumType, hiddenSet, next)
	return getTagResponse, nil
}

func (t *AdminTagsServiceStruct) getContentTagsForLibrary(gtt *domain.GetAdminTags) (*domain.GetTagsResponse, error) {
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

	getTag := domain.GetAdminTags{Text: gtt.Text, TagGroup: gtt.TagGroup, Type: gtt.Type}
	parents := []*string{parentHideOrderTags}
	adminElasticEntity, err := dtomapper.GetElasticTagEntity(getTag, []*string{}, parents, "", gtt.CurriculumType, "admin", getTag.TagGroup, getTag.Start, getTag.Limit)
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

	hiddenSet := make(map[string]bool)
	parentTagMappingData, err := t.ts.FetchByInParentTagMappingsByParentTagIdTagIds(tagIds, parentHideOrderTags)
	if err != nil {
		return nil, err
	}
	for _, v := range parentTagMappingData {
		if *v.ParentTagID == *parentHideOrderTags {
			hiddenSet[*v.TagID] = v.Hidden
		}
	}
	tagData, err = t.ts.OrderTags(tagData, gtt.Type, gtt.CurriculumType, parentHideOrderTags)
	if err != nil {
		return nil, err
	}

	getTagResponse, _ := dtomapper.GetTagResponse(tagData, gtt.Type, gtt.CurriculumType, hiddenSet, next)
	return getTagResponse, nil
}

func (t *AdminTagsServiceStruct) GetTestsSkillsForLibrary(gtt *domain.GetAdminTags) (*domain.GetTagsResponse, error) {

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

	getTag := domain.GetAdminTags{Text: gtt.Text, TagGroup: gtt.TagGroup, Type: gtt.Type}
	parents := []*string{parentTags}

	createElasticEntity, err := dtomapper.GetElasticTagEntityWithoutCurriculumType(getTag, nil, parents, "", "admin", gtt.TagGroup, gtt.Start, gtt.Limit)
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
	hiddenSet := make(map[string]bool)
	parentTagMappingData, err := t.ts.FetchByInParentTagMappingsByParentTagIdTagIds(filteredTags, parentTags)
	if err != nil {
		return nil, err
	}
	for _, v := range parentTagMappingData {
		if *v.ParentTagID == *parentTags {
			hiddenSet[*v.TagID] = v.Hidden
		}
	}
	tagData, err = t.ts.FetchTagLocaleMappingsByLocale(tagData, gtt.CountryId, gtt.Locale)
	if err != nil {
		return nil, err
	}
	tagData, err = t.ts.OrderTags(tagData, gtt.Type, curriculumType, parentTags)
	if err != nil {
		return nil, err
	}
	getTagResponse, _ := dtomapper.GetTagResponse(tagData, gtt.Type, nil, hiddenSet, next)
	return getTagResponse, nil
}

func (t *AdminTagsServiceStruct) GetCountriesTagsNew(tags *domain.GetCountriesNew) (getTagResponse *domain.GetCountriesNewResponse, err error) {
	rootCurriculumType := domain.CurriculumTypeEnum.Root
	countryType := domain.TagTypeEnum.Country
	filteredTagsResponse, err := t.ts.FetchFilteredTagsPaginatedForAdmin(&rootCurriculumType, &countryType, &tags.Start, &tags.Limit)
	if err != nil {
		return
	}
	next := -1
	if len(filteredTagsResponse) >= tags.Limit {
		next = tags.Start + tags.Limit
	}
	getTagResponse, _ = dtomapper.CreateGetCountriesNewResponse(filteredTagsResponse, tags.Locale, nil, &next, true)
	return getTagResponse, nil
}

func filterMultiGradeTags(tagData []*domain.Tags) []*domain.Tags {
	var tagDataWithoutMultiGrades []*domain.Tags
	for _, tag := range tagData {
		_, ok := tag.Attributes["multi_grade"]
		if !ok {
			tagDataWithoutMultiGrades = append(tagDataWithoutMultiGrades, tag)
		}
	}
	if len(tagDataWithoutMultiGrades) > 0 {
		tagData = tagDataWithoutMultiGrades
	}
	return tagData
}
