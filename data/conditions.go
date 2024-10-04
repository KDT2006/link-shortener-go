package data

type Conditioner interface {
	IsSatisfied(req Request) bool
}

func NewCountryEquals(value string) Conditioner {
	return countryEquals{value: value}
}

type countryEquals struct {
	value string
}

type browserIn struct {
	values []string
}

func (c countryEquals) IsSatisfied(req Request) bool {
	return req.CountryISO2 == c.value
}

func NewBrowserIn(values ...string) browserIn {
	return browserIn{
		values: values,
	}
}

func (c browserIn) IsSatisfied(req Request) bool {
	for _, value := range c.values {
		if req.Browser == value {
			return true
		}
	}

	return false
}
