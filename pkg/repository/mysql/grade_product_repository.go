package repository

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/converter"
	noonerror "bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"database/sql"
	"time"
)

type GradeProductRepo struct {
	db *sql.DB
}

var (
	selectGradeFromProduct = "SELECT * FROM grade_product WHERE product_id = ?"
)

func NewGradeProductRepository(db *sql.DB) *GradeProductRepo {
	return &GradeProductRepo{db}
}

func (t *GradeProductRepo) FetchGradesFromProductId(productId *string) (gradeProducts []*domain.GradeProduct, err error) {
	rows, err := t.db.Query(selectGradeFromProduct, *productId)
	if err != nil {
		logger.Client.Error("fetchGradesFromProductIdError", logger.GetErrorStack())
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchGradesFromProductIdError")
	}
	defer func() {
		_ = rows.Close()
	}()
	tagsList, err := gradeProductRowMapper(rows)
	if err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "fetchGradesFromProductIdError")
	}
	return tagsList, nil
}

func gradeProductRowMapper(rows *sql.Rows) (gradeProducts []*domain.GradeProduct, err error) {
	columns, err := rows.Columns()
	if err != nil {
		return
	}
	layout := "2006-01-02 15:04:05"
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		tag := &domain.GradeProduct{}
		err = rows.Scan(scanArgs...)
		if err != nil {
			return
		}
		for i, col := range values {
			switch columns[i] {
			case "id":
				tag.ID = converter.ConvertToStringPtr(string(col))
			case "folder_id":
				tag.FolderId = converter.ConvertToStringPtr(string(col))
			case "product_id":
				tag.ProductId = converter.ConvertToStringPtr(string(col))
			case "grade":
				tag.Grade = converter.ConvertToStringPtr(string(col))
			case "created_at":
				tag.CreatedAt, err = time.Parse(layout, string(col))
			case "updated_at":
				tag.UpdatedAt, err = time.Parse(layout, string(col))
			default:
				return nil, noonerror.New(noonerror.ErrInternalServer, "invalid column in grade_product table")
			}
			if err != nil {
				return nil, err
			}
		}
		gradeProducts = append(gradeProducts, tag)
	}
	return gradeProducts, nil
}
