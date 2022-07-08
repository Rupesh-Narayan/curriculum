package repository

// Initialize mysql connection pointer
import (
	"bitbucket.org/noon-micro/curriculum/config"
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"database/sql"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/go-sql-driver/mysql"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
)

var Db *sql.DB

type Repositories struct {
	Tags             domain.TagsRepository
	TagLocaleMapping domain.TagLocaleMappingRepository
	ParentTagMapping domain.ParentTagMappingRepository
	LegacyTagMapping domain.LegacyTagMappingRepository
	GradeProduct     domain.GradeProductRepository
	Db               *sql.DB
}

func InitializeMysql(config *config.Configuration) *Repositories {
	contextLogger := logger.Client.WithFields(logrus.Fields{
		"host":   config.MySqlHost,
		"dbName": config.MySqlDatabaseName,
	})
	uri := config.MySqlUserName + ":" + config.MySqlPassword + "@tcp(" + config.MySqlHost + ")/" + config.MySqlDatabaseName + "?charset=utf8"
	sqltrace.Register("mysql", mysql.MySQLDriver{}, sqltrace.WithServiceName("curriculum-mysql"))

	db, err := sqltrace.Open("mysql", uri)
	if err != nil {
		contextLogger.Error(logger.GetErrorStack(), "MySql connection failed", err)
		panic(err)
	}
	db.SetMaxOpenConns(config.MySqlMaxOpenConns)
	db.SetMaxIdleConns(config.MySqlMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(config.MySqlConnMaxLifetime) * time.Second)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	contextLogger.Info("MySql connected successfully")
	Db = db
	return &Repositories{
		Tags:             NewTagsRepository(db),
		TagLocaleMapping: NewTagLocaleMappingRepository(db),
		ParentTagMapping: NewParentTagMappingRepository(db),
		LegacyTagMapping: NewLegacyTagMappingRepository(db),
		GradeProduct:     NewGradeProductRepository(db),
		Db:               db,
	}
}
