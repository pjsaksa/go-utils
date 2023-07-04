package http

import (
	"testing"
)

func Test_SplitUrlPath(t *testing.T) {
	var data = []struct {
		input string
		want  []string
	}{
		{
			input: "",
			want:  nil,
		}, {
			input: "meh",
			want:  nil,
		}, {
			input: " /meh",
			want:  nil, // First character must be '/'
		}, {
			input: "/",
			want:  []string{""},
		}, {
			input: "//",
			want:  []string{"", ""},
		}, {
			input: "///",
			want:  []string{"", "", ""},
		}, {
			input: "/default.css",
			want:  []string{"default.css"},
		}, {
			input: "/u/",
			want:  []string{"u", ""},
		}, {
			input: "/sign-in",
			want:  []string{"sign-in"},
		}, {
			input: "/u/sign-out",
			want:  []string{"u", "sign-out"},
		}, {
			input: "/api/projects",
			want:  []string{"api", "projects"},
		},
	}

	for i := range data {
		if output := splitUrlPath(data[i].input); !compareUrlParts(output, data[i].want) {
			t.Errorf("FAIL: \"%s\"\n-> %q\n!= %q",
				data[i].input,
				output,
				data[i].want)
		}
	}
}

// ------------------------------------------------------------

func compareUrlParts(a []string, b []string) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
