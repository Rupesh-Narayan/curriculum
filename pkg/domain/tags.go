package domain

import (
	"database/sql"
	"time"
)

type Tags struct {
	ID              *string                `json:"id"`
	Type            *string                `json:"type" validate:"required"`
	Name            *string                `json:"name" validate:"required"`
	LocaleName      *string                `json:"locale_name"`
	CurriculumType  string                 `json:"curriculum_type"`
	CreatorId       *int64                 `json:"creator_id"`
	CreatorType     string                 `json:"creator_type"`
	Access          string                 `json:"access"`
	TagGroup        string                 `json:"tag_group"`
	LocaleAvailable bool                   `json:"locale_available"`
	CountryId       string                 `json:"country_id"`
	Publish         bool                   `json:"publish"`
	Attributes      map[string]interface{} `json:"attributes"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedAt       time.Time              `json:"created_at"`
}

type CreateTags struct {
	Type           *string                `json:"type"`
	Name           *string                `json:"name"`
	CurriculumType *string                `json:"curriculum_type"`
	CreatorId      *int64                 `json:"creator_id"`
	CreatorType    *string                `json:"creator_type"`
	TagGroup       *string                `json:"tag_group"`
	Access         string                 `json:"access"`
	CountryId      string                 `json:"country_id"`
	Attributes     map[string]interface{} `json:"attributes"`
	Hierarchy      []*string              `json:"hierarchy"`
	Identifier     []*string              `json:"identifier"`
}

type UpdateTags struct {
	ID             *string   `json:"id"`
	CurriculumType *string   `json:"curriculum_type"`
	TagGroup       *string   `json:"tag_group"`
	Hierarchy      []*string `json:"hierarchy"`
	Identifier     []*string `json:"attributes"`
}

type UpdateMultipleTags struct {
	IDs            []*string `json:"ids"`
	Type           *string   `json:"type"`
	CurriculumType *string   `json:"curriculum_type"`
	TagGroup       *string   `json:"tag_group"`
	Hierarchy      []*string `json:"hierarchy"`
}

type UpdateTag struct {
	ID         *string                `json:"id" validate:"required"`
	Type       *string                `json:"type"`
	Attributes map[string]interface{} `json:"attributes"`
	Name       *string                `json:"name"`
	Hidden     *bool                  `json:"hidden"`
}

type UpdateTagOrder struct {
	Type           *string   `json:"type" validate:"required"`
	CurriculumType *string   `json:"curriculum_type"`
	TagGroup       *string   `json:"tag_group"`
	Hierarchy      []*string `json:"hierarchy"`
	Orders         []*Order  `json:"orders"`
}

type Order struct {
	ID    *string `json:"id"`
	SqlId *string `json:"sql_id"`
	Order *int    `json:"order"`
}

type RemoveHierarchy struct {
	ID             *string   `json:"id"`
	CurriculumType *string   `json:"curriculum_type"`
	TagGroup       *string   `json:"tag_group"`
	Hierarchy      []*string `json:"hierarchy"`
	Identifier     []*string `json:"attributes"`
}

type GetTags struct {
	Text           *string   `json:"text"`
	Locale         *string   `json:"locale"`
	Type           *string   `json:"type"`
	CurriculumType *string   `json:"curriculum_type"`
	TagGroup       *string   `json:"tag_group"`
	CountryId      *string   `json:"country_id"`
	MultiGrade     *string   `json:"multi_grade"`
	Hierarchy      *string   `json:"hierarchy"`
	Identifier     []*string `json:"identifier"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
}

type GetCountries struct {
	Text           *string   `json:"text"`
	Type           *string   `json:"type"`
	CurriculumType *string   `json:"curriculum_type"`
	TagGroup       *string   `json:"tag_group"`
	Hierarchy      *string   `json:"hierarchy"`
	Identifier     []*string `json:"identifier"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
}

type GetAdminTags struct {
	Text           *string   `json:"text"`
	Type           *string   `json:"type"`
	CurriculumType *string   `json:"curriculum_type"`
	TagGroup       *string   `json:"tag_group"`
	CountryId      *string   `json:"country_id"`
	Locale         *string   `json:"locale"`
	Hierarchy      []*string `json:"hierarchy"`
	CreatorId      *int64    `json:"creator_id"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
}

type GetTeacherTags struct {
	Text           *string   `json:"text"`
	Type           *string   `json:"type"`
	CurriculumType *string   `json:"curriculum_type"`
	TagGroup       *string   `json:"tag_group"`
	CountryId      *string   `json:"country_id"`
	Locale         *string   `json:"locale"`
	Hierarchy      []*string `json:"hierarchy"`
	CreatorId      *int64    `json:"creator_id"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
}

type GetCountriesNew struct {
	IpAddress string  `json:ip_address`
	ISOCode   string  `json:iso_code`
	CountryId *string `json:"country_id"`
	Locale    *string `json:"locale"`
	Start     int     `json:"start"`
	Limit     int     `json:"limit"`
}

type GetGradesRequest struct {
	CountryId int `json:"country_id"`
	Start     int `json:"start"`
	Limit     int `json:"limit"`
}

type GetRpcTags struct {
	Text           *string   `json:"text"`
	Type           *string   `json:"type"`
	CurriculumType *string   `json:"curriculum_type"`
	TagGroup       *string   `json:"tag_group"`
	CountryId      *string   `json:"country_id"`
	BoardId        *string    `json:"board_id"`
	Locale         *string   `json:"locale"`
	Hierarchy      []*string `json:"hierarchy"`
	CreatorId      *int64    `json:"creator_id"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
}

type GetTagsByIds struct {
	TagIds    []*string `json:"tag_ids"`
	Locale    *string   `json:"locale"`
	CountryId *string   `json:"country_id"`
}

type CreateMultipleTags struct {
	Type           *string                `json:"type"`
	CreatorId      *int64                 `json:"creator_id"`
	CreatorType    *string                `json:"creator_type"`
	CurriculumType *string                `json:"curriculum_type"`
	TagGroup       *string                `json:"tag_group"`
	Attributes     map[string]interface{} `json:"attributes"`
	Hierarchy      []*string              `json:"hierarchy"`
	Identifier     []*string              `json:"identifier"`
	Tags           []*CreateTag           `json:"tags"`
}

type CreateTag struct {
	Name      *string `json:"name" validate:"required"`
	CountryId string  `json:"country_id"`
}

type TagLocale struct {
	ID     *string   `json:"id" validate:"required"`
	Name   *string   `json:"name" validate:"required"`
	Locale []*Locale `json:"locales" validate:"required"`
}

type Locale struct {
	CountryId *string `json:"country_id" validate:"required"`
	Locale    *string `json:"locale" validate:"required"`
}

type ValidateHierarchy struct {
	Hierarchies    [][]*string `json:"hierarchies" validate:"required"`
	CurriculumType *string     `json:"curriculum_type"`
}

type GetSuggestedTags struct {
	TagIds         []*string `json:"tag_ids" validate:"required"`
	CurriculumType *string   `json:"curriculum_type"`
	Locale         *string   `json:"locale" validate:"required"`
	CountryId      *string   `json:"country_id" validate:"required"`
}

type SuggestedTags struct {
	ID             *string          `json:"id"`
	Type           *string          `json:"type"`
	CurriculumType *string          `json:"curriculum_type,omitempty"`
	Name           *string          `json:"name"`
	Topics         []*SuggestedTags `json:"topics,omitempty"`
}

type GetTagsResponse struct {
	Tags []*TagResponse
	Meta *MetaResponse `json:"meta,omitempty"`
}

type GetCountriesResponse struct {
	Tags []*TagResponse
	Meta *MetaResponse `json:"meta,omitempty"`
}

type GetCountriesNewResponse struct {
	Tags []*CountriesAttributesResponse
	Meta *CountriesNewMetaResponse `json:"meta,omitempty"`
}

type GetGradesResponse struct {
	Grades []*GradesAttributesResponse
	Meta   *GradesMetaResponse `json:"meta,omitempty"`
}
type GetBoardsResponse struct {
	Boards []*BoardsAttributesResponse
	Meta   *BoardsMetaResponse `json:"meta,omitempty"`
}

type GetDegreesResponse struct {
	Degrees []*DegreesAttributesResponse
	Meta    *DegreesMetaResponse `json:"meta,omitempty"`
}

type GetMajorsResponse struct {
	Degrees []*MajorsAttributesResponse
	Meta    *MajorsMetaResponse `json:"meta,omitempty"`
}

type GetTagsResponseForProduct struct {
	Tags []*TagResponseForProduct
	Meta *MetaResponse `json:"meta,omitempty"`
}

type CreateTagResponse struct {
	ID *string `json:"id"`
}

type MetaResponse struct {
	IsIdentifier *bool `json:"is_identifier"`
	IsOrdered    *bool `json:"is_ordered"`
	Next         *int  `json:"next,omitempty"`
}

type CountriesNewMetaResponse struct {
	Next            *int                         `json:"next,omitempty"`
	SelectedCountry *CountriesAttributesResponse `json:"selected_country,omitempty"`

}

type GradesMetaResponse struct {
	HasUniversity *bool `json:"has_university"`
}
type BoardsMetaResponse struct {
	HasUniversity *bool `json:"has_university"`
}

type DegreesMetaResponse struct {
	Next *int `json:"next,omitempty"`
}

type MajorsMetaResponse struct {
	Next *int `json:"next,omitempty"`
}

type TagResponse struct {
	ID             *string                `json:"id"`
	Type           *string                `json:"type"`
	CurriculumType *string                `json:"curriculum_type,omitempty"`
	Grade          *int                   `json:"grade,omitempty"`
	Name           *string                `json:"name"`
	LocaleName     *string                `json:"locale_name"`
	Hidden         bool                   `json:"hidden"`
	Root           *string                `json:"root"`
	Attributes     map[string]interface{} `json:"attributes"`
	Identifiers    []*IdentifierResponse  `json:"identifiers,omitempty"`
	Locale         []*LocaleResponse      `json:"locales,omitempty"`
}

type TagResponseForProduct struct {
	ID             *string               `json:"id"`
	Type           *string               `json:"type"`
	CurriculumType *string               `json:"curriculum_type,omitempty"`
	Name           *string               `json:"name"`
	LocaleName     *string               `json:"locale_name"`
	Hidden         bool                  `json:"hidden"`
	Root           *string               `json:"root"`
	BackgroundPic  *string               `json:"background_pic"`
	Color          *string               `json:"color"`
	Pic            *string               `json:"pic"`
	NegativePic    *string               `json:"negative_pic"`
	Identifiers    []*IdentifierResponse `json:"identifiers,omitempty"`
	Locale         []*LocaleResponse     `json:"locales,omitempty"`
}

type CountriesNewResponse struct {
	ID             *string                      `json:"id"`
	Type           *string                      `json:"type"`
	CurriculumType *string                      `json:"curriculum_type,omitempty"`
	Name           *string                      `json:"name"`
	Hidden         bool                         `json:"hidden"`
	Root           *string                      `json:"root"`
	Attributes     *CountriesAttributesResponse `json:"my_identifiers,omitempty"`
}

type LocaleResponse struct {
	Locale    *string `json:"locale" validate:"required"`
	Name      *string `json:"name"`
	CountryId *string `json:"country_id"`
}

type IdentifierResponse struct {
	ID   *string `json:"id"`
	Type *string `json:"type"`
	Name *string `json:"name"`
}

type OnboardingAttributesResponse struct {
	Sms      *bool `json:"sms",omitempty`
	Whatsapp *bool `json:"whatsapp",omitempty`
	Facebook *bool `json:"facebook",omitempty`
}

type AudioConfigResponse struct {
	UseLatest    *bool   `json:"use_latest,omitempty"`
	EnableProxy  *bool   `json:"enable_proxy,omitempty"`
	ServerRegion *string `json:"server_region,omitempty"`
}

type AllowedLocaleAttributesResponse struct {
	Locale *string `json:"locale",omitempty`
	Name   *string `json:"name",omitempty`
}

type PhoneValidationAttributes struct {
	StartValues []*string `json:"start_values"`
	MinValue    *int      `json:"min_value,omitempty"`
	MaxValue    *int      `json:"max_value,omitempty"`
}

type CountriesAttributesResponse struct {
	ID                           *string                           `json:"id"`
	Name                         *string                           `json:"name"`
	FullName                     *string                           `json:"full_name,omitempty"`
	Locale                       *string                           `json:"locale,omitempty"`
	IsoCode                      *string                           `json:"iso_code,omitempty"`
	CallingCode                  *string                           `json:"calling_code,omitempty"`
	Currency                     *string                           `json:"currency,omitempty"`
	Flag                         *string                           `json:"flag,omitempty"`
	CurrencySubUnit              *string                           `json:"currency_sub_unit,omitempty"`
	CurrencySymbol               *string                           `json:"currency_symbol,omitempty"`
	CanUpdateCurriculumCountry   *bool                             `json:"can_update_curriculum_country,omitempty"`
	AudioConfigResponse          *AudioConfigResponse              `json:"audio_config,omitempty"`
	OnboardingAttributesResponse *OnboardingAttributesResponse     `json:"onboarding,omitempty"`
	PhoneValidation              *PhoneValidationAttributes        `json:"phone_validation,omitempty"`
	AllowedLocales               []AllowedLocaleAttributesResponse `json:"allowed_locales,omitempty"`
	PaymentEnabled               *bool                             `json:"payment_enabled,omitempty"`
	Hidden                       *bool                             `json:"hidden,omitempty"`
	Attributes                   map[string]interface{}            `json:"attributes,omitempty"`
}

type CollegeAttributesResponse struct {
	Name *string `json:"name"`
}
type BoardsAttributesResponse struct {
	ID    *string `json:"id"`
	Name  *string `json:"name,omitempty"`
}
type GradesAttributesResponse struct {
	ID    *string `json:"id"`
	Grade *int    `json:"grade"`
	Name  *string `json:"name,omitempty"`
}

type DegreesAttributesResponse struct {
	ID   *string `json:"id"`
	Name *string `json:"name,omitempty"`
}

type MajorsAttributesResponse struct {
	ID   *string `json:"id"`
	Name *string `json:"name,omitempty"`
}

type LegacyResponse struct {
	ID    string `json:"id"`
	TagId string `json:"tag_id,omitempty"`
	Type  string `json:"type"`
}

type curriculumTypeList struct {
	Default            string `json:"default"`
	Misc               string `json:"misc"`
	Root               string `json:"root"`
	K12                string `json:"k12"`
	University         string `json:"university"`
	TestPrep           string `json:"test_prep"`
	GeneralTestPrep    string `json:"general_test_prep"`
	K12TestPrep        string `json:"k12_test_prep"`
	UniversityTestPrep string `json:"university_test_prep"`
	Skill              string `json:"skill"`
	GeneralSkill       string `json:"general_skill"`
	K12Skill           string `json:"k12_skill"`
	UniversitySkill    string `json:"university_skill"`
}

type tagTypeList struct {
	Country    string `json:"country"`
	Grade      string `json:"grade"`
	Subject    string `json:"subject"`
	Curriculum string `json:"curriculum"`
	Degree     string `json:"degree"`
	Major      string `json:"major"`
	Course     string `json:"course"`
	Section    string `json:"section"`
	Test       string `json:"test"`
	Skill      string `json:"skill"`
	Chapter    string `json:"chapter"`
	Topic      string `json:"topic"`
	Board      string `json:"board"`
}

type tagGroupList struct {
	Curriculum string `json:"curriculum"`
	Content    string `json:"content"`
	Identifier string `json:"identifier"`
}

type accessList struct {
	Global  string `json:"global"`
	Teacher string `json:"teacher"`
}

type DefaultTags struct {
	MiscellaneousTag *TagResponse `json:"default_chapter"`
	ResourceTag      *TagResponse `json:"default_topic"`
}

var CurriculumTypeEnum = &curriculumTypeList{
	Default:            "default",
	Misc:               "misc",
	Root:               "root",
	K12:                "k12",
	University:         "university",
	TestPrep:           "test_prep",
	GeneralTestPrep:    "general_test_prep",
	K12TestPrep:        "k12_test_prep",
	UniversityTestPrep: "university_test_prep",
	Skill:              "skill",
	GeneralSkill:       "general_skill",
	K12Skill:           "k12_skill",
	UniversitySkill:    "university_skill",
}

var TagTypeEnum = &tagTypeList{
	Country:    "country",
	Grade:      "grade",
	Subject:    "subject",
	Curriculum: "curriculum",
	Degree:     "degree",
	Major:      "major",
	Course:     "course",
	Section:    "section",
	Test:       "test",
	Skill:      "skill",
	Chapter:    "chapter",
	Topic:      "topic",
	Board:		"board",
}

var TagGroupEnum = &tagGroupList{
	Curriculum: "curriculum",
	Content:    "content",
	Identifier: "identifier",
}

var AccessEnum = &accessList{
	Global:  "global",
	Teacher: "teacher",
}

type TagsRepository interface {
	FetchTags(*string) (*Tags, error)
	FetchFilteredTags(*string, *string) ([]*Tags, error)
	FetchByInTags([]*string) ([]*Tags, error)
	FetchByTagGroup(*string, *string) ([]*Tags, error)
	CreateTags(*sql.Tx, *Tags) (*string, error)
	DeleteTags(*sql.Tx, *string) error
	UpdateLocale(*sql.Tx, bool, *string) error
	UpdateTag(*UpdateTag) error
	ToggleTags(bool, []*string) error
	FetchFilteredTagsPaginated(*string, *string, *int, *int) ([]*Tags, error)
	FetchFilteredTagsPaginatedForAdmin(*string, *string, *int, *int) ([]*Tags, error)
}

type TagsService interface {
	FetchTags(*string) (*Tags, error)
	GetTagsConcurrent([]*string) ([]*Tags, error)
	FetchFilteredTags(*string, *string) ([]*Tags, error)
	FetchByTagGroup(*string, *string) ([]*Tags, error)
	FetchByInTags([]*string) ([]*Tags, error)
	CreateTags(*sql.Tx, *Tags) (*string, error)
	CreateParentTagMapping(*sql.Tx, *ParentTagMapping) error
	FetchParentTagMappings(*string) ([]*ParentTagMapping, error)
	FetchByInParentTagMappings([]*string) ([]*ParentTagMapping, error)
	FetchFilteredParentTagMappings(*string, *string) ([]*ParentTagMapping, error)
	FetchParentTagMappingByParentTagIdTagId(*string, *string) (*ParentTagMapping, error)
	FetchByInParentTagMappingsByParentTagIdTagIds([]*string, *string) ([]*ParentTagMapping, error)
	DeleteTags(*sql.Tx, *string) error
	UpdateLocale(*sql.Tx, bool, *string) error
	ToggleHideParentTagMapping(*sql.Tx, bool, *string, *string) error
	DeleteParentTagMapping(*sql.Tx, *string) error
	CreateTagLocaleMapping(*sql.Tx, *TagLocaleMapping) error
	FetchTagLocaleMappings(*string) ([]*TagLocaleMapping, error)
	FetchTagLocalesByTagIds(ids []*string, locale *string, countryId *string) ([]*TagLocaleMapping, error)
	FetchFilteredTagsPaginated(*string, *string, *int, *int) ([]*Tags, error)
	FetchFilteredTagsPaginatedForAdmin(*string, *string, *int, *int) ([]*Tags, error)
	IsCollegePresent(*string, *string) (bool, error)
	FetchTagLocaleMappingsByLocale([]*Tags, *string, *string) ([]*Tags, error)
	FetchTagLocaleMappingsByLocaleForContext([]*Tags, *string, *string) ([]*Tags, error)
	DeleteTagLocaleMapping(*sql.Tx, *TagLocaleMapping) error
	UpdateTag(*UpdateTag) error
	ToggleTags(bool, []*string) error
	FetchTagOrders(*string, *string) ([]*ParentTagMapping, error)
	UpdateTagOrders(*sql.Tx, []*Order, *string, *string) error
	OrderTags(tags []*Tags, tagType *string, curriculumType *string, hierarchy *string) ([]*Tags, error)
	FetchTagIdFromLegacyId(*string, *string) ([]*LegacyTagMapping, error)
	FetchLegacyIdFromTagId(*string) ([]*LegacyTagMapping, error)
	FetchLegacyIdFromTagIds([]*string) ([]*LegacyTagMapping, error)
	FetchGradesFromProductId(*string) ([]*GradeProduct, error)
}
