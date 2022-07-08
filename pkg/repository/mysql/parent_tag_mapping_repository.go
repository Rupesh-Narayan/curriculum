package repository

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/converter"
	noonerror "bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
)

type ParentTagMappingRepo struct {
	db *sql.DB
}

var (
	fetchParentTagMapping                    = "SELECT * FROM parent_tag_mapping WHERE tag_type = ? and parent_tag_id = ? and hidden = 0 and publish = 1"
	selectParentTagMapping                   = "SELECT * FROM parent_tag_mapping WHERE tag_id = ? and publish = 1"
	selectParentTagMappingByParentTagIds     = "SELECT * FROM parent_tag_mapping WHERE parent_tag_id = ? and tag_type = ? and publish = 1"
	selectParentTagMappingByParentTagIdTagId = "SELECT * FROM parent_tag_mapping WHERE tag_id = ? and parent_tag_id = ? and publish = 1"
	filterParentTagMapping                   = "select tag_id, parent_tag_id, parent_tag_type, hidden, publish from parent_tag_mapping where tag_id in (select tag_id from parent_tag_mapping where tag_type = ? and parent_tag_id = ? and publish = 1)"
	insertParentTagMapping                   = "INSERT INTO parent_tag_mapping(tag_id, tag_type, parent_tag_type, parent_tag_id, `order`, hidden, publish, created_at, updated_at) values(?,?,?,?,?,?,?,?,?)"
	toggleHideParentTagMapping               = "UPDATE parent_tag_mapping SET hidden = ?, updated_at = ? where id = ?"
	updateTagOrderParentTagMapping           = "UPDATE parent_tag_mapping SET `order` = ?, updated_at = ? where id = ?"
	deleteParentTagMapping                   = "UPDATE parent_tag_mapping SET publish = 0, updated_at = ? where id = ?"
)

func NewParentTagMappingRepository(db *sql.DB) *ParentTagMappingRepo {
	return &ParentTagMappingRepo{db}
}

func (t *ParentTagMappingRepo) CreateParentTagMapping(tx *sql.Tx, parentTagMapping *domain.ParentTagMapping) (err error) {
	txPresent := true
	if tx == nil {
		txPresent = false
		ctx := context.Background()
		tx, err = t.db.BeginTx(ctx, nil)
		if err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "createParentTagMappingContextCreationError")
		}
	}
	stmt, err := tx.Prepare(insertParentTagMapping)
	if err != nil {
		logger.Client.Error("createParentTagMappingError", logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "createParentTagMappingError")
	}
	_, err = stmt.Exec(parentTagMapping.TagID, parentTagMapping.TagType, parentTagMapping.ParentTagType, *parentTagMapping.ParentTagID, parentTagMapping.Order, parentTagMapping.Hidden, parentTagMapping.Publish, parentTagMapping.CreatedAt.UnixNano()/1000000, parentTagMapping.UpdatedAt.UnixNano()/1000000)
	if err != nil {
		logger.Client.Error("createParentTagMappingError", logger.GetErrorStack())
		return noonerror.New(noonerror.ErrInternalServer, "createParentTagMappingError")
	}
	if !txPresent {
		if err = tx.Commit(); err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "createParentTagMappingCommitError")
		}
	}
	return
}

func (t *ParentTagMappingRepo) FetchFilteredParentTagMappings(tagType *string, id *string) (parentTagMappings []*domain.ParentTagMapping, err error) {
	rows, err := t.db.Query(filterParentTagMapping, *tagType, *id)
	if err != nil {
		logger.Client.Error("fetchFilteredParentTagMappingsError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "parentTagMappingDBReadError")
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := parentTagMappingRowMapper(rows)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "parentTagMappingMapperError")
	}
	return tagsList, nil
}

func (t *ParentTagMappingRepo) FetchByInParentTagMappings(ids []*string) (parentTagMappings []*domain.ParentTagMapping, err error) {
	if len(ids) == 0 {
		return
	}
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	stmt := `SELECT * FROM parent_tag_mapping WHERE tag_id in (?` + strings.Repeat(",?", len(args)-1) + `) and publish = 1`
	rows, err := t.db.Query(stmt, args...)
	if err != nil {
		logger.Client.Error("fetchParentTagMappingsError", logger.GetErrorStack())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := parentTagMappingRowMapper(rows)
	if err != nil {
		return
	}
	return tagsList, nil
}

func (t *ParentTagMappingRepo) FetchParentTagMappings(id *string) (parentTagMappings []*domain.ParentTagMapping, err error) {
	rows, err := t.db.Query(selectParentTagMapping, *id)
	if err != nil {
		logger.Client.Error("fetchParentTagMappingsError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "parentTagMappingDBReadError")
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := parentTagMappingRowMapper(rows)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "parentTagMappingMapperError")
	}
	return tagsList, nil
}

func (t *ParentTagMappingRepo) FetchParentTagMappingByParentTagIdTagId(tagId *string, parentTagId *string) (parentTagMappings *domain.ParentTagMapping, err error) {
	rows, err := t.db.Query(selectParentTagMappingByParentTagIdTagId, *tagId, *parentTagId)
	if err != nil {
		logger.Client.Error("fetchParentTagMappingsError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "parentTagMappingDBReadError")
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := parentTagMappingRowMapper(rows)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "parentTagMappingMapperError")
	}
	if len(tagsList) == 0 {
		return
	}
	if len(tagsList) > 1 {
		return nil, noonerror.New(noonerror.ErrInternalServer, "parentTagMappingCountError")
	}
	return tagsList[0], nil
}

func (t *ParentTagMappingRepo) FetchByInParentTagMappingsByParentTagIdTagIds(ids []*string, parentTagId *string) (parentTagMappings []*domain.ParentTagMapping, err error) {
	if len(ids) == 0 {
		return
	}
	args := make([]interface{}, len(ids)+1)
	for i, id := range ids {
		args[i] = *id
	}
	args[len(ids)] = *parentTagId
	stmt := `SELECT * FROM parent_tag_mapping WHERE tag_id in (?` + strings.Repeat(",?", len(args)-2) + `) and parent_tag_id = ? and publish = 1`
	rows, err := t.db.Query(stmt, args...)
	if err != nil {
		logger.Client.Error("fetchParentTagMappingsError", logger.GetErrorStack())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := parentTagMappingRowMapper(rows)
	if err != nil {
		return
	}
	return tagsList, nil
}

func (t *ParentTagMappingRepo) FetchParentTagMappingsByParentTagIds(parentTagId *string, tagType *string) (parentTagMappings []*domain.ParentTagMapping, err error) {
	rows, err := t.db.Query(selectParentTagMappingByParentTagIds, *parentTagId, *tagType)
	if err != nil {
		logger.Client.Error("fetchParentTagMappingsByParentTagIdsError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "parentTagMappingsDBReadError")
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := parentTagMappingRowMapper(rows)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "parentTagMappingsMapperError")
	}
	return tagsList, nil
}

func (t *ParentTagMappingRepo) ToggleHideParentTagMapping(tx *sql.Tx, hidden bool, id *string) (err error) {
	txPresent := true
	if tx == nil {
		txPresent = false
		ctx := context.Background()
		tx, err = t.db.BeginTx(ctx, nil)
		if err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "toggleHideParentTagMappingContextCreationError")
		}
	}
	_, err = tx.Query(toggleHideParentTagMapping, hidden, time.Now().UnixNano()/1000000, *id)
	if err != nil {
		logger.Client.Error("toggleHideTagsError", logger.GetErrorStack())
		return
	}
	if !txPresent {
		if err = tx.Commit(); err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "toggleHideParentTagMappingCommitError")
		}
	}
	return
}

func (t *ParentTagMappingRepo) DeleteParentTagMapping(tx *sql.Tx, id *string) (err error) {
	txPresent := true
	if tx == nil {
		txPresent = false
		ctx := context.Background()
		tx, err = t.db.BeginTx(ctx, nil)
		if err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "deleteParentTagMappingContextCreationError")
		}
	}
	_, err = tx.Query(deleteParentTagMapping, time.Now().UnixNano()/1000000, *id)
	if err != nil {
		logger.Client.Error("deleteTagsError", logger.GetErrorStack())
		return
	}
	if !txPresent {
		if err = tx.Commit(); err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "deleteParentTagMappingCommitError")
		}
	}
	return
}

func (t *ParentTagMappingRepo) IsCollegePresent(tagType *string, tagId *string) (hasCollege bool, err error) {
	rows, err := t.db.Query(fetchParentTagMapping, *tagType, *tagId)
	if err != nil {
		logger.Client.Error("fetchParentTagMapping", logger.GetErrorStack())
		return false, nil
	}
	defer func() {
		_ = rows.Close()
	}()
	hasResults := false
	for rows.Next() {
		hasResults = true
	}
	return hasResults, nil
}

func (t *ParentTagMappingRepo) UpdateTagOrder(tx *sql.Tx, order *int, id *string) (err error) {
	txPresent := true
	if tx == nil {
		txPresent = false
		ctx := context.Background()
		tx, err = t.db.BeginTx(ctx, nil)
		if err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "updateTagOrderContextCreationError")
		}
	}
	_, err = tx.Query(updateTagOrderParentTagMapping, order, time.Now().UnixNano()/1000000, *id)
	if err != nil {
		logger.Client.Error("updateTagOrderError", logger.GetErrorStack())
		return
	}
	if !txPresent {
		if err = tx.Commit(); err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "updateTagOrderCommitError")
		}
	}
	return
}

func parentTagMappingRowMapper(rows *sql.Rows) (parentTagMappings []*domain.ParentTagMapping, err error) {
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
		tag := &domain.ParentTagMapping{}
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
			case "parent_tag_type":
				tag.ParentTagType = converter.ConvertToStringPtr(string(col))
			case "parent_tag_id":
				tag.ParentTagID = converter.ConvertToStringPtr(string(col))
			case "order":
				order, _ := strconv.Atoi(string(col))
				tag.Order = &order
			case "hidden":
				tag.Hidden, err = strconv.ParseBool(string(col))
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
		parentTagMappings = append(parentTagMappings, tag)
	}
	return parentTagMappings, nil
}
