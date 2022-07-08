package config

import (
	"os"
	"strconv"
)

// ServerConfig struct
type ServerConfig struct {
	*Configuration
}

// ConfigManager for prod
func (conf *ServerConfig) ConfigManager() *Configuration {
	conf.AuthHost = os.Getenv("BIFROST_HOST")
	conf.TranslationHost = os.Getenv("TRANSLATION_HOST")
	conf.RequestTimeout = os.Getenv("API_TIMEOUT")
	conf.MySqlPassword = os.Getenv("DB_PASS_WRITE")
	conf.MySqlHost = os.Getenv("DB_HOST_WRITE")
	conf.MySqlDatabaseName = os.Getenv("DB_NAME")
	conf.MySqlUserName = os.Getenv("DB_USER_WRITE")
	mySqlMaxOpenConns, _ := strconv.Atoi("DB_MAX_OPEN_CONNECTIONS")
	conf.MySqlMaxOpenConns = mySqlMaxOpenConns
	mySqlMaxIdleConns, _ := strconv.Atoi("DB_MAX_IDLE_CONNECTIONS")
	conf.MySqlMaxIdleConns = mySqlMaxIdleConns
	mySqlConnMaxLifetime, _ := strconv.Atoi("DB_CONNECTIONS_MAX_LIFETIME")
	conf.MySqlConnMaxLifetime = mySqlConnMaxLifetime
	conf.RedisHost = os.Getenv("REDIS_HOST")
	conf.RedisPort = os.Getenv("REDIS_PORT")
	conf.PublicAppPort = os.Getenv("PORT")
	conf.DataDogDEnabled = os.Getenv("DD_ENABLED")
	conf.DataDogAgentHost = os.Getenv("DD_AGENT_HOST")
	conf.DataDogVersion = os.Getenv("DD_VERSION")
	conf.DataDogEnv = os.Getenv("DD_ENV")
	conf.ElasticHost = os.Getenv("ELASTIC_HOST")
	conf.MiscTagId = os.Getenv("MISC_TAG_ID")
	conf.ResourceTagId = os.Getenv("RESOURCE_TAG_ID")
	conf.BoardTagId = os.Getenv("BOARD_TAG_ID")
	conf.DegreeTagId = os.Getenv("DEGREE_TAG_ID")
	conf.MajorTagId = os.Getenv("MAJOR_TAG_ID")
	conf.CourseTagId = os.Getenv("COURSE_TAG_ID")
	conf.UniversitySectionTagId = os.Getenv("UNIVERSITY_SECTION_TAG_ID")
	conf.GeoIpHost = "http://api.ipstack.com"
	conf.DefaultColor = os.Getenv("DEFAULT_COLOR")
	conf.DefaultPic = os.Getenv("DEFAULT_PIC")
	return conf.Configuration
}
