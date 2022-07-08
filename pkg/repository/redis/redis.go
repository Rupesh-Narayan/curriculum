package repository

import (
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	redistrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-redis/redis"
	"time"
)

// RedisClient pointer
var RedisClient *redistrace.Client

const (
	CurriculumPrefix                 string = "curriculum:tag:"
	CurriculumCountryPrefix          string = "curriculum:country:"
	CurriculumCountryAdminPrefix     string = "curriculum:country:admin:"
	CurriculumTagOrderPrefix         string = "curriculum:tag_order:"
	CurriculumGradeProductPrefix     string = "curriculum:grade_product:"
	CurriculumParentTagMappingPrefix string = "curriculum:parent_tag_mapping:"
	CurriculumTagLocaleMappingPrefix string = "curriculum:tag_locale_mapping:"
	CurriculumMultiGradePrefix       string = "curriculum:multi_grade:"
	MultiGradeTtl                           = 30 * time.Minute
	RedisTtl                                = 24 * time.Hour
)

// InitializeRedisClient Initialize redis client
func InitializeRedisClient(host string, port string) {
	contextLogger := logger.Client.WithFields(logrus.Fields{
		"host": host,
		"port": port,
	})
	url := host
	if port != "" {
		url = url + ":" + port
	}

	options := &redis.Options{Addr: url}
	RedisClient = redistrace.NewClient(options, redistrace.WithServiceName("curriculum-redis"))

	_, err := RedisClient.Ping().Result()
	if err != nil {
		contextLogger.Error("Redis ping failed and returned err : " + err.Error())
		//panic(err)
		return
	}

	contextLogger.Info("Redis Client Created Successfully For Url " + url)
}
