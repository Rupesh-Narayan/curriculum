package external

import (
	"bitbucket.org/noon-go/noonhttp"
	"bitbucket.org/noon-micro/curriculum/config"
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	noonerror "bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/helper"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"strconv"
)

type ElasticStruct struct {
	client *noonhttp.ClientEntity
}

const (
	createTagURL        = "/rpc/add_curriculum_tag"
	getTagsURL          = "/rpc/get_tags"
	searchTagsURL       = "/rpc/search_tags"
	updateTagURL        = "/rpc/update_curriculum_tag"
	addParentTagsURL    = "/rpc/manage_parent_tags"
	removeParentTagsURL = "/rpc/manage_parent_tags"
	hideParentTagsURL   = "/rpc/manage_parent_tags"
)

func NewElasticExternal(client *noonhttp.ClientEntity) *ElasticStruct {
	return &ElasticStruct{client: client}
}

func (e *ElasticStruct) CreateTag(createTagElastic *domain.CreateTagElastic) (err error) {
	t1 := helper.MakeTimestamp()
	out, err := json.Marshal(createTagElastic)
	if err != nil {
		logger.Client.Error("createTagElasticError", err, logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "createTagMarshallingError")
	}
	contextLogger := logger.Client.WithFields(logrus.Fields{
		"createTagDTO": string(out),
	})
	contextLogger.Info("create Tag Elastic Request")
	url := config.GetConfig().ElasticHost + createTagURL
	payload := make(map[string]interface{})
	err = json.Unmarshal(out, &payload)
	if err != nil {
		logger.Client.Error("createTagElasticError", err, logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "elasticPayloadMappingError")
	}
	resp, err := e.client.ServePost(url, getHeaders(), payload)
	if err != nil {
		logger.Client.Error("createTagElasticError", err, logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "createTagElasticError")
	}
	contextLogger.Info("create Tag Elastic Response", resp)
	t2 := helper.MakeTimestamp()
	contextLogger.Info("Time Taken: ", strconv.FormatInt(t2-t1, 10))
	return
}

func (e *ElasticStruct) GetTags(getTagsElastic *domain.GetTagsElastic) (tags []*string, next *int, err error) {
	if getTagsElastic.Text != nil {
		return e.GetTagsSearch(getTagsElastic)
	}
	t1 := helper.MakeTimestamp()
	if getTagsElastic.Limit == 0 {
		getTagsElastic.Limit = 100
	}
	out, err := json.Marshal(getTagsElastic)
	if err != nil {
		logger.Client.Error("getTagsElasticError", err, logger.GetErrorStack())
		return nil, nil, noonerror.New(noonerror.ErrInternalServer, "getTagsMarshallingError")
	}
	contextLogger := logger.Client.WithFields(logrus.Fields{
		"getTagsDTO": string(out),
	})
	contextLogger.Info("get Tags Elastic Request")
	url := config.GetConfig().ElasticHost + getTagsURL
	payload := make(map[string]interface{})
	err = json.Unmarshal(out, &payload)
	if err != nil {
		logger.Client.Error("getTagsElasticError", err, logger.GetErrorStack())
		return nil, nil, noonerror.New(noonerror.ErrInternalServer, "elasticPayloadMappingError")
	}
	resp, err := e.client.ServePost(url, getHeaders(), payload)
	if err != nil {
		logger.Client.Error("getTagsElasticError", err, logger.GetErrorStack())
		return nil, nil, noonerror.New(noonerror.ErrInternalServer, "getTagsElasticError")
	}
	contextLogger.Info("get Tags Elastic Response", string(resp))
	respBody := make(map[string]interface{})
	err = json.Unmarshal(resp, &respBody)
	if err != nil {
		logger.Client.Error("getTagsElasticError", err, logger.GetErrorStack())
		return nil, nil, noonerror.New(noonerror.ErrInternalServer, "elasticPayloadMappingError")
	}
	if v, ok := respBody["data"]; ok {
		var datas []interface{}
		if v != nil {
			datas, ok = v.([]interface{})
			if !ok {
				return nil, nil, noonerror.New(noonerror.ErrInternalServer, "getTagsResponseError")
			}
		}
		for _, v1 := range datas {
			if v1 == nil {
				continue
			}
			floatId, ok := v1.(float64)
			if !ok {
				return nil, nil, noonerror.New(noonerror.ErrInternalServer, "getTagsResponseError")
			}
			stringId := strconv.FormatFloat(floatId, 'f', -1, 64)
			tags = append(tags, &stringId)
		}
	}
	if v, ok := respBody["total"]; ok {
		var totalFloat float64
		next := -1
		if v != nil {
			totalFloat, ok = v.(float64)
			if !ok {
				return tags, &next, nil
			}
			totalInt := int(totalFloat)
			if getTagsElastic.Start+getTagsElastic.Limit < totalInt {
				next = getTagsElastic.Start + getTagsElastic.Limit
			}
			return tags, &next, nil
		}
	}
	t2 := helper.MakeTimestamp()
	contextLogger.Info("Time Taken: ", strconv.FormatInt(t2-t1, 10))
	return
}

func (e *ElasticStruct) GetTagsSearch(getTagsElastic *domain.GetTagsElastic) (tags []*string, next *int, err error) {
	t1 := helper.MakeTimestamp()
	if getTagsElastic.Limit == 0 {
		getTagsElastic.Limit = 100
	}
	out, err := json.Marshal(getTagsElastic)
	if err != nil {
		logger.Client.Error("getTagsElasticSearchError", err, logger.GetErrorStack())
		return nil, nil, noonerror.New(noonerror.ErrInternalServer, "getTagsSearchMarshallingError")
	}
	contextLogger := logger.Client.WithFields(logrus.Fields{
		"getTagsDTO": string(out),
	})
	contextLogger.Info("get Tags SearchElastic Request")
	url := config.GetConfig().ElasticHost + searchTagsURL
	payload := make(map[string]interface{})
	err = json.Unmarshal(out, &payload)
	if err != nil {
		logger.Client.Error("getTagsElasticSearchError", err, logger.GetErrorStack())
		return nil, nil, noonerror.New(noonerror.ErrInternalServer, "elasticPayloadMappingError")
	}
	resp, err := e.client.ServePost(url, getHeaders(), payload)
	if err != nil {
		logger.Client.Error("getTagsElasticSearchError", err, logger.GetErrorStack())
		return nil, nil, noonerror.New(noonerror.ErrInternalServer, "getTagsElasticError")
	}
	contextLogger.Info("get Tags Search Elastic Response", string(resp))
	respBody := make(map[string]interface{})
	err = json.Unmarshal(resp, &respBody)
	if err != nil {
		logger.Client.Error("getTagsSearchElasticError", err, logger.GetErrorStack())
		return nil, nil, noonerror.New(noonerror.ErrInternalServer, "elasticPayloadMappingError")
	}
	if v, ok := respBody["data"]; ok {
		var datas []interface{}
		if v != nil {
			datas, ok = v.([]interface{})
			if !ok {
				return nil, nil, noonerror.New(noonerror.ErrInternalServer, "getTagsSearchResponseError")
			}
		}
		for _, v1 := range datas {
			if v1 == nil {
				continue
			}
			floatId, ok := v1.(float64)
			if !ok {
				return nil, nil, noonerror.New(noonerror.ErrInternalServer, "getTagsSearchResponseError")
			}
			stringId := strconv.FormatFloat(floatId, 'f', -1, 64)
			tags = append(tags, &stringId)
		}
	}
	if v, ok := respBody["total"]; ok {
		var totalFloat float64
		next := -1
		if v != nil {
			totalFloat, ok = v.(float64)
			if !ok {
				return tags, &next, nil
			}
			totalInt := int(totalFloat)
			if getTagsElastic.Start+getTagsElastic.Limit < totalInt {
				next = getTagsElastic.Start + getTagsElastic.Limit
			}
			return tags, &next, nil
		}
	}
	t2 := helper.MakeTimestamp()
	contextLogger.Info("Time Taken: ", strconv.FormatInt(t2-t1, 10))
	return
}

func (e *ElasticStruct) UpdateTag(tagId *string, delete *bool, names []*domain.TagName) (err error) {
	t1 := helper.MakeTimestamp()
	contextLogger := logger.Client.WithFields(logrus.Fields{
		"tagId":  tagId,
		"delete": delete,
		"names":  names,
	})
	if delete == nil && len(names) == 0 {
		return
	}
	contextLogger.Info("update Tag Elastic Request")
	url := config.GetConfig().ElasticHost + updateTagURL
	payload := make(map[string]interface{})
	payload["id"] = tagId
	if delete != nil {
		payload["deleted"] = delete
	}
	if len(names) > 0 {
		payload["name"] = names
	}
	resp, err := e.client.ServePost(url, getHeaders(), payload)
	if err != nil {
		logger.Client.Error("updateTagElasticError", err, logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "updateTagElasticError")
	}
	contextLogger.Info("update Tag Elastic Response", resp)
	t2 := helper.MakeTimestamp()
	contextLogger.Info("Time Taken: ", strconv.FormatInt(t2-t1, 10))
	return
}

func (e *ElasticStruct) AddParentTags(tagId *string, parents []*string) (err error) {
	t1 := helper.MakeTimestamp()
	contextLogger := logger.Client.WithFields(logrus.Fields{
		"tagId":   tagId,
		"parents": parents,
	})

	contextLogger.Info("add parents Tags Elastic Request")
	url := config.GetConfig().ElasticHost + addParentTagsURL
	payload := make(map[string]interface{})
	payload["id"] = tagId
	if len(parents) == 0 {
		contextLogger.Info("no parents to add")
		return
	}
	payload["add"] = parents
	resp, err := e.client.ServePost(url, getHeaders(), payload)
	if err != nil {
		logger.Client.Error("addParentTagsElasticError", err, logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "addParentTagsElasticError")
	}
	contextLogger.Info("add parent Tags Elastic Response", resp)
	t2 := helper.MakeTimestamp()
	contextLogger.Info("Time Taken: ", strconv.FormatInt(t2-t1, 10))
	return
}

func (e *ElasticStruct) RemoveParentTags(tagId *string, parents []*string) (err error) {
	t1 := helper.MakeTimestamp()
	contextLogger := logger.Client.WithFields(logrus.Fields{
		"tagId":   tagId,
		"parents": parents,
	})

	contextLogger.Info("remove parents Tags Elastic Request")
	url := config.GetConfig().ElasticHost + removeParentTagsURL
	payload := make(map[string]interface{})
	payload["id"] = tagId
	if len(parents) == 0 {
		contextLogger.Info("no parents to remove")
		return
	}
	payload["remove"] = parents
	resp, err := e.client.ServePost(url, getHeaders(), payload)
	if err != nil {
		logger.Client.Error("removeParentTagsElasticError", err, logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "removeParentTagsElasticError")
	}
	contextLogger.Info("remove parent Tags Elastic Response", resp)
	t2 := helper.MakeTimestamp()
	contextLogger.Info("Time Taken: ", strconv.FormatInt(t2-t1, 10))
	return
}

func (e *ElasticStruct) HideParentTags(tagId *string, parents []*string) (err error) {
	t1 := helper.MakeTimestamp()
	contextLogger := logger.Client.WithFields(logrus.Fields{
		"tagId":   tagId,
		"parents": parents,
	})

	contextLogger.Info("hide parents Tags Elastic Request")
	url := config.GetConfig().ElasticHost + hideParentTagsURL
	payload := make(map[string]interface{})
	payload["id"] = tagId
	if len(parents) == 0 {
		contextLogger.Info("no parents to hide")
		return
	}
	payload["hide"] = parents
	resp, err := e.client.ServePost(url, getHeaders(), payload)
	if err != nil {
		logger.Client.Error("hideParentTagsElasticError", err, logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "hideParentTagsElasticError")
	}
	contextLogger.Info("hide parent Tags Elastic Response", resp)
	t2 := helper.MakeTimestamp()
	contextLogger.Info("Time Taken: ", strconv.FormatInt(t2-t1, 10))
	return
}

func getHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}
