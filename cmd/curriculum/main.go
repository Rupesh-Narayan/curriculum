package main

import (
	"bitbucket.org/noon-go/auth"
	translation "bitbucket.org/noon-go/translator"
	"bitbucket.org/noon-micro/curriculum/config"
	"bitbucket.org/noon-micro/curriculum/pkg/external"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/helper"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/httplib"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/middleware"
	"bitbucket.org/noon-micro/curriculum/pkg/repository/mysql"
	redis "bitbucket.org/noon-micro/curriculum/pkg/repository/redis"
	"bitbucket.org/noon-micro/curriculum/pkg/resource"
	"bitbucket.org/noon-micro/curriculum/pkg/service"
	"github.com/gorilla/handlers"
	"net/http"
	"strconv"
	"time"

	"os"

	"bitbucket.org/noon-go/noonhttp"
	"github.com/sirupsen/logrus"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func main() {
	//This is to extract the command line arguments
	if len(os.Args) < 2 {
		logrus.Info("Please start service with one of environment name : local, qa or prod")
		return
	}
	//Load Configuration File
	settingsFileName := os.Args[1]

	allowedHeaders := []string{"country", "x-client-time", "browser", "locale", "Authorization", "Accept", "Content-Type", "timezone" ,
		"Referer", "platform", "User-Agent", "device-details", "api-version", "os-details", "x-device-id", "resolution", "device_details", "os_details"}

	allowedMethods := []string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}

	logger.New()

	configFile := config.LoadConfiguration(settingsFileName)
	if configFile.DataDogDEnabled == "true" {
		if configFile.DataDogAgentHost != "" {
			dataDogURL := configFile.DataDogAgentHost
			logger.Client.Info("Connecting to DataDog Url : " + dataDogURL)
			tracer.WithServiceName("curriculum")
			tracer.WithAgentAddr(dataDogURL)
			tracer.WithEnv(configFile.DataDogEnv)
			tracer.WithServiceVersion(configFile.DataDogVersion)
			tracer.Start()
			defer tracer.Stop()
		}
	}

	requestTimeout := 2
	if configFile.RequestTimeout != "" {
		timeout, _ := strconv.Atoi(configFile.RequestTimeout)
		if timeout > 0 {
			requestTimeout = timeout
		}
	}

	helper.InitializeValidator()
	httpClient := noonhttp.Initialize(noonhttp.Config{Timeout: time.Duration(requestTimeout) * time.Second})
	noonAuthenticateEntity := auth.AuthenticateEntity{
		Client:      httpClient,
		BifrostHost: configFile.AuthHost,
	}
	httplib.InitializeHttp(httpClient)
	translation.Initialize(httpClient, configFile.TranslationHost)
	middleware.InitializeMiddleware(&noonAuthenticateEntity)
	repo := repository.InitializeMysql(configFile)
	redis.InitializeRedisClient(configFile.RedisHost, configFile.RedisPort)
	elastic := external.NewElasticExternal(httplib.Client)
	tagsService := service.NewTagsService(repo.Tags, repo.ParentTagMapping, repo.TagLocaleMapping, repo.LegacyTagMapping, repo.GradeProduct)
	adminTagsService := service.NewAdminTagsService(tagsService, elastic)
	geo := external.NewGeoIpExternal(httplib.Client)
	studentTagsService := service.NewStudentTagsService(tagsService, elastic, geo)
	rpcTagsService := service.NewRpcTagsService(tagsService, elastic)
	teacherTagsService := service.NewTeacherTagsService(tagsService, elastic, geo)
	r := httptrace.NewRouter(httptrace.WithServiceName("curriculum")).StrictSlash(false)
	mainRoutes := r.PathPrefix("/curriculum/v1/").Subrouter()
	//resource.NewTagsResource(mainRoutes, tagsService)
	resource.NewAdminTagsResource(mainRoutes, adminTagsService)
	resource.NewStudentTagsResource(mainRoutes, studentTagsService)
	resource.NewRpcTagsResource(r.Router, rpcTagsService)
	resource.NewHealthResource(r.Router, repo.Db)
	resource.NewTeacherTagsResource(mainRoutes, teacherTagsService)
	logger.Client.Info("Http Server Listens On Public Port " + configFile.PublicAppPort)
	logger.Client.Fatal(http.ListenAndServe(":"+configFile.PublicAppPort, handlers.CORS(
		handlers.AllowedHeaders(allowedHeaders),
		handlers.AllowedMethods(allowedMethods),
		handlers.AllowedOrigins([]string{"*"}))(r)))
}
