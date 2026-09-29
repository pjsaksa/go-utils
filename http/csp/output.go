package csp

import (
	"bytes"
	"fmt"
	"io"
)

var emptyPolicy = Policy{}

func Output(base *Policy, page *Policy) string {
	if base == nil {
		base = &emptyPolicy
	}
	if page == nil {
		page = &emptyPolicy
	}

	out := bytes.Buffer{}
	writeSep := false
	for i := range DirectiveCount {
		v := page[i]
		if v == nil {
			v = base[i]
		}
		if v != nil {
			if writeSep {
				io.WriteString(&out, "; ")
			} else {
				writeSep = true
			}
			fmt.Fprintf(&out, "%s %s", i, v.cspValue())
		}
	}
	return out.String()
}
