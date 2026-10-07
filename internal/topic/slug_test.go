package topic

import "testing"

func TestSlugTurkishUnicode(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{title: "İstanbul", want: "istanbul"},
		{title: "ğüşiöç", want: "gusioc"},
		{title: "Café & Kültür", want: "cafe-kultur"},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			if got := Slug(test.title); got != test.want {
				t.Fatalf("Slug(%q) = %q, want %q", test.title, got, test.want)
			}
		})
	}
}
