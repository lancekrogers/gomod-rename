package vendor

// This file should be skipped by default
import "github.com/old/example/vendored"

func init() {
	vendored.Setup()
}
