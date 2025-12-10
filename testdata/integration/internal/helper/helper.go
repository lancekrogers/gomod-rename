package helper

import "fmt"

// Format combines a greeting and name into a formatted string.
func Format(greeting, name string) string {
	return fmt.Sprintf("%s, %s!", greeting, name)
}
