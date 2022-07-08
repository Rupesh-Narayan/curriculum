package mapper

import (
	"bitbucket.org/noon-micro/curriculum/pkg/domain"
	noonerror "bitbucket.org/noon-micro/curriculum/pkg/lib/error"
	"bitbucket.org/noon-micro/curriculum/pkg/lib/flow"
	"bitbucket.org/noon-micro/curriculum/pkg/service/constant"
	"encoding/json"
	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	"strings"
)

func CreateGetTagResponse(tagData []*domain.Tags, tags *domain.GetTags, hiddenSet map[string]bool, next *int) (*domain.GetTagsResponse, error) {
	return GetTagResponse(tagData, tags.Type, tags.CurriculumType, hiddenSet, next)
}

func GetTagResponse(tagData []*domain.Tags, tagType *string, curriculumType *string, hiddenSet map[string]bool, next *int) (*domain.GetTagsResponse, error) {
	curriculumHierarchy, _ := flow.GetCurriculum(curriculumType)
	var ok bool
	meta := new(domain.MetaResponse)
	var tagResponses []*domain.TagResponse
	getTagResponse := new(domain.GetTagsResponse)
	if tagType != nil {
		_, ok = curriculumHierarchy[*tagType]
		if ok {
			isIdentifier := false
			meta.IsIdentifier = &isIdentifier
			isOrdered := curriculumHierarchy[*tagType].IsOrdered
			meta.IsOrdered = &isOrdered
		}
	}
	meta.Next = next
	for _, v := range tagData {
		tagResponse := new(domain.TagResponse)
		hidden, ok := hiddenSet[*v.ID]
		if !ok {
			tagResponse.Hidden = false
		} else {
			tagResponse.Hidden = hidden
		}
		tagResponse.ID = v.ID
		tagResponse.Type = v.Type
		tagResponse.Attributes = v.Attributes
		tagResponse.Name = v.Name
		tagResponse.LocaleName = v.LocaleName
		tagResponse.CurriculumType = &v.CurriculumType
		tagResponses = append(tagResponses, tagResponse)
	}
	getTagResponse.Tags = tagResponses
	getTagResponse.Meta = meta
	return getTagResponse, nil
}

func GetTagResponseForProduct(tagData []*domain.Tags, tagType *string, curriculumType *string, hiddenSet map[string]bool, next *int) (*domain.GetTagsResponseForProduct, error) {
	curriculumHierarchy, _ := flow.GetCurriculum(curriculumType)
	var ok bool
	meta := new(domain.MetaResponse)
	var tagResponses []*domain.TagResponseForProduct
	getTagResponse := new(domain.GetTagsResponseForProduct)
	if tagType != nil {
		_, ok = curriculumHierarchy[*tagType]
		if ok {
			isIdentifier := false
			meta.IsIdentifier = &isIdentifier
			isOrdered := curriculumHierarchy[*tagType].IsOrdered
			meta.IsOrdered = &isOrdered
		}
	}
	meta.Next = next
	for _, v := range tagData {
		tagResponse := new(domain.TagResponseForProduct)
		if v.Attributes != nil {
			data, _ := json.Marshal(v.Attributes)
			_ = json.Unmarshal(data, tagResponse)
		}
		hidden, ok := hiddenSet[*v.ID]
		if !ok {
			tagResponse.Hidden = false
		} else {
			tagResponse.Hidden = hidden
		}
		tagResponse.ID = v.ID
		tagResponse.Type = v.Type
		tagResponse.Name = v.Name
		tagResponse.LocaleName = v.LocaleName
		tagResponse.CurriculumType = &v.CurriculumType
		tagResponses = append(tagResponses, tagResponse)
	}
	getTagResponse.Tags = tagResponses
	getTagResponse.Meta = meta
	return getTagResponse, nil
}

func CreateGetCountriesResponse(tagData []*domain.Tags, tags *domain.GetCountries, hiddenSet map[string]bool, next *int) (*domain.GetCountriesResponse, error) {
	curriculumHierarchy, err := flow.GetCurriculum(tags.CurriculumType)
	if err != nil {
		return nil, err
	}
	var ok bool
	meta := new(domain.MetaResponse)
	var tagResponses []*domain.TagResponse
	getTagResponse := new(domain.GetCountriesResponse)
	if tags.Type != nil {
		_, ok = curriculumHierarchy[*tags.Type]
	}
	if ok {
		isIdentifier := curriculumHierarchy[*tags.Type].IsIdentifier
		meta.IsIdentifier = &isIdentifier
		isOrdered := curriculumHierarchy[*tags.Type].IsOrdered
		meta.IsOrdered = &isOrdered
	}
	meta.Next = next
	for _, v := range tagData {
		tagResponse := new(domain.TagResponse)
		hidden, ok := hiddenSet[*v.ID]
		if !ok {
			tagResponse.Hidden = false
		} else {
			tagResponse.Hidden = hidden
		}
		tagResponse.ID = v.ID
		tagResponse.Type = v.Type
		tagResponse.Attributes = v.Attributes
		tagResponse.Name = v.Name
		tagResponses = append(tagResponses, tagResponse)
	}
	getTagResponse.Tags = tagResponses
	getTagResponse.Meta = meta
	return getTagResponse, nil
}

func CreateGetCountriesNewResponse(tagData []*domain.Tags, locale *string, ipDomain *string, next *int, admin bool) (*domain.GetCountriesNewResponse, error) {
	meta := new(domain.CountriesNewMetaResponse)
	var countriesResponses []*domain.CountriesAttributesResponse
	getTagResponse := new(domain.GetCountriesNewResponse)
	meta.Next = next
	var defaultSelectedCountry = new(domain.CountriesAttributesResponse)
	for _, v := range tagData {
		countriesAttributeResponse := new(domain.CountriesAttributesResponse)
		defaultPaymentEnabled := false
		countriesAttributeResponse.PaymentEnabled = &defaultPaymentEnabled
		if v.Attributes["full_name"] != nil {
			fullNameAttribute := v.Attributes["full_name"].(string)
			countriesAttributeResponse.FullName = &fullNameAttribute
		}

		if v.Attributes["payment_enabled"] != nil {
			paymentEnabledAttribute := v.Attributes["payment_enabled"].(bool)
			countriesAttributeResponse.PaymentEnabled = &paymentEnabledAttribute
		}

		if v.Attributes["locale"] != nil {
			localeAttribute := v.Attributes["locale"].(string)
			countriesAttributeResponse.Locale = &localeAttribute
		}

		if v.Attributes["calling_code"] != nil {
			callingCodeAttribute := v.Attributes["calling_code"].(string)
			countriesAttributeResponse.CallingCode = &callingCodeAttribute
		}

		if v.Attributes["currency"] != nil {
			currencyAttribute := v.Attributes["currency"].(string)
			countriesAttributeResponse.Currency = &currencyAttribute
		}

		if v.Attributes["flag"] != nil {
			flagAttribute := v.Attributes["flag"].(string)
			countriesAttributeResponse.Flag = &flagAttribute
		}

		if v.Attributes["currency_sub_unit"] != nil {
			currencySubUnitAttribute := v.Attributes["currency_sub_unit"].(string)
			countriesAttributeResponse.CurrencySubUnit = &currencySubUnitAttribute
		}

		if v.Attributes["currency_symbol"] != nil {
			currencySymbolAttribute := v.Attributes["currency_symbol"].(string)
			countriesAttributeResponse.CurrencySymbol = &currencySymbolAttribute
		}

		if v.Attributes["can_update_curriculum_country"] != nil {
			canUpdateCurriculumCountry := v.Attributes["can_update_curriculum_country"].(bool)
			countriesAttributeResponse.CanUpdateCurriculumCountry = &canUpdateCurriculumCountry
		}

		if v.Attributes["onboarding"] != nil {
			foomap := v.Attributes["onboarding"]
			aa := foomap.(map[string]interface{})
			onboardingAttributes := new(domain.OnboardingAttributesResponse)
			smsAttribute := aa["sms"].(bool)
			onboardingAttributes.Sms = &smsAttribute
			whatsappAttribute := aa["whatsapp"].(bool)
			onboardingAttributes.Whatsapp = &whatsappAttribute
			facebookAttribute := aa["facebook"].(bool)
			onboardingAttributes.Facebook = &facebookAttribute
			countriesAttributeResponse.OnboardingAttributesResponse = onboardingAttributes
		}

		if v.Attributes["audio_config"] != nil {
			audioConfig := v.Attributes["audio_config"]
			aa, ok := audioConfig.(map[string]interface{})
			if ok {
				audioConfigAttributes := new(domain.AudioConfigResponse)
				useLatestAttribute, ok := aa["use_latest"].(bool)
				if ok {
					audioConfigAttributes.UseLatest = &useLatestAttribute
				}
				enableProxyAttribute, ok := aa["enable_proxy"].(bool)
				if ok {
					audioConfigAttributes.EnableProxy = &enableProxyAttribute
				}
				serverRegionAttribute, ok := aa["server_region"].(string)
				if ok {
					audioConfigAttributes.ServerRegion = &serverRegionAttribute
				}
				countriesAttributeResponse.AudioConfigResponse = audioConfigAttributes
			}
		}

		if v.Attributes["phone_validation"] != nil {
			foomap := v.Attributes["phone_validation"]
			phoneValidationKeys := foomap.(map[string]interface{})
			phoneValidationAttributes := new(domain.PhoneValidationAttributes)
			startValues := phoneValidationKeys["start_values"].([]interface{})
			aString := make([]*string, len(startValues))
			for i, v3 := range startValues {
				var aStringaa = v3.(string)
				aString[i] = &aStringaa
			}
			phoneValidationAttributes.StartValues = aString
			minValue := phoneValidationKeys["min_value"].(float64)
			var minValueInt int = int(minValue)
			phoneValidationAttributes.MinValue = &minValueInt
			maxValue := phoneValidationKeys["max_value"].(float64)
			var maxValueInt int = int(maxValue)
			phoneValidationAttributes.MaxValue = &maxValueInt
			countriesAttributeResponse.PhoneValidation = phoneValidationAttributes
		}

		if v.Attributes["allowed_locales"] != nil {
			foomap := v.Attributes["allowed_locales"]
			phoneValidationAttributes := []domain.AllowedLocaleAttributesResponse{}

			mapstructure.Decode(foomap, &phoneValidationAttributes)
			countriesAttributeResponse.AllowedLocales = phoneValidationAttributes
		}

		if v.Attributes["iso_code"] != nil {
			isoCodeAttribute := v.Attributes["iso_code"].(string)
			countriesAttributeResponse.IsoCode = &isoCodeAttribute
			if ipDomain != nil && *countriesAttributeResponse.IsoCode == *ipDomain {
				meta.SelectedCountry = countriesAttributeResponse
			}
			if *countriesAttributeResponse.IsoCode == "SA" {
				defaultSelectedCountry = countriesAttributeResponse
			}
		}

		countriesAttributeResponse.ID = v.ID
		countriesAttributeResponse.Name = v.Name
		if locale != nil && *locale == constant.DefaultLocale && countriesAttributeResponse.FullName != nil && len(*countriesAttributeResponse.FullName) > 0 {
			countriesAttributeResponse.Name = countriesAttributeResponse.FullName
		}
		if admin {
			hidden := !v.Publish
			countriesAttributeResponse.Hidden = &hidden
			countriesAttributeResponse.Attributes = v.Attributes
		}
		countriesResponses = append(countriesResponses, countriesAttributeResponse)
	}
	if meta.SelectedCountry == nil {
		meta.SelectedCountry = defaultSelectedCountry
	}
	getTagResponse.Tags = countriesResponses
	getTagResponse.Meta = meta
	return getTagResponse, nil
}

func CreateGetTagResponseWithIdentifiers(tagData []*domain.Tags, tags *domain.GetTags, setIdentifiers map[string]*domain.Tags, setTagIdentifiers map[string][]*string, next *int) (*domain.GetTagsResponse, error) {
	curriculumHierarchy, err := flow.GetCurriculum(tags.CurriculumType)
	if err != nil {
		return nil, err
	}
	meta := new(domain.MetaResponse)
	var tagResponses []*domain.TagResponse
	getTagResponse := new(domain.GetTagsResponse)
	_, ok := curriculumHierarchy[*tags.Type]
	if ok {
		isIdentifier := false
		meta.IsIdentifier = &isIdentifier
		isOrdered := curriculumHierarchy[*tags.Type].IsOrdered
		meta.IsOrdered = &isOrdered
	}
	meta.Next = next
	for _, v := range tagData {
		tagResponse := new(domain.TagResponse)
		tagResponse.ID = v.ID
		tagResponse.Type = v.Type
		tagResponse.Attributes = v.Attributes
		tagResponse.Name = v.Name
		tagResponse.Hidden = false
		var identifierTags []*domain.IdentifierResponse
		identifiers, ok := setTagIdentifiers[*v.ID]
		if ok {
			for _, v := range identifiers {
				if *v == "true" || *v == "false" {
					if *v == "true" {
						tagResponse.Hidden = true
					}
					continue
				}
				if strings.Contains(*v, ".") {
					tagResponse.Root = v
				} else {
					identifierData, okInner := setIdentifiers[*v]
					if okInner && identifierData != nil {
						identifier := new(domain.IdentifierResponse)
						identifier.ID = identifierData.ID
						identifier.Name = identifierData.Name
						identifier.Type = identifierData.Type
						identifierTags = append(identifierTags, identifier)
					}
				}
			}
		}
		tagResponse.Identifiers = identifierTags
		tagResponses = append(tagResponses, tagResponse)
	}
	getTagResponse.Tags = tagResponses
	getTagResponse.Meta = meta
	return getTagResponse, nil
}

func CreateElasticTagEntity(tagId *string, ct *domain.CreateTags, parents []*string, access string) (cte *domain.CreateTagElastic, err error) {
	mappedCurriculumType, err := flow.CurriculumMapper(ct.CurriculumType)
	if err != nil {
		return
	}
	cte = new(domain.CreateTagElastic)
	if err = copier.Copy(cte, ct); err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "mapperError")
	}
	var locale = constant.DefaultLocale
	name := domain.TagName{
		Value:  ct.Name,
		Locale: &locale,
	}
	cte.ID = tagId
	cte.CurriculumType = mappedCurriculumType
	cte.Name = []*domain.TagName{&name}
	cte.Parents = parents
	cte.Deleted = false
	if len(access) > 0 {
		cte.Access = &access
	}
	return cte, nil
}

func GetElasticTagEntity(tags interface{}, parents []*string, hiddenParents []*string, access string, curriculumType *string, creatorType string, tagGroup *string, start, limit int) (gte *domain.GetTagsElastic, err error) {
	gte = new(domain.GetTagsElastic)
	if err = copier.Copy(gte, tags); err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "mapperError")
	}
	mappedCurriculumType, err := flow.CurriculumMapper(curriculumType)
	if err != nil {
		return
	}
	gte.TagGroup = tagGroup
	gte.CurriculumType = mappedCurriculumType
	gte.CreatorType = &creatorType
	if len(hiddenParents) > 0 {
		gte.HiddenParents = hiddenParents
	}
	gte.Parents = parents
	if len(access) > 0 {
		gte.Access = &access
	}
	gte.Start = start
	gte.Limit = limit
	return gte, nil
}

func GetElasticTagEntityWithoutCurriculumType(tags interface{}, parents []*string, hiddenParents []*string, access string, creatorType string, tagGroup *string, start, limit int) (gte *domain.GetTagsElastic, err error) {
	gte = new(domain.GetTagsElastic)
	if err = copier.Copy(gte, tags); err != nil {
		return nil, noonerror.New(noonerror.ErrInternalServer, "mapperError")
	}
	gte.TagGroup = tagGroup
	gte.CreatorType = &creatorType
	if len(hiddenParents) > 0 {
		gte.HiddenParents = hiddenParents
	}
	gte.Parents = parents
	if len(access) > 0 {
		gte.Access = &access
	}
	gte.Start = start
	gte.Limit = limit
	return gte, nil
}

func CreateDefaultTagResponse(tagData []*domain.Tags) (*domain.DefaultTags, error) {

	defaultTags := new(domain.DefaultTags)

	for _, v := range tagData {
		tagResponse := new(domain.TagResponse)
		tagResponse.Hidden = false
		tagResponse.ID = v.ID
		tagResponse.Type = v.Type
		tagResponse.Attributes = v.Attributes
		tagResponse.Name = v.Name
		tagResponse.LocaleName = v.LocaleName
		if *v.Type == "chapter" {
			defaultTags.MiscellaneousTag = tagResponse
		} else {
			defaultTags.ResourceTag = tagResponse
		}
	}

	return defaultTags, nil
}

func SuggestedTagResponse(tagData []*domain.Tags) (suggestedTags []*domain.SuggestedTags, err error) {
	for _, v := range tagData {
		tagResponse := new(domain.SuggestedTags)
		tagResponse.ID = v.ID
		tagResponse.Type = v.Type
		tagResponse.Name = v.Name
		suggestedTags = append(suggestedTags, tagResponse)
	}

	return suggestedTags, nil
}

func GetGeoRequestEntity(ipAddress *string) (getGeoIpRequestDto *domain.GetGeoIp) {
	getGeoIpRequestDto = new(domain.GetGeoIp)
	getGeoIpRequestDto.Ip = ipAddress
	return getGeoIpRequestDto
}

func CreateGetGradesResponse(tagData []*domain.Tags, hasCollege *bool) (*domain.GetGradesResponse, error) {
	GradeTagMap := map[string]int{
		"251": 1,
		"252": 2,
		"253": 3,
		"254": 4,
		"255": 5,
		"256": 6,
		"257": 7,
		"258": 8,
		"259": 9,
		"260": 10,
		"261": 11,
		"262": 12,
	}
	defaultGrade := constant.DefaultGrade
	meta := new(domain.GradesMetaResponse)
	var gradeResponses []*domain.GradesAttributesResponse
	getGradesResponse := new(domain.GetGradesResponse)
	for _, v := range tagData {
		gradeResponse := new(domain.GradesAttributesResponse)
		gradeResponse.ID = v.ID
		gradeResponse.Name = v.Name
		if val, ok := GradeTagMap[*gradeResponse.ID]; ok {
			gradeResponse.Grade = &val
		}
		if gradeResponse.Grade == nil {
			gradeResponse.Grade = &defaultGrade
		}
		gradeResponses = append(gradeResponses, gradeResponse)
	}
	getGradesResponse.Grades = gradeResponses
	hasUniversityFlag := hasCollege
	meta.HasUniversity = hasUniversityFlag
	getGradesResponse.Meta = meta
	return getGradesResponse, nil
}

func CreateGetDegreesResponse(tagData []*domain.Tags, next *int) (*domain.GetDegreesResponse, error) {
	var degreesResponses []*domain.DegreesAttributesResponse
	getDegreesResponse := new(domain.GetDegreesResponse)
	for _, v := range tagData {
		degreeResponse := new(domain.DegreesAttributesResponse)
		degreeResponse.ID = v.ID
		degreeResponse.Name = v.Name
		degreesResponses = append(degreesResponses, degreeResponse)
	}
	getDegreesResponse.Degrees = degreesResponses
	meta := new(domain.DegreesMetaResponse)
	meta.Next = next
	getDegreesResponse.Meta = meta
	return getDegreesResponse, nil
}
func CreateGetBoardsResponse(tagData []*domain.Tags, hasCollege *bool) (*domain.GetBoardsResponse, error) {
	var boardsResponses []*domain.BoardsAttributesResponse
	getBoardsResponse := new(domain.GetBoardsResponse)
	for _, v := range tagData {
		boardResponse := new(domain.BoardsAttributesResponse)
		boardResponse.ID = v.ID
		boardResponse.Name = v.Name
		boardsResponses = append(boardsResponses, boardResponse)
	}
	getBoardsResponse.Boards = boardsResponses
	meta := new(domain.BoardsMetaResponse)
	hasUniversityFlag := hasCollege
	meta.HasUniversity = hasUniversityFlag
	getBoardsResponse.Meta = meta
	return getBoardsResponse, nil
}

func CreateGetMajorsResponse(tagData []*domain.Tags, next *int) (*domain.GetMajorsResponse, error) {
	var majorsResponses []*domain.MajorsAttributesResponse
	getMajorsResponse := new(domain.GetMajorsResponse)
	for _, v := range tagData {
		majorResponse := new(domain.MajorsAttributesResponse)
		majorResponse.ID = v.ID
		majorResponse.Name = v.Name
		majorsResponses = append(majorsResponses, majorResponse)
	}
	meta := new(domain.MajorsMetaResponse)
	meta.Next = next
	getMajorsResponse.Meta = meta
	getMajorsResponse.Degrees = majorsResponses
	return getMajorsResponse, nil
}
