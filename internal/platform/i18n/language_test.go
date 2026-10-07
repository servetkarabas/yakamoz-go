package i18n

import "testing"

func TestResolveFallbackChain(t *testing.T) {
	available := map[string]bool{"tr": true, "en": true}
	tests := []struct {
		name       string
		requested  string
		original   string
		configured string
		want       string
	}{
		{name: "requested", requested: "en", original: "tr", configured: "de", want: "en"},
		{name: "original", requested: "de", original: "tr", configured: "en", want: "tr"},
		{name: "configured", requested: "de", original: "fr", configured: "en", want: "en"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Resolve(test.requested, test.original, test.configured, available); got != test.want {
				t.Fatalf("Resolve() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	if got := Normalize("TR-tr"); got != "tr" {
		t.Fatalf("Normalize() = %q, want tr", got)
	}
}
