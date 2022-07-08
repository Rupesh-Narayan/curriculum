package domain

type StudentTagsService interface {
	GetCountries(tags *GetCountries) (*GetCountriesResponse, error)
	GetCountriesNew(tags *GetCountriesNew) (*GetCountriesNewResponse, error)
	GetGrades(tags *GetTags) (*GetGradesResponse, error)
	GetBoards(tags *GetTags) (*GetBoardsResponse, error)
	GetDegrees(tags *GetTags) (*GetDegreesResponse, error)
	GetMajors(tags *GetTags) (*GetMajorsResponse, error)
}
