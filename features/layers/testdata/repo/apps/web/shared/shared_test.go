package shared_test

import (
	"testing"

	api "example.com/repo/api/shared" // want `an import into another app's source is forbidden`
)

func TestName(t *testing.T) {
	if api.Name == "" {
		t.Fatal("no name")
	}
}
