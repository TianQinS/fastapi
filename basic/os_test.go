package basic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOs(t *testing.T) {
	out, err := Exec("dir")
	assert.Equal(t, nil, err)
	assert.Equal(t, true, len(out) > 0)
}
