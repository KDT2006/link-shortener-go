package data

import (
	"fmt"
	"strings"
)

type inmemLinkStore struct {
	data []Link
}

func NewInmemLinkStore() LinkStorer {
	l1 := NewLink("https://go.dev")
	l1.ID = "1"
	l1.Short = "go"
	l1.DefaultDestination = NewDestination("https://golang.com/", nil, 0)

	l2 := NewLink("https://google.com/")
	l2.Short = "google"
	l2.Destinations = []Destination{
		NewDestination("https://google.com/", NewCountryEquals("US"), 1),
		NewDestination("https://google.co.in/", NewCountryEquals("IN"), 1),
		NewDestination("https://appleNotFound.com/", NewBrowserIn("Safari"), 2),
	}

	l3 := NewLink("https://example.com/")
	l3.Short = "example"
	l3.Destinations = []Destination{
		NewDestination("https://example.com/", NewCountryEquals("US"), 1),
		NewDestination("https://example.in/", NewCountryEquals("IN"), 1),
	}

	return &inmemLinkStore{
		data: []Link{
			l1,
			l2,
			l3,
		},
	}
}

func (store *inmemLinkStore) SaveLink(link Link) error {
	for i, l := range store.data {
		if l.ID == link.ID {
			store.data[i] = link
			return nil
		}
	}

	store.data = append(store.data, link)
	return nil
}

func (store *inmemLinkStore) GetLinkByID(id string) (*Link, error) {
	for _, l := range store.data {
		if l.ID == id {
			return &l, nil
		}
	}

	return nil, fmt.Errorf("link '%s' not found by id", id)
}

func (store *inmemLinkStore) GetLinkByShort(short string) (*Link, error) {
	short = strings.Trim(short, "/")

	for _, l := range store.data {
		if l.Short == short {
			return &l, nil
		}
	}

	return nil, fmt.Errorf("link '%s' not found by short", short)
}
