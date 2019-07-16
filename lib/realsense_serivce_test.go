package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeDepthThresholds(t *testing.T) {

	actual := MakeDepthThresholds(8, 128)
	assert.Equal(t, []uint32{0, 16, 32, 48, 64, 80, 96, 112}, actual)

	actual = MakeDepthThresholds(16, 128)
	assert.Equal(t, []uint32{0, 8, 16, 24, 32, 40, 48, 56, 64, 72, 80, 88, 96, 104, 112, 120}, actual)
}
