package greeter

import "github.com/testold/myproject/internal/helper"

// Greet returns a greeting for the given name.
func Greet(name string) string {
	return helper.Format("Hello", name)
}
