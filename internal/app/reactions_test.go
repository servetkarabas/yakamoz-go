package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/karabas/yakamoz/internal/config"
	"github.com/karabas/yakamoz/internal/uuid"
)

func TestAnonymousReactionHTTPSetAndToggle(t *testing.T) {
	cfg := config.Config{DefaultLanguage: "tr", AITimeout: time.Second}
	server := httptest.NewServer(NewMemory(cfg, slog.New(slog.NewTextHandler(io.Discard, nil))))
	defer server.Close()
	topicID, visitorID := uuid.New(), uuid.New()

	vote := func(reaction string) reactionSummary {
		body, err := json.Marshal(map[string]string{
			"target": "topic", "target_id": topicID.String(), "visitor_id": visitorID.String(), "reaction": reaction,
		})
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequest(http.MethodPut, server.URL+"/api/v1/reactions", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("reaction status = %d", response.StatusCode)
		}
		var summary reactionSummary
		if err := json.NewDecoder(response.Body).Decode(&summary); err != nil {
			t.Fatal(err)
		}
		return summary
	}

	if got := vote("like"); got.Likes != 1 || got.UserReaction != reactionLike {
		t.Fatalf("like response = %#v", got)
	}
	if got := vote("dislike"); got.Likes != 0 || got.Dislikes != 1 || got.UserReaction != reactionDislike {
		t.Fatalf("dislike response = %#v", got)
	}
	if got := vote(""); got.Dislikes != 0 || got.UserReaction != "" {
		t.Fatalf("cleared response = %#v", got)
	}
}

func TestMemoryReactionStoreTogglesVotesByVisitorAndTarget(t *testing.T) {
	store := newMemoryReactionStore()
	itemID := uuid.New()
	visitorID := uuid.New()
	otherVisitorID := uuid.New()

	if err := store.Set(context.Background(), reactionTopic, itemID, visitorID, reactionLike); err != nil {
		t.Fatal(err)
	}
	if err := store.Set(context.Background(), reactionTopic, itemID, otherVisitorID, reactionLike); err != nil {
		t.Fatal(err)
	}
	if err := store.Set(context.Background(), reactionComment, itemID, visitorID, reactionDislike); err != nil {
		t.Fatal(err)
	}

	values, err := store.List(context.Background(), reactionTopic, []uuid.UUID{itemID}, visitorID)
	if err != nil {
		t.Fatal(err)
	}
	if got := values[itemID]; got.Likes != 2 || got.Dislikes != 0 || got.UserReaction != reactionLike {
		t.Fatalf("topic reactions = %#v", got)
	}

	if err := store.Set(context.Background(), reactionTopic, itemID, visitorID, reactionDislike); err != nil {
		t.Fatal(err)
	}
	values, err = store.List(context.Background(), reactionTopic, []uuid.UUID{itemID}, visitorID)
	if err != nil {
		t.Fatal(err)
	}
	if got := values[itemID]; got.Likes != 1 || got.Dislikes != 1 || got.UserReaction != reactionDislike {
		t.Fatalf("updated topic reactions = %#v", got)
	}

	if err := store.Set(context.Background(), reactionTopic, itemID, visitorID, ""); err != nil {
		t.Fatal(err)
	}
	values, err = store.List(context.Background(), reactionComment, []uuid.UUID{itemID}, visitorID)
	if err != nil {
		t.Fatal(err)
	}
	if got := values[itemID]; got.Likes != 0 || got.Dislikes != 1 || got.UserReaction != reactionDislike {
		t.Fatalf("comment reactions = %#v", got)
	}
}
