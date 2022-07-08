package config

// Configuration main struct
type Configuration struct {
	AuthHost               string
	TranslationHost        string
	RequestTimeout         string
	MySqlHost              string
	MySqlDatabaseName      string
	MySqlUserName          string
	MySqlPassword          string
	MySqlMaxOpenConns      int
	MySqlMaxIdleConns      int
	MySqlConnMaxLifetime   int
	RedisHost              string
	RedisPort              string
	PublicAppPort          string
	DataDogDEnabled        string
	DataDogAgentHost       string
	DataDogVersion         string
	DataDogEnv             string
	ElasticHost            string
	GeoIpHost              string
	MiscTagId              string
	ResourceTagId          string
	BoardTagId             string
	DegreeTagId            string
	MajorTagId             string
	CourseTagId            string
	UniversitySectionTagId string
	DefaultColor           string
	DefaultPic             string
}

type Config interface {
	ConfigManager() *Configuration
}

var loadedConfiguration *Configuration

// GetConfig from any class
func GetConfig() *Configuration {
	return loadedConfiguration
}

// LoadConfiguration for all env
func LoadConfiguration(settings string) *Configuration {

	conf := Configuration{}
	var config Config
	switch settings {
	case "local":
		config = &LocalConfig{&conf}
	default:
		config = &ServerConfig{&conf}
	}
	config.ConfigManager()
	loadedConfiguration = &conf
	return &conf

}
