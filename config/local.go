package config

// LocalConfig struct
type LocalConfig struct {
	*Configuration
}

// ConfigManager to mange the config details for Locals
func (conf *LocalConfig) ConfigManager() *Configuration {
	conf.AuthHost = "http://bifrost.prod-rpc.non.sa"
	conf.TranslationHost = "http://translations.prod-rpc.non.sa"
	conf.RequestTimeout = "10000"
	conf.MySqlHost = "aurora-prod-micro-cluster.cluster-ro-cres8iqjkrdw.eu-central-1.rds.amazonaws.com:3306"
	conf.MySqlDatabaseName = "folders_srv"
	conf.MySqlUserName = "bhavik_reader"
	conf.MySqlPassword = "6WSUU!&Kv4b@!j2M"
	conf.MySqlMaxOpenConns = 25
	conf.MySqlMaxIdleConns = 25
	conf.MySqlConnMaxLifetime = 300
	conf.RedisHost = "redis-qa-001.fdmtkw.0001.euw1.cache.amazonaws.com."
	conf.RedisPort = "6379"
	conf.PublicAppPort = "8002"
	conf.ElasticHost = "http://elastic.prod-rpc.non.sa"
	conf.GeoIpHost = "http://api.ipstack.com"
	conf.MiscTagId = "22678"
	conf.ResourceTagId = "22677"
	conf.BoardTagId = "24681"
	conf.DegreeTagId = "22667"
	conf.MajorTagId = "22668"
	conf.CourseTagId = "22669"
	conf.UniversitySectionTagId = "22670"
	conf.DefaultColor = "#1A8DFF"
	conf.DefaultPic = "http://cdn.non.sa/product/default.png"
	return conf.Configuration
}
