package domain

import (
	"time"
)

type GradeProduct struct {
	ID        *string   `json:"id"`
	FolderId  *string   `json:"folder_id"`
	ProductId *string   `json:"product_id"`
	Grade     *string   `json:"grade"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type GradeProductRepository interface {
	FetchGradesFromProductId(*string) ([]*GradeProduct, error)
}
