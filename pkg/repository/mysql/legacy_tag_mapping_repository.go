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

type LegacyTagMappingRepo struct {
	db *sql.DB
}

var (
	selectTagId    = "SELECT * FROM legacy_tag_mapping WHERE legacy_id_type = ? and legacy_id = ?"
	selectLegacyId = "SELECT * FROM legacy_tag_mapping WHERE tag_id = ?"
)

func NewLegacyTagMappingRepository(db *sql.DB) *LegacyTagMappingRepo {
	return &LegacyTagMappingRepo{db}
}

func (t *LegacyTagMappingRepo) FetchTagIdFromLegacyId(legacyType *string, id *string) (legacyTagMappings []*domain.LegacyTagMapping, err error) {
	rows, err := t.db.Query(selectTagId, *legacyType, *id)
	if err != nil {
		logger.Client.Error("fetchTLegacyTagMappingError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchLegacyTagMappingError")
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := legacyTagMappingRowMapper(rows)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchTLegacyTagMappingError")
	}
	return tagsList, nil
}

func (t *LegacyTagMappingRepo) FetchLegacyIdFromTagId(tagId *string) (legacyTagMappings []*domain.LegacyTagMapping, err error) {
	rows, err := t.db.Query(selectLegacyId, *tagId)
	if err != nil {
		logger.Client.Error("fetchTLegacyTagMappingError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchLegacyTagMappingError")
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := legacyTagMappingRowMapper(rows)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchTLegacyTagMappingError")
	}
	return tagsList, nil
}

func (t *LegacyTagMappingRepo) FetchLegacyIdFromTagIds(tagIds []*string) (legacyTagMappings []*domain.LegacyTagMapping, err error) {
	if len(tagIds) == 0 {
		return
	}
	args := make([]interface{}, len(tagIds))
	for i, id := range tagIds {
		args[i] = id
	}
	stmt := `SELECT * FROM legacy_tag_mapping WHERE tag_id in (?` + strings.Repeat(",?", len(args)-1) + `)`
	rows, err := t.db.Query(stmt, args...)
	if err != nil {
		logger.Client.Error("fetchTLegacyTagMappingError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchLegacyTagMappingError")
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := legacyTagMappingRowMapper(rows)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchTLegacyTagMappingError")
	}
	return tagsList, nil
}

func legacyTagMappingRowMapper(rows *sql.Rows) (legacyTagMappings []*domain.LegacyTagMapping, err error) {
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
		tag := &domain.LegacyTagMapping{}
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
			case "tag_id_type":
				tag.TagIdType = converter.ConvertToStringPtr(string(col))
			case "legacy_id_type":
				tag.LegacyIdType = converter.ConvertToStringPtr(string(col))
			case "legacy_id":
				tag.LegacyId = converter.ConvertToStringPtr(string(col))
			case "created_at":
				var timeMilli int64
				timeMilli, err = strconv.ParseInt(string(col), 10, 64)
				tag.CreatedAt = time.Unix(0, timeMilli*int64(time.Millisecond)).UTC()
			case "updated_at":
				var timeMilli int64
				timeMilli, err = strconv.ParseInt(string(col), 10, 64)
				tag.UpdatedAt = time.Unix(0, timeMilli*int64(time.Millisecond)).UTC()
			default:
				return nil, noonerror.New(noonerror.ErrInternalServer, "invalid column in tag_legacy_mapping table")
			}
			if err != nil {
				return nil, err
			}
		}
		legacyTagMappings = append(legacyTagMappings, tag)
	}
	return legacyTagMappings, nil
}
