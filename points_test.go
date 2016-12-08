package main

import (
	"testing"

	"github.com/nordsieck/defect"
)

func TestClean(t *testing.T) {
	cases := map[string]string{
		"foo bar (123)":     "foo bar",
		"foo bar baz (543)": "foo bar baz",
	}

	for in, expected := range cases {
		defect.Equal(t, Clean(in), expected)
	}
}

func TestShrink(t *testing.T) {
	defect.DeepEqual(t, Shrink([]Small{
		{Name: "foo bar (123)", Wscid: 123},
		{Name: "foo bar baz (543)", Wscid: 543},
	}), []Small{
		{Name: "foo bar", Wscid: 123},
		{Name: "foo bar baz", Wscid: 543},
	})
}
