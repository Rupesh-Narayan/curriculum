package response

import "bitbucket.org/noon-micro/curriculum/pkg/domain"

type AdminTagResponseSearchDTO struct {
	ID         *string                `json:"id"`
	Type       *string                `json:"type"`
	Name       *string                `json:"name"`
	LocaleName *string                `json:"locale_name"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type TagLocaleInfoResponseDTO struct {
	ID             *string                  `json:"id"`
	Type           *string                  `json:"type"`
	CurriculumType *string                  `json:"curriculum_type,omitempty"`
	Grade          *int                     `json:"grade,omitempty"`
	Name           *string                  `json:"name"`
	Attributes     map[string]interface{}   `json:"attributes"`
	Locale         []*domain.LocaleResponse `json:"locales"`
}

type GetTeacherTagsResponseDTO struct {
	Tags []*TeacherTagResponseDTO
	Meta *MetaResponseDTO `json:"meta,omitempty"`
}

type GetAdminTagsResponseDTO struct {
	Tags []*AdminTagResponseDTO
	Meta *MetaResponseDTO `json:"meta,omitempty"`
}

type GetRpcTagsResponseDTO struct {
	Tags []*RpcTagResponseDTO
	Meta *MetaResponseDTO `json:"meta,omitempty"`
}

type MetaResponseDTO struct {
	Next *int `json:"next"`
}

type TeacherTagResponseDTO struct {
	ID         *string                `json:"id"`
	Type       *string                `json:"type"`
	Name       *string                `json:"name"`
	LocaleName *string                `json:"locale_name,omitempty"`
	Attributes map[string]interface{} `json:"attributes"`
}

type AdminTagResponseDTO struct {
	ID         *string                `json:"id"`
	Type       *string                `json:"type"`
	Name       *string                `json:"name"`
	LocaleName *string                `json:"locale_name,omitempty"`
	Hidden     bool                   `json:"hidden"`
	Attributes map[string]interface{} `json:"attributes"`
}

type RpcTagResponseDTO struct {
	ID            *string `json:"id"`
	Type          *string `json:"type"`
	Name          *string `json:"name"`
	LocaleName    *string `json:"locale_name,omitempty"`
	BackgroundPic *string `json:"background_pic"`
	Color         *string `json:"color"`
	Pic           *string `json:"pic"`
	NegativePic   *string `json:"negative_pic"`
}
