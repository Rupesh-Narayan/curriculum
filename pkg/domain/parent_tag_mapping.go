package domain

import (
	"database/sql"
	"time"
)

type ParentTagMapping struct {
	ID            *string   `json:"id"`
	TagID         *string   `json:"tag_id" validate:"required"`
	TagType       *string   `json:"tag_type" validate:"required"`
	ParentTagType *string   `json:"parent_tag_type" validate:"required"`
	ParentTagID   *string   `json:"parent_tag_id"`
	Order         *int      `json:"order"`
	Hidden        bool      `json:"hidden"`
	Publish       bool      `json:"publish"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type ParentTagMappingRepository interface {
	CreateParentTagMapping(*sql.Tx, *ParentTagMapping) error
	FetchParentTagMappings(*string) ([]*ParentTagMapping, error)
	FetchByInParentTagMappings([]*string) ([]*ParentTagMapping, error)
	FetchFilteredParentTagMappings(*string, *string) ([]*ParentTagMapping, error)
	FetchParentTagMappingsByParentTagIds(*string, *string) ([]*ParentTagMapping, error)
	FetchParentTagMappingByParentTagIdTagId(*string, *string) (*ParentTagMapping, error)
	FetchByInParentTagMappingsByParentTagIdTagIds([]*string, *string) ([]*ParentTagMapping, error)
	ToggleHideParentTagMapping(*sql.Tx, bool, *string) error
	UpdateTagOrder(*sql.Tx, *int, *string) error
	DeleteParentTagMapping(*sql.Tx, *string) error
	IsCollegePresent(*string, *string) (bool, error)
}

type ParentTagMappingService interface {
}
