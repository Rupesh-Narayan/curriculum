package domain

type AdminTagsService interface {
	CreateAdminTags(*string, *CreateTags) (*TagResponse, error)
	UpdateAdminTags(*string, *UpdateTags) (*TagResponse, error)
	UpdateMultipleAdminTags(*UpdateMultipleTags) ([]*string, error)
	UpdateTagOrder(*string, *UpdateTagOrder) error
	RemoveAdminTagFromHierarchy(*string, *RemoveHierarchy) (*TagResponse, error)
	RemoveIdentifierTag(*string) error
	MigrateToElastic(*string, *string) error
	GetTags(tags *GetTags) (*GetTagsResponse, error)
	GetTag(id *string) (tagResponse *TagResponse, err error)
	GetTagsSearch(tags *GetTags) (*GetTagsResponse, error)
	UpdateTagLocale(*string, *TagLocale) error
	UpdateTag(*UpdateTag) error
	GetAdminTags(tags *GetAdminTags) (getTagResponse *GetTagsResponse, err error)
	GetTestsSkillsForLibrary(gtt *GetAdminTags) (*GetTagsResponse, error)
	GetCountriesTagsNew(tags *GetCountriesNew) (getTagResponse *GetCountriesNewResponse, err error)
}
