package i18n

import (
	"net/http"
	"strings"
)

var Supported = []string{"tr", "en", "de", "fr", "es", "ar", "fa", "ur", "ru"}

func Normalize(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if index := strings.IndexByte(value, '-'); index >= 0 {
		value = value[:index]
	}
	return value
}

func IsSupported(value string) bool {
	value = Normalize(value)
	for _, supported := range Supported {
		if supported == value {
			return true
		}
	}
	return false
}

func Requested(r *http.Request, fallback string) string {
	if value := r.URL.Query().Get("lang"); value != "" {
		return Normalize(value)
	}
	if value := r.Header.Get("Accept-Language"); value != "" {
		if comma := strings.IndexByte(value, ','); comma >= 0 {
			value = value[:comma]
		}
		value = strings.TrimSpace(strings.SplitN(value, ";", 2)[0])
		if value != "" {
			return Normalize(value)
		}
	}
	return Normalize(fallback)
}

func Resolve(requested, original, configured string, available map[string]bool) string {
	for _, candidate := range []string{requested, original, configured} {
		candidate = Normalize(candidate)
		if candidate != "" && available[candidate] {
			return candidate
		}
	}
	for language := range available {
		return language
	}
	return ""
}
