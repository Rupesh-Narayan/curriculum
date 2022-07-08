package domain

type TeacherTagsService interface {
	GetTeacherTags(tags *GetTeacherTags) (*GetTagsResponse, error)
	GetTestsSkillsForLibrary(gtt *GetTeacherTags) (*GetTagsResponse, error)
	SearchTeacherTags(tags *GetTeacherTags) ([]*TagResponse, error)
	GetCountriesTagsNew(tags *GetCountriesNew) (getTagResponse *GetCountriesNewResponse, err error)
}
