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

func TestMemorySeedHasUniquePrivilegedRolesAndManyAuthors(t *testing.T) {
	cfg := config.Config{DefaultLanguage: "tr", AITimeout: 2 * time.Second}
	server := httptest.NewServer(NewMemory(cfg, slog.New(slog.NewTextHandler(io.Discard, nil))))
	defer server.Close()

	response, err := http.Get(server.URL + "/api/v1/authors?limit=100")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var users []struct {
		Nickname string `json:"nickname"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(response.Body).Decode(&users); err != nil {
		t.Fatal(err)
	}
	roles := make(map[string]int)
	currentAuthors := map[string]bool{"yakamoz": false, "geceyazari": false, "marti": false}
	for _, user := range users {
		roles[user.Role]++
		if _, exists := currentAuthors[user.Nickname]; exists {
			currentAuthors[user.Nickname] = user.Role == "author"
		}
	}
	if roles["admin"] != 1 || roles["reviewer"] != 1 || roles["author"] < 10 {
		t.Fatalf("seed role counts = %#v", roles)
	}
	for nickname, isAuthor := range currentAuthors {
		if !isAuthor {
			t.Errorf("existing user %q should remain an author", nickname)
		}
	}
}

func TestMemorySeedContainsMultilingualTopicsAndComments(t *testing.T) {
	cfg := config.Config{DefaultLanguage: "tr", AITimeout: 2 * time.Second}
	server := httptest.NewServer(NewMemory(cfg, slog.New(slog.NewTextHandler(io.Discard, nil))))
	defer server.Close()

	response, err := http.Get(server.URL + "/api/v1/topics?lang=de&status=published&limit=100")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var topics []struct {
		ID             string `json:"id"`
		Slug           string `json:"slug"`
		Title          string `json:"title"`
		ServedLanguage string `json:"served_language"`
	}
	if err := json.NewDecoder(response.Body).Decode(&topics); err != nil {
		t.Fatal(err)
	}
	var yakamoz struct {
		ID    string
		Title string
		Lang  string
	}
	for _, value := range topics {
		if value.Slug == "yakamoz" {
			yakamoz.ID, yakamoz.Title, yakamoz.Lang = value.ID, value.Title, value.ServedLanguage
			break
		}
	}
	if yakamoz.ID == "" || yakamoz.Title != "Mondschein auf dem Meer" || yakamoz.Lang != "de" {
		t.Fatalf("German seed topic = %#v", yakamoz)
	}

	response, err = http.Get(server.URL + "/api/v1/comments?topic_id=" + yakamoz.ID + "&limit=100")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var comments []struct {
		Language string `json:"language"`
	}
	if err := json.NewDecoder(response.Body).Decode(&comments); err != nil {
		t.Fatal(err)
	}
	languages := make(map[string]bool)
	for _, value := range comments {
		languages[value.Language] = true
	}
	for _, language := range []string{"tr", "en", "de", "fr", "es"} {
		if !languages[language] {
			t.Errorf("seed comments missing language %q: %#v", language, comments)
		}
	}

	response, err = http.Get(server.URL + "/api/v1/comments/counts?topic_ids=" + yakamoz.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var counts map[string]int
	if err := json.NewDecoder(response.Body).Decode(&counts); err != nil {
		t.Fatal(err)
	}
	if counts[yakamoz.ID] != len(comments) {
		t.Fatalf("comment count = %d, want %d", counts[yakamoz.ID], len(comments))
	}
}

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
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	if err := json.NewDecoder(response.Body).Decode(&author); err != nil {
		t.Fatal(err)
	}
	if author.Role != "author" {
		t.Fatalf("role for omitted role = %q, want author", author.Role)
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
	var translated struct {
		ServedLanguage string `json:"served_language"`
		Title          string `json:"title"`
	}
	if err := json.NewDecoder(response.Body).Decode(&translated); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if translated.ServedLanguage != "en" || translated.Title != "[en] İstanbul" {
		t.Fatalf("translated response = %#v", translated)
	}
}
