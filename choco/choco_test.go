package choco

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestChoco(t *testing.T) {
	md, err := GetPkgMetadata("activeperl", "5.24.2.2403")
	assert.Nil(t, err)
	fmt.Printf("%+v\n", md)
}
