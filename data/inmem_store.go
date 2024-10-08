package data

import (
	"fmt"
	"strings"
)

type inmemLinkStore struct {
	data []Link
}

func NewInmemLinkStore() LinkStorer {
	return &inmemLinkStore{}
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
			if l.Active {
				return &l, nil
			}
		}
	}

	return nil, fmt.Errorf("link '%s' not found by short", short)
}

func (store *inmemLinkStore) ToggleLinkByShort(short string) (*Link, error) {
	short = strings.Trim(short, "/")

	for i, l := range store.data {
		if l.Short == short {
			store.data[i].Active = !store.data[i].Active
			return &store.data[i], nil
		}
	}

	return nil, fmt.Errorf("link '%s' not found by short", short)
}
