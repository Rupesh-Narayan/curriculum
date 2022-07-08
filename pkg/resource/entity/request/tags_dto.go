package request

type CreateTagsForAdminDTO struct {
	Type           *string                `json:"type" validate:"required,min=1"`
	Name           *string                `json:"name" validate:"required,min=1"`
	CurriculumType *string                `json:"curriculum_type" validate:"required,oneof=k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill misc"`
	TagGroup       *string                `json:"tag_group" validate:"required,oneof=curriculum content identifier"`
	Access         string                 `json:"access" validate:"oneof=teacher global"`
	CountryId      int                    `json:"country_id"`
	Attributes     map[string]interface{} `json:"attributes"`
	Hierarchy      []*string              `json:"hierarchy" validate:"contains-nil"`
	Identifier     []*string              `json:"identifier" validate:"contains-nil"`
}

type UpdateHierarchyDTO struct {
	ID             *string   `json:"id" validate:"required,min=1"`
	CurriculumType *string   `json:"curriculum_type" validate:"required,oneof=k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill"`
	TagGroup       *string   `json:"tag_group" validate:"required,oneof=curriculum content"`
	Hierarchy      []*string `json:"hierarchy" validate:"contains-nil"`
	Identifier     []*string `json:"identifier" validate:"contains-nil"`
}

type UpdateMultipleHierarchyDTO struct {
	IDs            []*string `json:"ids" validate:"required,contains-nil"`
	Type           *string   `json:"type" validate:"required,min=1"`
	CurriculumType *string   `json:"curriculum_type" validate:"required,oneof=k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill"`
	TagGroup       *string   `json:"tag_group" validate:"required,oneof=content"`
	Hierarchy      []*string `json:"hierarchy" validate:"contains-nil"`
}

type UpdateTagOrderDTO struct {
	Type           *string     `json:"type" validate:"required,min=1"`
	CurriculumType *string     `json:"curriculum_type" validate:"required,oneof=k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill"`
	TagGroup       *string     `json:"tag_group" validate:"required,oneof=curriculum content"`
	Hierarchy      []*string   `json:"hierarchy" validate:"required,contains-nil"`
	Orders         []*OrderDTO `json:"orders" validate:"contains-nil"`
}

type OrderDTO struct {
	ID    *string `json:"id" validate:"required,min=1"`
	Order *int    `json:"order" validate:"required"`
}

type UpdateTagDTO struct {
	ID         *string                `json:"id" validate:"required,min=1"`
	Attributes map[string]interface{} `json:"attributes"`
	Name       *string                `json:"name"`
	Hidden     *bool                  `json:"hidden"`
}

type RemoveHierarchyDTO struct {
	ID             *string   `json:"id" validate:"required,min=1"`
	CurriculumType *string   `json:"curriculum_type" validate:"required,oneof=k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill"`
	TagGroup       *string   `json:"tag_group" validate:"required,oneof=curriculum content"`
	Hierarchy      []*string `json:"hierarchy" validate:"contains-nil"`
	Identifier     []*string `json:"identifier" validate:"contains-nil"`
}

type GetTagsDTO struct {
	Type           *string   `json:"type" validate:"required,min=1"`
	CurriculumType *string   `json:"curriculum_type" validate:"required,oneof=k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill misc"`
	TagGroup       *string   `json:"tag_group" validate:"required,oneof=curriculum content"`
	CountryId      *string   `json:"country_id"`
	Locale         *string   `json:"locale"`
	MultiGrade     *string   `json:"multi_grade"`
	Hierarchy      *string   `json:"hierarchy" validate:"required"`
	Identifier     []*string `json:"identifier"`
}

type GetCountriesNewDTO struct {
	IpAddress string `json:"ip_address"`
	ISOCode	  string  `json:iso_code`
	CountryId      *string   `json:"country_id"`
	Locale         *string   `json:"locale"`
	Start     int    `json:"start"`
	Limit     int    `json:"limit"`
}

type GetGradesRequestDTO struct {
	CountryId string `json:"country_id"`
	Start     int    `json:"start"`
	Limit     int    `json:"limit"`
}

type GetTeacherTagsDTO struct {
	Text           *string   `json:"text"`
	Type           *string   `json:"type" validate:"required,min=1"`
	CurriculumType *string   `json:"curriculum_type" validate:"required,oneof=default k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill misc"`
	TagGroup       *string   `json:"tag_group"`
	CountryId      *string   `json:"country_id"`
	Locale         *string   `json:"locale"`
	CreatorId      *int64    `json:"creator_id" validate:"required"`
	Hierarchy      []*string `json:"hierarchy" validate:"required"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
}

type GetAdminTagsDTO struct {
	Text           *string   `json:"text"`
	Type           *string   `json:"type" validate:"required,min=1"`
	CurriculumType *string   `json:"curriculum_type" validate:"required,oneof=default k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill misc"`
	TagGroup       *string   `json:"tag_group"`
	CountryId      *string   `json:"country_id"`
	Locale         *string   `json:"locale"`
	CreatorId      *int64    `json:"creator_id" validate:"required"`
	Hierarchy      []*string `json:"hierarchy" validate:"required"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
}

type GetTagsSearchDTO struct {
	Text           *string   `json:"text"`
	Locale         *string   `json:"locale"`
	Type           *string   `json:"type"`
	CurriculumType *string   `json:"curriculum_type" validate:"required,oneof=k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill misc"`
	TagGroup       *string   `json:"tag_group" validate:"required,oneof=curriculum content identifier"`
	CountryId      *string   `json:"country_id"`
	Hierarchy      *string   `json:"hierarchy"`
	Identifier     []*string `json:"identifier"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
}

type RemoveIdentifierDTO struct {
	ID       *string `json:"id" validate:"required,min=1"`
	TagGroup *string `json:"tag_group" validate:"required,oneof=identifier"`
}

type MigrateToElastic struct {
	Start *string `json:"start" validate:"required,min=1"`
	End   *string `json:"end" validate:"required,min=1"`
}

type TagLocaleDTO struct {
	ID     *string      `json:"id" validate:"required,min=1"`
	Name   *string      `json:"name" validate:"required,min=1"`
	Locale []*LocaleDTO `json:"locales" validate:"required"`
}

type LocaleDTO struct {
	CountryId *string `json:"country_id" validate:"required,min=1"`
	Locale    *string `json:"locale" validate:"required,min=1"`
}
