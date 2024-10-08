package data

import (
	"fmt"
	"slices"
	"time"

	"github.com/jaevor/go-nanoid"
)

var newId, err = nanoid.Standard(25)
var newShort, _ = nanoid.CustomASCII("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", 8)

func NewLink(full string) Link {
	return Link{
		ID:                 newId(),
		Short:              newShort(),
		Destinations:       []Destination{},
		DefaultDestination: Destination{},
		Active:             true,
	}
}

type Link struct {
	ID                 string
	Short              string
	Destinations       []Destination
	DefaultDestination Destination
	Active             bool
}

func (l *Link) SortDestinations() {
	slices.SortFunc(l.Destinations, func(i, j Destination) int {
		return j.Priority - i.Priority
	})
}

func (l *Link) PickDestination(req Request) (*Destination, error) {
	l.SortDestinations()
	for _, d := range l.Destinations {
		if d.Condition != nil {
			if d.Condition.IsSatisfied(req) {
				return &d, nil
			}
		}
	}

	if l.DefaultDestination.URL != "" {
		return &l.DefaultDestination, nil
	}

	return nil, fmt.Errorf("no destination found")
}

type Destination struct {
	ID        string
	URL       string
	Priority  int
	Condition Conditioner
}

type Request struct {
	At          time.Time
	IP          string
	Browser     string
	CountryISO2 string
}

func NewDestination(url string, cond Conditioner, priority int) Destination {
	return Destination{
		ID:        newId(),
		URL:       url,
		Condition: cond,
		Priority:  priority,
	}
}

func NewRequest(ip, country string) Request {
	return Request{
		At:          time.Now(),
		IP:          ip,
		CountryISO2: country,
	}
}

type LinkStorer interface {
	SaveLink(link Link) error
	GetLinkByID(id string) (*Link, error)
	GetLinkByShort(short string) (*Link, error)
	// TODO(maybe) UpdateLinkByShort(short string) (*Link, error)
	ToggleLinkByShort(short string) (*Link, error)
}
