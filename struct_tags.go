package nemesis

import (
	"errors"
	"strings"
)

var TooManyFieldsError = errors.New("too many fields in struct tag")

// ParseStructTags takes a comma separated tag string and structures it into a map based on a schema
func ParseStructTags(tag string, schema []string) (map[string]string, error) {
	parts := strings.Split(tag, ",")
	if len(parts) > len(schema) {
		return nil, TooManyFieldsError
	}
	var result = make(map[string]string)
	for i := range parts {
		if parts[i] != "" {
			result[schema[i]] = parts[i]
		}
	}
	return result, nil
}
