package app

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/karabas/yakamoz/internal/config"
)

func TestHTTPAuthorTopicLanguageFlow(t *testing.T) {
	cfg := config.Config{DefaultLanguage: "tr", AITimeout: 2 * time.Second}
	server := httptest.NewServer(NewMemory(cfg, slog.New(slog.NewTextHandler(io.Discard, nil))))
	defer server.Close()
	authorBody := `{"nickname":"ada","email":"ada@example.com","preferred_language":"tr"}`
	response, err := http.Post(server.URL+"/api/v1/authors", "application/json", bytes.NewBufferString(authorBody))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("author status = %d", response.StatusCode)
	}
	var author struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(response.Body).Decode(&author); err != nil {
		t.Fatal(err)
	}
	topicBody := []byte(`{"title":"İstanbul","description":"şehir","language":"tr","author_id":"` + author.ID + `"}`)
	response, err = http.Post(server.URL+"/api/v1/topics", "application/json", bytes.NewReader(topicBody))
	if err != nil {
		t.Fatal(err)
	}
	var topic struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(response.Body).Decode(&topic); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if topic.Slug != "istanbul" {
		t.Fatalf("slug = %q", topic.Slug)
	}
	request, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/topics/"+topic.ID+"?lang=en", nil)
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var fallback struct {
		ServedLanguage string `json:"served_language"`
		Title          string `json:"title"`
	}
	if err := json.NewDecoder(response.Body).Decode(&fallback); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if fallback.ServedLanguage != "tr" || fallback.Title != "İstanbul" {
		t.Fatalf("fallback response = %#v", fallback)
	}
}
