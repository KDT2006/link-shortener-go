package data

import "github.com/jaevor/go-nanoid"

var newId, err = nanoid.Standard(25)
var newShort, _ = nanoid.CustomASCII("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", 8)

func NewLink(full string) Link {
	return Link{
		ID:      newId(),
		Short:   newShort(),
		FullURL: full,
		Active:  true,
	}
}

type Link struct {
	ID      string
	Short   string
	FullURL string
	Active  bool
}

type LinkStorer interface {
	SaveLink(link Link) error
	GetLinkByID(id string) (*Link, error)
	GetLinkByShort(short string) (*Link, error)
}
