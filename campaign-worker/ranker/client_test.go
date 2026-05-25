package ranker_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilovecroissant/ai-marketing-engine/campaign-worker/ranker"
)

// newMock spins up an in-process HTTP server that returns `body` for any request.
func newMock(t *testing.T, body any) (*httptest.Server, *ranker.Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv, ranker.New(srv.URL)
}

// ── ExtractKeywords ──────────────────────────────────────────────────────────

func TestExtractKeywords_OK(t *testing.T) {
	_, client := newMock(t, map[string]any{
		"keywords": []map[string]any{
			{"term": "marketing", "score": 0.42},
			{"term": "digital", "score": 0.31},
		},
	})

	kws, err := client.ExtractKeywords("digital marketing content", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kws) != 2 {
		t.Fatalf("expected 2 keywords, got %d", len(kws))
	}
	if kws[0].Term != "marketing" {
		t.Errorf("expected first keyword 'marketing', got %q", kws[0].Term)
	}
	if kws[0].Score != 0.42 {
		t.Errorf("expected score 0.42, got %f", kws[0].Score)
	}
}

func TestExtractKeywords_ServerDown(t *testing.T) {
	client := ranker.New("http://localhost:19999") // nothing listening here
	_, err := client.ExtractKeywords("some text", 5)
	if err == nil {
		t.Fatal("expected error when server is down, got nil")
	}
}

// ── RankContent ──────────────────────────────────────────────────────────────

func TestRankContent_OK(t *testing.T) {
	_, client := newMock(t, map[string]any{
		"seo":         0.72,
		"engagement":  0.65,
		"readability": 0.80,
		"total":       0.718,
	})

	result, err := client.RankContent("some article text", []string{"marketing", "seo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SEO != 0.72 {
		t.Errorf("expected SEO=0.72, got %f", result.SEO)
	}
	if result.Total != 0.718 {
		t.Errorf("expected Total=0.718, got %f", result.Total)
	}
}

func TestRankContent_AllScoresInUnitInterval(t *testing.T) {
	_, client := newMock(t, map[string]any{
		"seo": 0.5, "engagement": 0.5, "readability": 0.5, "total": 0.5,
	})

	result, err := client.RankContent("text", []string{"kw"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for name, v := range map[string]float64{
		"SEO": result.SEO, "Engagement": result.Engagement,
		"Readability": result.Readability, "Total": result.Total,
	} {
		if v < 0 || v > 1 {
			t.Errorf("%s=%f is outside [0,1]", name, v)
		}
	}
}

// ── CheckSimilarity ──────────────────────────────────────────────────────────

func TestCheckSimilarity_Duplicate(t *testing.T) {
	_, client := newMock(t, map[string]any{
		"similarity":   0.91,
		"is_duplicate": true,
	})

	result, err := client.CheckSimilarity("text a", "text b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsDuplicate {
		t.Error("expected IsDuplicate=true")
	}
	if result.Similarity != 0.91 {
		t.Errorf("expected Similarity=0.91, got %f", result.Similarity)
	}
}

func TestCheckSimilarity_NotDuplicate(t *testing.T) {
	_, client := newMock(t, map[string]any{
		"similarity":   0.12,
		"is_duplicate": false,
	})

	result, err := client.CheckSimilarity("apples", "quantum physics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsDuplicate {
		t.Error("expected IsDuplicate=false")
	}
}

func TestCheckSimilarity_ServerDown(t *testing.T) {
	client := ranker.New("http://localhost:19999")
	_, err := client.CheckSimilarity("a", "b")
	if err == nil {
		t.Fatal("expected error when server is down, got nil")
	}
}

// ── ranker error does not panic ───────────────────────────────────────────────

func TestRankContent_BadJSON_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json {{{")) // malformed response
	}))
	defer srv.Close()

	client := ranker.New(srv.URL)
	_, err := client.RankContent("text", []string{"kw"})
	if err == nil {
		t.Fatal("expected decode error on malformed JSON, got nil")
	}
}
