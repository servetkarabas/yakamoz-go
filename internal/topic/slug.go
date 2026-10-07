package topic

import (
	"strings"
	"unicode"
)

var transliterations = strings.NewReplacer(
	"ı", "i", "İ", "I", "ğ", "g", "Ğ", "G", "ü", "u", "Ü", "U",
	"ş", "s", "Ş", "S", "ö", "o", "Ö", "O", "ç", "c", "Ç", "C",
	"ä", "a", "Ä", "A", "ë", "e", "Ë", "E", "ï", "i", "Ï", "I",
	"ñ", "n", "Ñ", "N", "ß", "ss",
	"é", "e", "É", "E", "á", "a", "Á", "A", "í", "i", "Í", "I",
	"ó", "o", "Ó", "O", "ú", "u", "Ú", "U",
)

func Slug(title string) string {
	title = transliterations.Replace(strings.ToLower(strings.TrimSpace(title)))
	var b strings.Builder
	separator := false
	for _, char := range title {
		if unicode.Is(unicode.Mn, char) {
			continue
		}
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			b.WriteRune(char)
			separator = false
			continue
		}
		if b.Len() > 0 {
			separator = true
		}
		if separator && (b.Len() == 0 || !strings.HasSuffix(b.String(), "-")) {
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
