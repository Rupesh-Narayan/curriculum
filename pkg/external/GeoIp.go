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

type GeoIpStruct struct {
	client *noonhttp.ClientEntity
}

const (
	getGeoIp = "/rpc/add_curriculum_tag"
)

func NewGeoIpExternal(client *noonhttp.ClientEntity) *GeoIpStruct {
	return &GeoIpStruct{client: client}
}

func (e *GeoIpStruct) GetGeoIp(getGeoIpRequest *domain.GetGeoIp) (countryCode *string, err error) {
	if *getGeoIpRequest.Ip == "" {
		countryCode := "SA"
		return &countryCode, nil
	}
	t1 := helper.MakeTimestamp()
	out, err := json.Marshal(getGeoIpRequest)
	if err != nil {
		logger.Client.Error("getGeoIpRequestError", err, logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "getGeoIpRequestError")
	}
	contextLogger := logger.Client.WithFields(logrus.Fields{
		"getGeoIpRequestDto": string(out),
	})
	contextLogger.Info("get Tags Elastic Request")
	strPointerValue := *getGeoIpRequest.Ip
	var pathParam = "/" + strPointerValue
	url := config.GetConfig().GeoIpHost + pathParam
	payload := make(map[string]string)
	payload["access_key"] = "e5bf2c82292aa79f569d3aec5a040bc7"
	payload["format"] = "1"
	if err != nil {
		countryCode := "SA"
		return &countryCode, nil
	}
	resp, err := e.client.ServeGet(url, getHeaders1(), payload)
	if err != nil {
		logger.Client.Error("hideParentTagsElasticError", err, logger.GetErrorStack())
		countryCode := "SA"
		return &countryCode, nil
	}
	contextLogger.Info("get Tags Elastic Response", string(resp))
	respBody := make(map[string]interface{})
	err = json.Unmarshal(resp, &respBody)
	if err != nil {
		//logger.Client.Error("getTagsElasticError", err, logger.GetErrorStack())
		countryCode := "SA"
		return &countryCode, nil
	}
	if respBody["country_code"] != nil {
		countryCode := respBody["country_code"].(string)
		return &countryCode, nil
	}
	t2 := helper.MakeTimestamp()
	contextLogger.Info("Time Taken: ", strconv.FormatInt(t2-t1, 10))
	return
}

func getHeaders1() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}
