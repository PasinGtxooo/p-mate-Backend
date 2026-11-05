package utils

// UpdateNullableString updates a nullable string field
func UpdateNullableString(target **string, input *string) {
	if input == nil {
		*target = nil
	} else if *input == "" {
		*target = nil
	} else {
		*target = input
	}
}

// UpdateNullableInt updates a nullable int field
func UpdateNullableInt(target **int, input *int) {
	if input == nil {
		*target = nil
	} else if *input == 0 {
		*target = nil
	} else {
		*target = input
	}
}

func UpdateNullableFloat(target **float64, input *float64) {
	if input == nil {
		*target = nil
	} else if *input == 0 {
		*target = nil
	} else {
		*target = input
	}
}

// UpdateNullableJSON updates a nullable JSON field
func UpdateNullableJSON(target *[]byte, input []byte) {
	if len(input) == 0 {
		*target = nil
	} else {
		*target = input
	}
}
