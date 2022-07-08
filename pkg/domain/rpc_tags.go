package domain

type RpcTagsService interface {
	GetTags(tags *GetTags) (*GetTagsResponse, error)
	CreateTags(*CreateMultipleTags) ([]*TagResponse, error)
	GetTagsByIds(*GetTagsByIds, bool) ([]*TagResponse, error)
	ValidateHierarchy(*ValidateHierarchy) error
	GetDefaultTags() (*DefaultTags, error)
	GetSuggestedCurriculum(*GetSuggestedTags) ([]*SuggestedTags, error)
	GetLegacyDataFromTagId(*string) ([]*LegacyResponse, error)
	GetLegacyDataFromTagIds([]*string) ([]*LegacyResponse, error)
	GetTagDataFromLegacyId(*string, *string) (legacyResponse []*LegacyResponse, err error)
	GetGradeTags(*string, *string) ([]*LegacyResponse, error)
	GetRpcTags(tags *GetRpcTags) (*GetTagsResponseForProduct, error)
}
