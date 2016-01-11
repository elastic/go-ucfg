package ucfg

import "strings"

type tagOptions struct {
	squash bool
}

func parseTags(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	opts := tagOptions{}
	for _, opt := range s[1:] {
		if opt == "squash" {
			opts.squash = true
		}
	}
	return s[0], opts
}

func fieldName(tagName, structName string) string {
	if tagName != "" {
		return tagName
	}
	return strings.ToLower(structName)
}
