package hidden

// This file should be skipped by default
import "github.com/old/example/secret"

func init() {
	secret.Init()
}
