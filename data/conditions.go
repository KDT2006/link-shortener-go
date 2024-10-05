package data

type Conditioner interface {
	IsSatisfied(req Request) bool
}

// group predicates
type and struct {
	children []Conditioner
}

func NewAnd(children ...Conditioner) Conditioner {
	return and{children: children}
}

func (a and) IsSatisfied(req Request) bool {
	if len(a.children) == 0 {
		return false
	}
	for _, child := range a.children {
		if !child.IsSatisfied(req) {
			return false
		}
	}

	return true
}

type or struct {
	children []Conditioner
}

func NewOr(children ...Conditioner) Conditioner {
	return or{children: children}
}

func (o or) IsSatisfied(req Request) bool {
	if len(o.children) == 0 {
		return true
	}

	for _, child := range o.children {
		if child.IsSatisfied(req) {
			return true
		}
	}

	return false
}

type not struct {
	child Conditioner
}

func NewNot(child Conditioner) Conditioner {
	return not{child: child}
}

func (n not) IsSatisfied(req Request) bool {
	return !n.child.IsSatisfied(req)
}

// single conditions / attribute-based conditions
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
