package executors

import (
	"encoding/json"
	"strings"
	"text/template"
)

var (
	funcMap = template.FuncMap{
		"toJson":       toJson,
		"toPrettyJson": toPrettyJson,
		"indent":       indent,
		"nindent":      nindent,
	}
)

// toJson encodes an item into a JSON string
func toJson(v interface{}) string {
	// ref: https://github.com/Masterminds/sprig/blob/v3.2.3/defaults.go#L108
	output, _ := json.Marshal(v)
	return string(output)
}

func toPrettyJson(v interface{}) string {
	// ref: https://github.com/Masterminds/sprig/blob/v3.2.3/defaults.go#L122
	output, _ := json.MarshalIndent(v, "", "  ")
	return string(output)
}

func indent(spaces int, v string) string {
	// ref: https://github.com/Masterminds/sprig/blob/v3.2.3/strings.go#L109
	pad := strings.Repeat(" ", spaces)
	return pad + strings.Replace(v, "\n", "\n"+pad, -1)
}

func nindent(spaces int, v string) string {
	// ref: https://github.com/Masterminds/sprig/blob/v3.2.3/strings.go#L114
	return "\n" + indent(spaces, v)
}
