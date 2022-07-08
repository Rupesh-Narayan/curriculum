package domain

import (
	"database/sql"
	"time"
)

type TagLocaleMapping struct {
	ID        *string   `json:"id"`
	TagID     *string   `json:"tag_id" validate:"required"`
	CountryId *string   `json:"country_id"`
	Locale    *string   `json:"locale"`
	Name      *string   `json:"name"`
	Publish   bool      `json:"publish"`
	TagType   *string   `json:"tag_type"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type TagLocaleMappingRepository interface {
	CreateTagLocaleMapping(*sql.Tx, *TagLocaleMapping) error
	FetchTagLocaleMappings(*string) ([]*TagLocaleMapping, error)
	FetchTagLocaleMappingByLocale(*string, *string, *string) (*TagLocaleMapping, error)
	DeleteTagLocaleMapping(*sql.Tx, *string) error
	FetchTagLocalesByTagIds(ids []*string, locale *string, countryId *string) ([]*TagLocaleMapping, error)
}

type TagLocaleTagMappingService interface {
}
