package main

import (
	"link-shortner/data"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleRedirection(t *testing.T) {
	link1 := data.NewLink("https://google.com/")
	assert.Equal(t, "https://google.com/", link1.FullURL)
}
