package request

type CreateTagsRPCDTO struct {
	Type           *string                `json:"type" validate:"required"`
	CreatorId      *int64                 `json:"creator_id" validate:"required"`
	CurriculumType *string                `json:"curriculum_type" validate:"required,oneof=default k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill misc"`
	TagGroup       *string                `json:"tag_group" validate:"required,oneof=curriculum content identifier"`
	Attributes     map[string]interface{} `json:"attributes"`
	Hierarchy      []*string              `json:"hierarchy" validate:"contains-nil"`
	Identifier     []*string              `json:"identifier" validate:"contains-nil"`
	Tags           []*CreateTagDTO        `json:"tags" validate:"required,min=1,contains-nil"`
}

type CreateTagDTO struct {
	Name      *string `json:"name" validate:"required"`
	CountryId string  `json:"country_id"`
}

type ValidateHierarchyDTO struct {
	Hierarchies    [][]*string `json:"hierarchies" validate:"required,min=1,contains-nil"`
	CurriculumType *string     `json:"curriculum_type" validate:"required,oneof=default k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill misc"`
}

type GetSuggestedTagsDTO struct {
	TagIds         []*string `json:"tag_ids" validate:"required,min=1,contains-nil"`
	CurriculumType *string   `json:"curriculum_type" validate:"required,oneof=default k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill misc"`
	Locale         *string   `json:"locale"`
	CountryId      *string   `json:"country_id"`
}

type GetTagsRPCDTO struct {
	TagIds    []*string `json:"tag_ids" validate:"required,min=1,max=100,contains-nil"`
	Locale    *string   `json:"locale"`
	CountryId *string   `json:"country_id"`
}

type GetTagsByHierarchyRPCDTO struct {
	Type           *string   `json:"type" validate:"required"`
	CurriculumType *string   `json:"curriculum_type" validate:"required,oneof=default k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill misc"`
	TagGroup       *string   `json:"tag_group" validate:"required,oneof=curriculum content"`
	Hierarchy      *string   `json:"hierarchy" validate:"required"`
	Identifier     []*string `json:"identifier" validate:"contains-nil"`
}

type GetLegacyDataFromTagIdDTO struct {
	ID *string `json:"id" validate:"required"`
}

type GetLegacyDataFromTagIdsDTO struct {
	IDs []*string `json:"ids" validate:"required,min=1,contains-nil"`
}

type GetTagDataFromLegacyIdDTO struct {
	Type *string `json:"type" validate:"required"`
	ID   *string `json:"id" validate:"required"`
}

type GetGradeTagsDTO struct {
	Grade     *string `json:"grade" validate:"required"`
	ProductId *string `json:"product_id"`
}

type GetTagsByHierarchyDTO struct {
	Type           *string   `json:"type" validate:"required"`
	CurriculumType *string   `json:"curriculum_type" validate:"required,oneof=default k12 university k12_test_prep university_test_prep general_test_prep k12_skill university_skill general_skill misc"`
	TagGroup       *string   `json:"tag_group"`
	CountryId      *string   `json:"country_id"`
	Locale         *string   `json:"locale"`
	Hierarchy      []*string `json:"hierarchy" validate:"required"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
}

type GenericProductsDTO struct {
	CountryId *string `json:"country_id" validate:"required"`
	BoardId   *string `json:"board_id"`
	GradeId   *string `json:"grade_id" validate:"required"`
	DegreeId  *string `json:"degree_id" validate:"required"`
	MajorId   *string `json:"major_id" validate:"required"`
	Locale    *string `json:"locale"`
	Test      *bool   `json:"test"`
	Skill     *bool   `json:"skill"`
}

type GetK12ProductsDTO struct {
	CountryId *string `json:"country_id" validate:"required"`
	BoardId   *string `json:"board_id"`
	GradeId   *string `json:"grade_id" validate:"required"`
	Locale    *string `json:"locale"`
	Test      *bool   `json:"test"`
	Skill     *bool   `json:"skill"`
}

type GetUniversityProductsDTO struct {
	CountryId *string `json:"country_id" validate:"required"`
	DegreeId  *string `json:"degree_id" validate:"required"`
	MajorId   *string `json:"major_id" validate:"required"`
	Locale    *string `json:"locale"`
	Test      *bool   `json:"test"`
	Skill     *bool   `json:"skill"`
}

type GetTestAndSkillsDTO struct {
	CountryId *string `json:"country_id" validate:"required"`
	Locale    *string `json:"locale"`
	Test      *bool   `json:"test"`
	Skill     *bool   `json:"skill"`
}
