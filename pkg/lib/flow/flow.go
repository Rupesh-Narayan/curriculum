package flow

import (
	. "bitbucket.org/noon-micro/curriculum/pkg/domain"
	noonerror "bitbucket.org/noon-micro/curriculum/pkg/lib/error"
)

type Struct struct {
	Level        int
	IsIdentifier bool
	IsOrdered    bool
}

type CurriculumFactory map[string]Struct

var K12 = map[string]Struct{
	TagTypeEnum.Country:    {Level: 1, IsIdentifier: false},
	TagTypeEnum.Board:      {Level: 2, IsIdentifier: false, IsOrdered: true},
	TagTypeEnum.Grade:      {Level: 3, IsIdentifier: false, IsOrdered: true},
	TagTypeEnum.Subject:    {Level: 4, IsIdentifier: false},
	TagTypeEnum.Curriculum: {Level: 5, IsIdentifier: true},
	TagTypeEnum.Chapter:    {Level: 6, IsIdentifier: false, IsOrdered: true},
	TagTypeEnum.Topic:      {Level: 7, IsIdentifier: false, IsOrdered: true},
}

var University = map[string]Struct{
	TagTypeEnum.Country: {Level: 1, IsIdentifier: false},
	TagTypeEnum.Degree:  {Level: 2, IsIdentifier: false},
	TagTypeEnum.Major:   {Level: 3, IsIdentifier: false},
	TagTypeEnum.Course:  {Level: 4, IsIdentifier: false},
	TagTypeEnum.Section: {Level: 5, IsIdentifier: true},
	TagTypeEnum.Chapter: {Level: 6, IsIdentifier: false, IsOrdered: true},
	TagTypeEnum.Topic:   {Level: 7, IsIdentifier: false, IsOrdered: true},
}

var K12TestPrep = map[string]Struct{
	TagTypeEnum.Country: {Level: 1, IsIdentifier: false},
	TagTypeEnum.Test:    {Level: 2, IsIdentifier: false},
	TagTypeEnum.Section: {Level: 3, IsIdentifier: true},
	TagTypeEnum.Chapter: {Level: 4, IsIdentifier: false, IsOrdered: true},
	TagTypeEnum.Topic:   {Level: 5, IsIdentifier: false, IsOrdered: true},
}

var UniversityTestPrep = map[string]Struct{
	TagTypeEnum.Country: {Level: 1, IsIdentifier: false},
	TagTypeEnum.Test:    {Level: 2, IsIdentifier: false},
	TagTypeEnum.Section: {Level: 3, IsIdentifier: true},
	TagTypeEnum.Chapter: {Level: 4, IsIdentifier: false, IsOrdered: true},
	TagTypeEnum.Topic:   {Level: 5, IsIdentifier: false, IsOrdered: true},
}

var GeneralTestPrep = map[string]Struct{
	TagTypeEnum.Country: {Level: 1, IsIdentifier: false},
	TagTypeEnum.Test:    {Level: 2, IsIdentifier: false},
	TagTypeEnum.Section: {Level: 3, IsIdentifier: true},
	TagTypeEnum.Chapter: {Level: 4, IsIdentifier: false, IsOrdered: true},
	TagTypeEnum.Topic:   {Level: 5, IsIdentifier: false, IsOrdered: true},
}

var K12Skill = map[string]Struct{
	TagTypeEnum.Country: {Level: 1, IsIdentifier: false},
	TagTypeEnum.Skill:   {Level: 2, IsIdentifier: false},
	TagTypeEnum.Section: {Level: 3, IsIdentifier: true},
	TagTypeEnum.Chapter: {Level: 4, IsIdentifier: false, IsOrdered: true},
	TagTypeEnum.Topic:   {Level: 5, IsIdentifier: false, IsOrdered: true},
}

var UniversitySkill = map[string]Struct{
	TagTypeEnum.Country: {Level: 1, IsIdentifier: false},
	TagTypeEnum.Skill:   {Level: 2, IsIdentifier: false},
	TagTypeEnum.Section: {Level: 3, IsIdentifier: true},
	TagTypeEnum.Chapter: {Level: 4, IsIdentifier: false, IsOrdered: true},
	TagTypeEnum.Topic:   {Level: 5, IsIdentifier: false, IsOrdered: true},
}

var GeneralSkill = map[string]Struct{
	TagTypeEnum.Country: {Level: 1, IsIdentifier: false},
	TagTypeEnum.Skill:   {Level: 2, IsIdentifier: false},
	TagTypeEnum.Section: {Level: 3, IsIdentifier: true},
	TagTypeEnum.Chapter: {Level: 4, IsIdentifier: false, IsOrdered: true},
	TagTypeEnum.Topic:   {Level: 5, IsIdentifier: false, IsOrdered: true},
}

var Misc = map[string]Struct{}

func GetCurriculum(curriculumType *string) (cf CurriculumFactory, err error) {
	if curriculumType == nil {
		return nil, noonerror.New(noonerror.ErrParamMissing, "curriculumTypeInvalid")
	}
	if *curriculumType == CurriculumTypeEnum.K12 {
		return K12, nil
	} else if *curriculumType == CurriculumTypeEnum.University {
		return University, nil
	} else if *curriculumType == CurriculumTypeEnum.K12TestPrep {
		return K12TestPrep, nil
	} else if *curriculumType == CurriculumTypeEnum.UniversityTestPrep {
		return UniversityTestPrep, nil
	} else if *curriculumType == CurriculumTypeEnum.GeneralTestPrep {
		return GeneralTestPrep, nil
	} else if *curriculumType == CurriculumTypeEnum.K12Skill {
		return K12Skill, nil
	} else if *curriculumType == CurriculumTypeEnum.UniversitySkill {
		return UniversitySkill, nil
	} else if *curriculumType == CurriculumTypeEnum.GeneralSkill {
		return GeneralSkill, nil
	} else if *curriculumType == CurriculumTypeEnum.Misc {
		return Misc, nil
	}
	return nil, noonerror.New(noonerror.ErrParamMissing, "curriculumTypeInvalid")
}

func CurriculumMapper(curriculumType *string) (ct *string, err error) {
	if curriculumType == nil {
		return nil, noonerror.New(noonerror.ErrParamMissing, "curriculumTypeInvalid")
	}
	var mappedCurriculumType string
	if *curriculumType == CurriculumTypeEnum.K12 || *curriculumType == CurriculumTypeEnum.K12TestPrep || *curriculumType == CurriculumTypeEnum.K12Skill {
		mappedCurriculumType = CurriculumTypeEnum.K12
		return &mappedCurriculumType, nil
	} else if *curriculumType == CurriculumTypeEnum.University || *curriculumType == CurriculumTypeEnum.UniversityTestPrep || *curriculumType == CurriculumTypeEnum.UniversitySkill {
		mappedCurriculumType = CurriculumTypeEnum.University
		return &mappedCurriculumType, nil
	} else if *curriculumType == CurriculumTypeEnum.GeneralTestPrep {
		mappedCurriculumType = CurriculumTypeEnum.TestPrep
		return &mappedCurriculumType, nil
	} else if *curriculumType == CurriculumTypeEnum.GeneralSkill {
		mappedCurriculumType = CurriculumTypeEnum.Skill
		return &mappedCurriculumType, nil
	}
	return curriculumType, nil
}

func GetFilterIdentifier(curriculumType *string) (ct *string) {
	var tagType string
	switch *curriculumType {
	case CurriculumTypeEnum.K12:
		tagType = TagTypeEnum.Curriculum
	default:
		tagType = TagTypeEnum.Section
	}
	return &tagType
}
