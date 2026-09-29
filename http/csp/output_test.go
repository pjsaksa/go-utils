package csp

import (
	"testing"
)

func TestWrite(t *testing.T) {
	for _, u := range []struct {
		base   Policy
		page   Policy
		expect string
	}{
		{
			// empty
		},
		{
			base: Policy{
				DefaultSrc: None,
			},
			expect: "default-src 'none'",
		},
		{
			base: Policy{
				DefaultSrc: None,
				StyleSrc:   String("foo"),
			},
			page: Policy{
				DefaultSrc: String("bar"),
				ScriptSrc:  Self,
			},
			expect: "default-src bar; script-src 'self'; style-src foo",
		},
	} {
		content, err := Output(u.base, u.page)
		if err != nil {
			t.Errorf("error: %s", err)
		} else if u.expect != content {
			t.Errorf("\n%q\n ..should be.. \n%q", content, u.expect)
		}
	}
}
