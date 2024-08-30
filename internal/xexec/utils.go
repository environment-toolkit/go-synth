package xexec

import "fmt"

// Format environment variables for command execution.
func formatEnvVars(envVars map[string]string) []string {
	formatted := make([]string, 0, len(envVars))
	for k, v := range envVars {
		formatted = append(formatted, fmt.Sprintf("%s=%s", k, v))
	}
	return formatted
}
