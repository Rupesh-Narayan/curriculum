package repository

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/converter"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type TagsRepo struct {
	db *sql.DB
}

var (
	selectTags                  = "SELECT * FROM tags WHERE id = ?"
	insertTags                  = "INSERT INTO tags(type, name, curriculum_type, creator_id, creator_type, access, tag_group, locale_available, country_id, publish, attributes, created_at, updated_at) values(?,?,?,?,?,?,?,?,?,?,?,?,?)"
	filterTags                  = "select id, type, name, attributes from tags where curriculum_type = ? and type = ? and publish = 1"
	filterTagsByTagGroup        = "select id, type, name, attributes from tags where tag_group = ? and publish = 1"
	filterTagsByTagGroupAndType = "select id, type, name, attributes from tags where tag_group = ? and type = ? and publish = 1"
	deleteTags                  = "UPDATE tags SET publish = 0, updated_at = ? where id = ?"
	updateLocale                = "UPDATE tags SET locale_available = ?, updated_at = ? where id = ?"
	filterTagsPaginated         = "select id, type, name, attributes, publish from tags where curriculum_type = ? and type = ? and publish = 1 limit ? offset ?"
	filterTagsPaginatedForAdmin = "select id, type, name, attributes, publish from tags where curriculum_type = ? and type = ? limit ? offset ?"
)

func NewTagsRepository(db *sql.DB) *TagsRepo {
	return &TagsRepo{db}
}

func (t *TagsRepo) FetchTags(id *string) (tags *domain.Tags, err error) {
	rows, err := t.db.Query(selectTags, *id)
	if err != nil {
		logger.Client.Error("fetchTagsError", logger.GetErrorStack())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := tagsRowMapper(rows)
	if err != nil {
		return
	}
	if len(tagsList) == 1 {
		return tagsList[0], nil
	} else if len(tagsList) > 1 {
		return nil, noonerror.New(noonerror.ErrInternalServer, "tagDBReadError")
	}
	return
}

func (t *TagsRepo) FetchByInTags(ids []*string) (tags []*domain.Tags, err error) {
	if len(ids) == 0 {
		return
	}
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	stmt := `SELECT * FROM tags WHERE id in (?` + strings.Repeat(",?", len(args)-1) + `)`
	rows, err := t.db.Query(stmt, args...)
	if err != nil {
		logger.Client.Error("fetchTagsError", logger.GetErrorStack())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := tagsRowMapper(rows)
	if err != nil {
		return
	}
	return tagsList, nil
}

func (t *TagsRepo) FetchFilteredTags(curriculumType *string, tagType *string) (tags []*domain.Tags, err error) {
	rows, err := t.db.Query(filterTags, *curriculumType, *tagType)
	if err != nil {
		logger.Client.Error("fetchFilteredTagsError", logger.GetErrorStack())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := tagsRowMapper(rows)
	if err != nil {
		return
	}
	return tagsList, nil
}

func (t *TagsRepo) FetchFilteredTagsPaginated(curriculumType *string, tagType *string, start *int, limit *int) (tags []*domain.Tags, err error) {
	rows, err := t.db.Query(filterTagsPaginated, *curriculumType, *tagType, *limit, *start)
	if err != nil {
		logger.Client.Error("fetchFilteredTagsError", logger.GetErrorStack())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := tagsRowMapper(rows)
	if err != nil {
		return
	}
	return tagsList, nil
}

func (t *TagsRepo) FetchFilteredTagsPaginatedForAdmin(curriculumType *string, tagType *string, start *int, limit *int) (tags []*domain.Tags, err error) {
	rows, err := t.db.Query(filterTagsPaginatedForAdmin, *curriculumType, *tagType, *limit, *start)
	if err != nil {
		logger.Client.Error("fetchFilteredTagsError", logger.GetErrorStack())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := tagsRowMapper(rows)
	if err != nil {
		return
	}
	return tagsList, nil
}

func (t *TagsRepo) FetchByTagGroup(tagGroup *string, tagType *string) (tags []*domain.Tags, err error) {
	var rows *sql.Rows
	if tagType != nil {
		rows, err = t.db.Query(filterTagsByTagGroupAndType, *tagGroup, *tagType)
	} else {
		rows, err = t.db.Query(filterTagsByTagGroup, *tagGroup)
	}
	if err != nil {
		logger.Client.Error("fetchByTagGroupError", logger.GetErrorStack())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := tagsRowMapper(rows)
	if err != nil {
		return
	}
	return tagsList, nil
}

func (t *TagsRepo) CreateTags(tx *sql.Tx, tags *domain.Tags) (id *string, err error) {
	var attributesStringPtr *string
	if tags.Attributes != nil {
		attributes, _ := json.Marshal(tags.Attributes)
		attributesString := string(attributes)
		attributesStringPtr = &attributesString
	}
	stmt, err := tx.Prepare(insertTags)
	if err != nil {
		logger.Client.Error("createTagError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "createTagError")
	}
	res, err := stmt.Exec(tags.Type, tags.Name, tags.CurriculumType, tags.CreatorId, tags.CreatorType, tags.Access, tags.TagGroup, tags.LocaleAvailable, tags.CountryId, tags.Publish, attributesStringPtr, tags.CreatedAt.UnixNano()/1000000, tags.UpdatedAt.UnixNano()/1000000)
	if err != nil {
		logger.Client.Error("createTagError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "createTagError")
	}
	insertId, err := res.LastInsertId()
	if err != nil {
		logger.Client.Error("createTagError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "createTagError")
	}
	insertStringId := strconv.FormatInt(insertId, 10)
	return &insertStringId, nil
}

func (t *TagsRepo) UpdateTag(updateTag *domain.UpdateTag) (err error) {
	queryString := "UPDATE tags SET "
	var updateFields []interface{}
	updated := false
	if updateTag.Name != nil {
		queryString += "name = ?, "
		updateFields = append(updateFields, *updateTag.Name)
		updated = true
	}
	if updateTag.Hidden != nil {
		queryString += "publish = ?, "
		updateFields = append(updateFields, !*updateTag.Hidden)
		updated = true
	}
	if updateTag.Attributes != nil {
		attributes, _ := json.Marshal(updateTag.Attributes)
		attributesString := string(attributes)
		queryString += "attributes = ?, "
		updateFields = append(updateFields, attributesString)
		updated = true
	}
	queryString += "updated_at = ? where id = ?"
	updateFields = append(updateFields, time.Now().UnixNano()/1000000, *updateTag.ID)
	if updated {
		rows, err := t.db.Query(queryString, updateFields...)
		if err != nil {
			logger.Client.Error("updateTagError", logger.GetErrorStack())
			return noonerror.New(noonerror.ErrInternalServer, "updateTagError")
		}
		_ = rows.Close()
	}
	return
}

func (t *TagsRepo) DeleteTags(tx *sql.Tx, id *string) (err error) {
	txPresent := true
	if tx == nil {
		txPresent = false
		ctx := context.Background()
		tx, err = t.db.BeginTx(ctx, nil)
		if err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "deleteTagsContextCreationError")
		}
	}
	_, err = tx.Query(deleteTags, time.Now().UnixNano()/1000000, *id)
	if err != nil {
		logger.Client.Error("deleteTagsError", logger.GetErrorStack())
		return
	}
	if !txPresent {
		if err = tx.Commit(); err != nil {
			return noonerror.New(noonerror.ErrInternalServer, "deleteTagsCommitError")
		}
	}
	return
}

func (t *TagsRepo) UpdateLocale(tx *sql.Tx, localeAvailable bool, id *string) (err error) {
	_, err = tx.Query(updateLocale, localeAvailable, time.Now().UnixNano()/1000000, *id)
	if err != nil {
		logger.Client.Error("updateLocaleError", logger.GetErrorStack())
		return
	}
	return
}

func (t *TagsRepo) ToggleTags(publish bool, ids []*string) (err error) {
	if len(ids) == 0 {
		return
	}
	args := make([]interface{}, len(ids)+2)
	args[0] = publish
	args[1] = time.Now().UnixNano() / 1000000
	for i, id := range ids {
		args[i+2] = id
	}
	stmt := `UPDATE tags SET publish = ?, updated_at = ? WHERE id in (?` + strings.Repeat(",?", len(args)-1) + `)`
	rows, err := t.db.Query(stmt, args...)
	if err != nil {
		logger.Client.Error("toggleTagsError", logger.GetErrorStack())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	return
}

func tagsRowMapper(rows *sql.Rows) (tags []*domain.Tags, err error) {
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
		tag := &domain.Tags{}
		err = rows.Scan(scanArgs...)
		if err != nil {
			return
		}
		for i, col := range values {
			switch columns[i] {
			case "id":
				tag.ID = converter.ConvertToStringPtr(string(col))
			case "type":
				tag.Type = converter.ConvertToStringPtr(string(col))
			case "name":
				tag.Name = converter.ConvertToStringPtr(string(col))
			case "curriculum_type":
				tag.CurriculumType = string(col)
			case "creator_id":
				creatorId, _ := strconv.ParseInt(string(col), 10, 64)
				tag.CreatorId = converter.ConvertToInt64Ptr(creatorId)
			case "creator_type":
				tag.CreatorType = string(col)
			case "access":
				tag.Access = string(col)
			case "tag_group":
				tag.TagGroup = string(col)
			case "locale_available":
				tag.LocaleAvailable, err = strconv.ParseBool(string(col))
			case "country_id":
				tag.CountryId = string(col)
			case "publish":
				tag.Publish, err = strconv.ParseBool(string(col))
			case "attributes":
				attributes := make(map[string]interface{})
				err = json.Unmarshal(col, &attributes)
				if len(attributes) == 0 {
					attributes = nil
				}
				tag.Attributes = attributes
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
		}
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}
