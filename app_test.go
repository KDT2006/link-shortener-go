package main

import (
	"link-shortner/data"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV2(t *testing.T) {
	store := data.NewInmemLinkStore()

	// we hit the short url
	short := "go"

	// our app gets the link by short
	link, err := store.GetLinkByShort(short)
	assert.NoError(t, err)
	assert.Equal(t, "1", link.ID)

	// our app builds a request object
	req := data.NewRequest("1.1.1.1", "US")
	req.Browser = "chrome"

	// our app gets the destination
	dest, err := link.PickDestination(req)
	assert.NoError(t, err)
	assert.Equal(t, "https://golang.com/", dest.URL)
}
