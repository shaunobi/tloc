package buildinfo

import "testing"

func TestNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "release tag", input: "v1.2.3", want: "1.2.3"},
		{name: "release version", input: "1.2.3-rc.1", want: "1.2.3-rc.1"},
		{name: "whitespace", input: " v1.2.3 ", want: "1.2.3"},
		{name: "development", input: "(devel)", want: ""},
		{name: "empty", input: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := normalize(test.input); got != test.want {
				t.Fatalf("normalize(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestVersionPrefersLinkerValue(t *testing.T) {
	previous := version
	version = "v9.8.7"
	t.Cleanup(func() { version = previous })

	if got := Version(); got != "9.8.7" {
		t.Fatalf("Version() = %q, want %q", got, "9.8.7")
	}
}
