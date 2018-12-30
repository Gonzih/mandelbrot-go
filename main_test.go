package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslate(t *testing.T) {
	res := tranlate(10, 0, 20, 0, 400)
	assert.Equal(t, res, float64(200))
}
