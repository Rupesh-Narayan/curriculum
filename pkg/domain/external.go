package domain

type Elastic interface {
	CreateTag(*CreateTagElastic) error
	GetTags(*GetTagsElastic) ([]*string, *int, error)
	GetTagsSearch(*GetTagsElastic) ([]*string, *int, error)
	UpdateTag(*string, *bool, []*TagName) error
	AddParentTags(*string, []*string) error
	RemoveParentTags(*string, []*string) error
	HideParentTags(*string, []*string) error
}

type GeoIp interface {
	GetGeoIp(*GetGeoIp) (*string, error)
}

type CreateTagElastic struct {
	ID             *string    `json:"id"`
	Type           *string    `json:"type"`
	Name           []*TagName `json:"name"`
	CurriculumType *string    `json:"curriculum_type"`
	CreatorId      *int64     `json:"creator_id"`
	CreatorType    *string    `json:"creator_type"`
	Access         *string    `json:"access"`
	TagGroup       *string    `json:"tag_group"`
	CountryId      string     `json:"country_id"`
	Parents        []*string  `json:"parents"`
	Deleted        bool       `json:"deleted"`
}

type GetTagsElastic struct {
	Text           *string   `json:"text,omitempty"`
	Type           *string   `json:"type"`
	CurriculumType *string   `json:"curriculum_type"`
	CreatorId      *int64    `json:"creator_id"`
	CreatorType    *string   `json:"creator_type"`
	Access         *string   `json:"access"`
	TagGroup       *string   `json:"tag_group"`
	CountryId      string    `json:"country_id"`
	HiddenParents  []*string `json:"hidden_active_parents"`
	Parents        []*string `json:"parents"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
}

type GetGeoIp struct {
	Ip *string `json:"ip_address"`
}

type TagName struct {
	Value  *string `json:"value"`
	Locale *string `json:"locale"`
}
