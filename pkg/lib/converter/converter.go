package converter

func ConvertToStringPtr(input string) *string {
	if len(input) == 0 {
		return nil
	} else {
		return &input
	}
}

func ConvertToInt64Ptr(input int64) *int64 {
	if input == 0 {
		return nil
	} else {
		return &input
	}
}

func ConvertToIntPtr(input int) *int {
	if input == 0 {
		return nil
	} else {
		return &input
	}
}
