package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnd(t *testing.T) {
	req := NewRequest("1.1.1.1", "US")
	req.Browser = "chrome"

	{
		emptyAnd := NewAnd()
		got := emptyAnd.IsSatisfied(req)
		assert.False(t, got)
	}

	{
		and := NewAnd(NewBrowserIn("safari"))
		got := and.IsSatisfied(req)
		assert.False(t, got)
	}

	{
		and := NewAnd(NewBrowserIn("chrome"))
		got := and.IsSatisfied(req)
		assert.True(t, got)
	}
}

func TestOr(t *testing.T) {
	req := NewRequest("1.1.1.1", "US")
	req.Browser = "chrome"

	{
		emptyOr := NewOr()
		got := emptyOr.IsSatisfied(req)
		assert.True(t, got)
	}

	{
		or := NewOr(NewBrowserIn("safari"), NewBrowserIn("chrome"))
		got := or.IsSatisfied(req)
		assert.True(t, got)
	}

	{
		or := NewAnd(NewBrowserIn("safari"), NewBrowserIn("edge"))
		got := or.IsSatisfied(req)
		assert.False(t, got)
	}
}

func TestNot(t *testing.T) {
	req := NewRequest("1.1.1.1", "US")
	req.Browser = "chrome"

	cond := NewNot(NewBrowserIn("chrome"))
	got := cond.IsSatisfied(req)
	assert.False(t, got)
}
