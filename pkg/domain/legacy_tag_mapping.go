package domain

import (
	"time"
)

type LegacyTagMapping struct {
	ID           *string   `json:"id"`
	TagID        *string   `json:"tag_id" validate:"required"`
	TagIdType    *string   `json:"tag_id_type"`
	LegacyIdType *string   `json:"legacy_id_type"`
	LegacyId     *string   `json:"legacy_id"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type LegacyTagMappingRepository interface {
	FetchTagIdFromLegacyId(*string, *string) ([]*LegacyTagMapping, error)
	FetchLegacyIdFromTagId(*string) ([]*LegacyTagMapping, error)
	FetchLegacyIdFromTagIds([]*string) ([]*LegacyTagMapping, error)
}
