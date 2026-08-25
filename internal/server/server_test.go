package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/bShaak/habitui/internal/api"
	"github.com/bShaak/habitui/internal/models"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	svc, err := api.Open(dbPath)
	if err != nil {
		t.Fatalf("api.Open: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	ts := httptest.NewServer(New(svc))
	t.Cleanup(ts.Close)
	return ts
}

func doJSON(t *testing.T, method, url string, body any) (int, []byte) {
	t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, rd)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read response: %v", err)
	}
	return resp.StatusCode, buf.Bytes()
}

func TestHabitLifecycleOverHTTP(t *testing.T) {
	ts := newTestServer(t)

	// Create
	status, body := doJSON(t, "POST", ts.URL+"/api/v1/habits", models.Habit{
		Name: "Read", Goal: 2, Frequency: "daily",
	})
	if status != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", status, body)
	}
	var habit models.Habit
	if err := json.Unmarshal(body, &habit); err != nil {
		t.Fatalf("decode create response: %v\n%s", err, body)
	}
	if habit.ID == 0 || habit.Name != "Read" {
		t.Fatalf("created habit = %+v", habit)
	}

	// Get
	status, body = doJSON(t, "GET", fmt.Sprintf("%s/api/v1/habits/%d", ts.URL, habit.ID), nil)
	if status != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", status, body)
	}
	var got models.Habit
	_ = json.Unmarshal(body, &got)
	if got.Name != "Read" {
		t.Errorf("got %+v, want name Read", got)
	}

	// Update
	got.Description = "every evening"
	status, body = doJSON(t, "PUT", fmt.Sprintf("%s/api/v1/habits/%d", ts.URL, habit.ID), got)
	if status != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", status, body)
	}

	// List (day summary)
	status, body = doJSON(t, "GET", ts.URL+"/api/v1/habits", nil)
	if status != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", status, body)
	}
	var sum api.DaySummary
	if err := json.Unmarshal(body, &sum); err != nil {
		t.Fatalf("decode list: %v\n%s", err, body)
	}
	if len(sum.Habits) != 1 || sum.Habits[0].Name != "Read" || !sum.Habits[0].Due {
		t.Fatalf("summary = %+v", sum)
	}

	// Delete
	status, _ = doJSON(t, "DELETE", fmt.Sprintf("%s/api/v1/habits/%d", ts.URL, habit.ID), nil)
	if status != http.StatusNoContent {
		t.Fatalf("delete status = %d", status)
	}
	status, _ = doJSON(t, "GET", fmt.Sprintf("%s/api/v1/habits/%d", ts.URL, habit.ID), nil)
	if status != http.StatusNotFound {
		t.Errorf("get after delete status = %d, want 404", status)
	}
}

func TestCompleteAndToggleOverHTTP(t *testing.T) {
	ts := newTestServer(t)

	status, body := doJSON(t, "POST", ts.URL+"/api/v1/habits", models.Habit{Name: "Run", Goal: 1})
	if status != http.StatusCreated {
		t.Fatalf("create status = %d", status)
	}
	var habit models.Habit
	_ = json.Unmarshal(body, &habit)

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Complete on a specific date.
	url := fmt.Sprintf("%s/api/v1/habits/%d/complete?date=%s", ts.URL, habit.ID, yesterday)
	status, body = doJSON(t, "POST", url, nil)
	if status != http.StatusOK {
		t.Fatalf("complete status = %d, body = %s", status, body)
	}
	var res api.CompleteResult
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("decode complete: %v\n%s", err, body)
	}
	if res.AlreadyComplete || res.CompletionCount != 1 {
		t.Errorf("result = %+v; want first completion", res)
	}

	// Toggle adds then removes.
	toggleURL := fmt.Sprintf("%s/api/v1/habits/%d/toggle?date=%s", ts.URL, habit.ID, yesterday)
	status, body = doJSON(t, "POST", toggleURL, nil)
	if status != http.StatusOK {
		t.Fatalf("toggle status = %d, body = %s", status, body)
	}
	var tog api.ToggleResult
	_ = json.Unmarshal(body, &tog)
	if len(tog.RemovedIDs) != 1 || tog.CompletionCount != 0 {
		t.Errorf("toggle result = %+v; want removal of one completion", tog)
	}

	// Completions filtered by range.
	status, body = doJSON(t, "GET", ts.URL+"/api/v1/completions?from=2000-01-01&to="+time.Now().Format("2006-01-02"), nil)
	if status != http.StatusOK {
		t.Fatalf("completions status = %d", status)
	}
	var comps []models.Completion
	_ = json.Unmarshal(body, &comps)
	if len(comps) != 0 {
		t.Errorf("expected no completions after untoggle, got %d", len(comps))
	}
}

func TestErrorHandlingOverHTTP(t *testing.T) {
	ts := newTestServer(t)

	cases := []struct {
		name   string
		method string
		url    string
		body   any
		want   int
	}{
		{"missing habit", "GET", ts.URL + "/api/v1/habits/999", nil, http.StatusNotFound},
		{"bad id", "GET", ts.URL + "/api/v1/habits/abc", nil, http.StatusBadRequest},
		{"bad date", "GET", ts.URL + "/api/v1/habits?date=nope", nil, http.StatusBadRequest},
		{"bad json", "POST", ts.URL + "/api/v1/habits", "not-an-object}", http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, _ := doJSON(t, tc.method, tc.url, tc.body)
			if status != tc.want {
				t.Errorf("status = %d, want %d", status, tc.want)
			}
		})
	}
}
