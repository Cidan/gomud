package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	gmap := NewMap(10)
	assert.NotNil(t, gmap)
}
