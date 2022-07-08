package repository

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/converter"
	noonerror "bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"database/sql"
	"strconv"
	"strings"
	"time"
)

type TagLocaleMappingRepo struct {
	db *sql.DB
}

var (
	selectTagLocaleMapping        = "SELECT * FROM tag_locale_mapping WHERE tag_id = ? and publish = 1"
	fetchTagLocaleMappingByLocale = "SELECT * FROM tag_locale_mapping WHERE tag_id = ? and country_id = ? and locale = ? and publish = 1"
	insertTagLocaleMapping        = "INSERT INTO tag_locale_mapping(tag_id, locale, country_id, `name`, publish, tag_type, created_at, updated_at) values(?,?,?,?,?,?,?,?)"
	deleteTagLocaleMapping        = "UPDATE tag_locale_mapping SET publish = 0, updated_at = ? where id = ?"
)

func NewTagLocaleMappingRepository(db *sql.DB) *TagLocaleMappingRepo {
	return &TagLocaleMappingRepo{db}
}

func (t *TagLocaleMappingRepo) CreateTagLocaleMapping(tx *sql.Tx, tagLocaleMapping *domain.TagLocaleMapping) (err error) {
	stmt, err := tx.Prepare(insertTagLocaleMapping)
	if err != nil {
		logger.Client.Error("createTagLocaleMappingError", logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "createTagLocaleMappingError")
	}
	_, err = stmt.Exec(tagLocaleMapping.TagID, tagLocaleMapping.Locale, *tagLocaleMapping.CountryId, tagLocaleMapping.Name, tagLocaleMapping.Publish, tagLocaleMapping.TagType, tagLocaleMapping.CreatedAt.UnixNano()/1000000, tagLocaleMapping.UpdatedAt.UnixNano()/1000000)
	if err != nil {
		logger.Client.Error("createTagLocaleMappingError", logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "createTagLocaleMappingError")
	}
	return
}

func (t *TagLocaleMappingRepo) FetchTagLocaleMappings(id *string) (tagLocaleMappings []*domain.TagLocaleMapping, err error) {
	rows, err := t.db.Query(selectTagLocaleMapping, *id)
	if err != nil {
		logger.Client.Error("fetchTagLocaleMappingError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchTagLocaleMappingError")
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := tagLocaleMappingRowMapper(rows)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchTagLocaleMappingError")
	}
	return tagsList, nil
}

func (t *TagLocaleMappingRepo) FetchTagLocaleMappingByLocale(tagId *string, countryId *string, locale *string) (tagLocaleMappings *domain.TagLocaleMapping, err error) {
	rows, err := t.db.Query(fetchTagLocaleMappingByLocale, *tagId, *countryId, *locale)
	if err != nil {
		logger.Client.Error("fetchTagLocaleMappingByLocaleError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchTagLocaleMappingByLocaleError")
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := tagLocaleMappingRowMapper(rows)
	if err != nil {
		return
	}
	if len(tagsList) == 1 {
		return tagsList[0], nil
	} else if len(tagsList) > 1 {
		return nil, noonerror.New(noonerror.ErrInternalServer, "tagLocaleMappingDBReadError")
	}
	return
}

func (t *TagLocaleMappingRepo) DeleteTagLocaleMapping(tx *sql.Tx, id *string) (err error) {
	_, err = tx.Query(deleteTagLocaleMapping, time.Now().UnixNano()/1000000, *id)
	if err != nil {
		logger.Client.Error("deleteTagsError", logger.GetErrorStack())
		return
	}
	return
}

func (t *TagLocaleMappingRepo) FetchTagLocalesByTagIds(ids []*string, locale *string, countryId *string) (tagLocaleMappings []*domain.TagLocaleMapping, err error) {
	if len(ids) == 0 {
		return
	}
	args := make([]interface{}, len(ids)+2)
	args[0] = locale
	args[1] = countryId
	for i, id := range ids {
		args[i+2] = id
	}
	stmt := `SELECT * FROM tag_locale_mapping WHERE locale = ? and country_id = ? and tag_id in (?` + strings.Repeat(",?", len(args)-3) + `) and publish = 1`
	rows, err := t.db.Query(stmt, args...)
	if err != nil {
		logger.Client.Error("fetchParentTagMappingsError", logger.GetErrorStack())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := tagLocaleMappingRowMapper(rows)
	if err != nil {
		return
	}
	return tagsList, nil
}

func tagLocaleMappingRowMapper(rows *sql.Rows) (tagLocaleMappings []*domain.TagLocaleMapping, err error) {
	columns, err := rows.Columns()
	if err != nil {
		return
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		tag := &domain.TagLocaleMapping{}
		err = rows.Scan(scanArgs...)
		if err != nil {
			return
		}
		for i, col := range values {
			switch columns[i] {
			case "id":
				tag.ID = converter.ConvertToStringPtr(string(col))
			case "tag_id":
				tag.TagID = converter.ConvertToStringPtr(string(col))
			case "tag_type":
				tag.TagType = converter.ConvertToStringPtr(string(col))
			case "locale":
				tag.Locale = converter.ConvertToStringPtr(string(col))
			case "country_id":
				tag.CountryId = converter.ConvertToStringPtr(string(col))
			case "name":
				tag.Name = converter.ConvertToStringPtr(string(col))
			case "publish":
				tag.Publish, err = strconv.ParseBool(string(col))
			case "created_at":
				var timeMilli int64
				timeMilli, err = strconv.ParseInt(string(col), 10, 64)
				tag.CreatedAt = time.Unix(0, timeMilli*int64(time.Millisecond)).UTC()
			case "updated_at":
				var timeMilli int64
				timeMilli, err = strconv.ParseInt(string(col), 10, 64)
				tag.UpdatedAt = time.Unix(0, timeMilli*int64(time.Millisecond)).UTC()
			default:
				return nil, noonerror.New(noonerror.ErrInternalServer, "invalid column in tags table")
			}
			if err != nil {
				return nil, err
			}
		}
		tagLocaleMappings = append(tagLocaleMappings, tag)
	}
	return tagLocaleMappings, nil
}
