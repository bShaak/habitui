// Package server exposes an api.Service over HTTP so native front-end clients
// can control habits. All endpoints speak JSON and live under /api/v1.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bShaak/habitui/internal/api"
	"github.com/bShaak/habitui/internal/models"
)

const dateFormat = "2006-01-02"

// Server wraps an api.Service with HTTP handlers.
type Server struct {
	svc api.Service
}

// New returns an http.Handler serving the API backed by svc.
func New(svc api.Service) http.Handler {
	s := &Server{svc: svc}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/habits", s.listHabits)
	mux.HandleFunc("POST /api/v1/habits", s.createHabit)
	mux.HandleFunc("GET /api/v1/habits/{id}", s.getHabit)
	mux.HandleFunc("PUT /api/v1/habits/{id}", s.updateHabit)
	mux.HandleFunc("DELETE /api/v1/habits/{id}", s.deleteHabit)
	mux.HandleFunc("POST /api/v1/habits/{id}/complete", s.completeHabit)
	mux.HandleFunc("POST /api/v1/habits/{id}/toggle", s.toggleCompletion)
	mux.HandleFunc("GET /api/v1/completions", s.listCompletions)

	return logRequests(mux)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func (s *Server) listHabits(w http.ResponseWriter, r *http.Request) {
	date, err := queryDate(r, "date", time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	sum, err := s.svc.DaySummary(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

func (s *Server) getHabit(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	h, err := s.svc.GetHabit(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h)
}

func (s *Server) createHabit(w http.ResponseWriter, r *http.Request) {
	var h models.Habit
	if !decodeBody(w, r, &h) {
		return
	}
	created, err := s.svc.CreateHabit(r.Context(), &h)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) updateHabit(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var h models.Habit
	if !decodeBody(w, r, &h) {
		return
	}
	h.ID = id
	if err := s.svc.UpdateHabit(r.Context(), &h); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h)
}

func (s *Server) deleteHabit(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := s.svc.DeleteHabit(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) completeHabit(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	date, err := queryDate(r, "date", time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.svc.CompleteHabit(r.Context(), id, date)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) toggleCompletion(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	date, err := queryDate(r, "date", time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.svc.ToggleCompletionForDate(r.Context(), id, date)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) listCompletions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ctx := r.Context()

	switch {
	case q.Get("habit_id") != "":
		id, err := strconv.ParseInt(q.Get("habit_id"), 10, 64)
		if err != nil || id <= 0 {
			writeError(w, http.StatusBadRequest, "invalid habit_id")
			return
		}
		comps, err := completionsForHabit(ctx, s.svc, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, comps)
	case q.Get("from") != "" || q.Get("to") != "":
		from, to, ok := parseRange(w, q.Get("from"), q.Get("to"))
		if !ok {
			return
		}
		comps, err := s.svc.CompletionsByDateRange(ctx, from, to)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, comps)
	default:
		comps, err := s.svc.ListCompletions(ctx)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, comps)
	}
}

func completionsForHabit(ctx context.Context, svc api.Service, id int64) ([]models.Completion, error) {
	all, err := svc.ListCompletions(ctx)
	if err != nil {
		return nil, err
	}
	var out []models.Completion
	for _, c := range all {
		if c.HabitID == id {
			out = append(out, c)
		}
	}
	return out, nil
}

func parseRange(w http.ResponseWriter, fromStr, toStr string) (time.Time, time.Time, bool) {
	badRequest := func(msg string) (time.Time, time.Time, bool) {
		writeError(w, http.StatusBadRequest, msg)
		return time.Time{}, time.Time{}, false
	}
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	to := from
	if fromStr != "" {
		t, err := time.ParseInLocation(dateFormat, fromStr, time.Local)
		if err != nil {
			return badRequest(fmt.Sprintf("invalid from %q (use %s)", fromStr, dateFormat))
		}
		from = t
	}
	if toStr != "" {
		t, err := time.ParseInLocation(dateFormat, toStr, time.Local)
		if err != nil {
			return badRequest(fmt.Sprintf("invalid to %q (use %s)", toStr, dateFormat))
		}
		to = t
	}
	if to.Before(from) {
		return badRequest("to is before from")
	}
	return from, to, true
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return 0, false
	}
	return id, true
}

func queryDate(r *http.Request, key string, def time.Time) (time.Time, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def, nil
	}
	return time.ParseInLocation(dateFormat, raw, time.Local)
}

func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
