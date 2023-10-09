package utils

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestPackage(t *testing.T) {
	dir, err := UnpackPackageFromUrl("https://community.chocolatey.org/api/v2/package/ActivePerl/5.24.2.2403")
	assert.Nil(t, err)
	println(dir)
}
