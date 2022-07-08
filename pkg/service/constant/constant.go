package constant

const (
	RootCurriculum      = "root"
	DerivedCurriculum   = "derived"
	HierarchyCurriculum = "hierarchy"
	ReadAccessType      = "read"
	WriteAccessType     = "write"
	TagLimit            = 50
	OrderMax            = 1000
	DefaultLocale       = "en"
	DefaultGrade        = 99
)

var (
	UniversityProductsMap = map[string]struct{}{
		"24":  {},
		"27":  {},
		"29":  {},
		"30":  {},
		"44":  {},
		"54":  {},
		"95":  {},
		"98":  {},
		"105": {},
	}
	GradeTagMap = map[string]string{
		"1":  "251",
		"2":  "252",
		"3":  "253",
		"4":  "254",
		"5":  "255",
		"6":  "256",
		"7":  "257",
		"8":  "258",
		"9":  "259",
		"10": "260",
		"11": "261",
		"12": "262",
	}
	MultiGradeMap = map[string]struct{}{
		"9": {},
	}
)
