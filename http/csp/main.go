package csp

import (
	"strings"
)

type Policy [DirectiveCount]Value

type Directive int

type Value interface {
	cspValue() string
}

//
// Array of values
//

type Array []Value

func (a Array) cspValue() string {
	if n := len(a); n > 0 {
		s := make([]string, n)
		for i := range a {
			s[i] = a[i].cspValue()
		}
		return strings.Join(s, " ")
	} else {
		return ""
	}
}

//
// Const value
//

type constValue int

//
// String value
//

type String string

func (s String) cspValue() string {
	return string(s)
}
