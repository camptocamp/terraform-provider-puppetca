package log

func MergeFields(fieldsList ...map[string]any) map[string]any {
	mergedFields := make(map[string]any)

	for _, fields := range fieldsList {
		for key, value := range fields {
			mergedFields[key] = value
		}
	}

	return mergedFields
}

func ErrorField(err error) map[string]any {
	return map[string]any{
		"error": err,
	}
}
