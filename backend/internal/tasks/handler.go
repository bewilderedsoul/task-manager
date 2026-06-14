package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"taskmanager/internal/events"
	"taskmanager/internal/httpx"
	"taskmanager/internal/middleware"
	"taskmanager/internal/models"
	"taskmanager/internal/validate"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

type Handler struct {
	repo *Repository
	hub  *events.Hub
}

func NewHandler(repo *Repository, hub *events.Hub) *Handler {
	return &Handler{repo: repo, hub: hub}
}

// ── Create ──────────────────────────────────────────────────────────────────

type createRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=200"`
	Description string  `json:"description" validate:"max=5000"`
	Status      string  `json:"status" validate:"omitempty,oneof=todo in_progress done"`
	Priority    string  `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate     *string `json:"dueDate"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	var in createRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	if errs := validate.Struct(in); errs != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation failed", errs)
		return
	}

	due, err := parseOptionalTime(in.DueDate)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation failed", map[string]string{"dueDate": "must be an RFC3339 date-time"})
		return
	}

	params := CreateParams{
		UserID:      user.ID,
		Title:       in.Title,
		Description: in.Description,
		Status:      defaultIfEmpty(in.Status, "todo"),
		Priority:    defaultIfEmpty(in.Priority, "medium"),
		DueDate:     due,
	}

	task, err := h.repo.Create(r.Context(), params)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not create task")
		return
	}

	h.hub.Publish(user.ID, events.Event{Type: "task.created", Data: task})
	httpx.JSON(w, http.StatusCreated, map[string]any{"task": task})
}

// ── List ────────────────────────────────────────────────────────────────────

var allowedSort = map[string]bool{"created_at": true, "due_date": true, "priority": true}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	q := r.URL.Query()

	status := q.Get("status")
	if status != "" && status != "todo" && status != "in_progress" && status != "done" {
		httpx.Error(w, http.StatusUnprocessableEntity, "invalid status filter")
		return
	}

	sortBy := q.Get("sort")
	if sortBy == "" {
		sortBy = "created_at"
	}
	if !allowedSort[sortBy] {
		httpx.Error(w, http.StatusUnprocessableEntity, "invalid sort field")
		return
	}

	order := strings.ToLower(q.Get("order"))
	if order == "" {
		order = "desc"
	}
	if order != "asc" && order != "desc" {
		httpx.Error(w, http.StatusUnprocessableEntity, "order must be asc or desc")
		return
	}

	params := ListParams{
		ViewerID:      user.ID,
		ViewerIsAdmin: user.Role == models.RoleAdmin,
		AllUsers:      q.Get("scope") == "all",
		Status:        status,
		Search:        strings.TrimSpace(q.Get("search")),
		SortBy:        sortBy,
		Order:         order,
		Page:          atoiDefault(q.Get("page"), 1, 1),
		PageSize:      atoiDefault(q.Get("pageSize"), defaultPageSize, 1),
	}
	if params.PageSize > maxPageSize {
		params.PageSize = maxPageSize
	}

	result, err := h.repo.List(r.Context(), params)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list tasks")
		return
	}
	httpx.JSON(w, http.StatusOK, result)
}

// ── Get one ─────────────────────────────────────────────────────────────────

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	task, err := h.repo.GetByID(r.Context(), id, user.ID, user.Role == models.RoleAdmin)
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not fetch task")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"task": task})
}

// ── Update (PATCH) ──────────────────────────────────────────────────────────

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	// Decode into a raw map so we can tell "field omitted" from "field set to null".
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var p UpdateParams
	fieldErrs := map[string]string{}

	if v, ok := raw["title"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			fieldErrs["title"] = "must be a string"
		} else if s = strings.TrimSpace(s); s == "" || len(s) > 200 {
			fieldErrs["title"] = "must be between 1 and 200 characters"
		} else {
			p.Title = &s
		}
	}
	if v, ok := raw["description"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || len(s) > 5000 {
			fieldErrs["description"] = "must be a string up to 5000 characters"
		} else {
			p.Description = &s
		}
	}
	if v, ok := raw["status"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || !oneOf(s, "todo", "in_progress", "done") {
			fieldErrs["status"] = "must be one of: todo in_progress done"
		} else {
			p.Status = &s
		}
	}
	if v, ok := raw["priority"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || !oneOf(s, "low", "medium", "high") {
			fieldErrs["priority"] = "must be one of: low medium high"
		} else {
			p.Priority = &s
		}
	}
	if v, ok := raw["dueDate"]; ok {
		if string(v) == "null" {
			p.ClearDue = true
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				fieldErrs["dueDate"] = "must be an RFC3339 date-time or null"
			} else if t, perr := time.Parse(time.RFC3339, s); perr != nil {
				fieldErrs["dueDate"] = "must be an RFC3339 date-time or null"
			} else {
				p.DueDate = &t
			}
		}
	}

	if len(fieldErrs) > 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation failed", fieldErrs)
		return
	}

	task, err := h.repo.Update(r.Context(), id, user.ID, user.Role == models.RoleAdmin, p)
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not update task")
		return
	}

	h.hub.Publish(task.UserID, events.Event{Type: "task.updated", Data: task})
	httpx.JSON(w, http.StatusOK, map[string]any{"task": task})
}

// ── Delete ──────────────────────────────────────────────────────────────────

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.repo.Delete(r.Context(), id, user.ID, user.Role == models.RoleAdmin); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, "task not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "could not delete task")
		return
	}

	h.hub.Publish(user.ID, events.Event{Type: "task.deleted", Data: map[string]string{"id": id.String()}})
	w.WriteHeader(http.StatusNoContent)
}

// ── Activity (bonus) ────────────────────────────────────────────────────────

func (h *Handler) Activity(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	// Reuse GetByID purely for its visibility check.
	if _, err := h.repo.GetByID(r.Context(), id, user.ID, user.Role == models.RoleAdmin); err != nil {
		httpx.Error(w, http.StatusNotFound, "task not found")
		return
	}

	log, err := h.repo.Activity(r.Context(), id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not fetch activity")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"activity": log})
}

// ── SSE stream (bonus: real-time) ───────────────────────────────────────────

func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	flusher, ok := w.(http.Flusher)
	if !ok {
		httpx.Error(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsubscribe := h.hub.Subscribe(user.ID)
	defer unsubscribe()

	// Initial comment so the client knows the stream is open.
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second) // keep-alive ping
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		case ev := <-ch:
			payload, _ := json.Marshal(ev)
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid task id")
		return uuid.Nil, false
	}
	return id, true
}

func parseOptionalTime(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func defaultIfEmpty(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func oneOf(v string, opts ...string) bool {
	for _, o := range opts {
		if v == o {
			return true
		}
	}
	return false
}

func atoiDefault(s string, def, min int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < min {
		return def
	}
	return n
}
