package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPath(t *testing.T) {
	path := NewPath(10)
	assert.NotNil(t, path)
}
