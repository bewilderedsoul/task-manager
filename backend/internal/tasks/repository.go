package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskmanager/internal/models"
)

// ErrNotFound is returned when a task does not exist or is not visible to the caller.
var ErrNotFound = errors.New("task not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateParams carries the validated fields for a new task.
type CreateParams struct {
	UserID      uuid.UUID
	Title       string
	Description string
	Status      string
	Priority    string
	DueDate     *time.Time
}

// UpdateParams carries optional fields for a PATCH. Nil pointers are left unchanged.
type UpdateParams struct {
	Title       *string
	Description *string
	Status      *string
	Priority    *string
	DueDate     *time.Time
	ClearDue    bool // when true, due_date is explicitly set to NULL
}

// ListParams controls filtering, search, sort and pagination.
type ListParams struct {
	// Scope: when ViewerIsAdmin && AllUsers, list tasks across all users.
	// Otherwise results are restricted to ViewerID.
	ViewerID      uuid.UUID
	ViewerIsAdmin bool
	AllUsers      bool

	Status string // "" means no status filter
	Search string // case-insensitive title search; "" means no search

	SortBy string // created_at | due_date | priority
	Order  string // asc | desc

	Page     int // 1-based
	PageSize int
}

// ListResult is a page of tasks plus the total count for pagination UIs.
type ListResult struct {
	Tasks    []models.Task `json:"tasks"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

const taskColumns = `id, user_id, title, description, status, priority, due_date, created_at, updated_at`

func scanTask(row pgx.Row) (*models.Task, error) {
	var t models.Task
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) Create(ctx context.Context, p CreateParams) (*models.Task, error) {
	const q = `
		INSERT INTO tasks (user_id, title, description, status, priority, due_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + taskColumns

	row := r.pool.QueryRow(ctx, q, p.UserID, p.Title, p.Description, p.Status, p.Priority, p.DueDate)
	task, err := scanTask(row)
	if err != nil {
		return nil, err
	}
	r.logActivity(ctx, task.ID, p.UserID, "created", map[string]any{"title": task.Title})
	return task, nil
}

// GetByID fetches a single task, enforcing ownership unless the viewer is an admin.
func (r *Repository) GetByID(ctx context.Context, id, viewerID uuid.UUID, viewerIsAdmin bool) (*models.Task, error) {
	q := `SELECT ` + taskColumns + ` FROM tasks WHERE id = $1`
	args := []any{id}
	if !viewerIsAdmin {
		q += ` AND user_id = $2`
		args = append(args, viewerID)
	}

	task, err := scanTask(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return task, nil
}

// List returns a filtered, sorted, paginated page of tasks plus a total count.
func (r *Repository) List(ctx context.Context, p ListParams) (*ListResult, error) {
	var where []string
	var args []any

	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}

	if !(p.ViewerIsAdmin && p.AllUsers) {
		add("user_id = $%d", p.ViewerID)
	}
	if p.Status != "" {
		add("status = $%d", p.Status)
	}
	if p.Search != "" {
		add("title ILIKE $%d", "%"+p.Search+"%")
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	// Count first (for pagination metadata).
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM tasks"+whereClause, args...).Scan(&total); err != nil {
		return nil, err
	}

	orderClause := buildOrderClause(p.SortBy, p.Order)

	limit := p.PageSize
	offset := (p.Page - 1) * p.PageSize
	args = append(args, limit, offset)
	q := "SELECT " + taskColumns + " FROM tasks" + whereClause + orderClause +
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.Task, 0, p.PageSize)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ListResult{Tasks: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// buildOrderClause maps the validated sort field to safe SQL. Inputs are already
// constrained to a whitelist by the handler, so this never interpolates raw user text.
func buildOrderClause(sortBy, order string) string {
	dir := "ASC"
	if strings.EqualFold(order, "desc") {
		dir = "DESC"
	}
	switch sortBy {
	case "due_date":
		// NULL due dates always sort last regardless of direction.
		return fmt.Sprintf(" ORDER BY due_date %s NULLS LAST, created_at DESC", dir)
	case "priority":
		// Map the text enum to a rank so ordering is semantic, not alphabetical.
		return fmt.Sprintf(" ORDER BY CASE priority WHEN 'high' THEN 3 WHEN 'medium' THEN 2 ELSE 1 END %s, created_at DESC", dir)
	default: // created_at
		return fmt.Sprintf(" ORDER BY created_at %s", dir)
	}
}

func (r *Repository) Update(ctx context.Context, id, viewerID uuid.UUID, viewerIsAdmin bool, p UpdateParams) (*models.Task, error) {
	var sets []string
	var args []any
	add := func(col string, val any) {
		args = append(args, val)
		sets = append(sets, fmt.Sprintf("%s = $%d", col, len(args)))
	}

	if p.Title != nil {
		add("title", *p.Title)
	}
	if p.Description != nil {
		add("description", *p.Description)
	}
	if p.Status != nil {
		add("status", *p.Status)
	}
	if p.Priority != nil {
		add("priority", *p.Priority)
	}
	if p.ClearDue {
		sets = append(sets, "due_date = NULL")
	} else if p.DueDate != nil {
		add("due_date", *p.DueDate)
	}

	if len(sets) == 0 {
		// Nothing to change — just return the current row (still enforces visibility).
		return r.GetByID(ctx, id, viewerID, viewerIsAdmin)
	}
	sets = append(sets, "updated_at = now()")

	args = append(args, id)
	q := "UPDATE tasks SET " + strings.Join(sets, ", ") + fmt.Sprintf(" WHERE id = $%d", len(args))
	if !viewerIsAdmin {
		args = append(args, viewerID)
		q += fmt.Sprintf(" AND user_id = $%d", len(args))
	}
	q += " RETURNING " + taskColumns

	task, err := scanTask(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	r.logActivity(ctx, task.ID, viewerID, "updated", changedFields(p))
	return task, nil
}

func (r *Repository) Delete(ctx context.Context, id, viewerID uuid.UUID, viewerIsAdmin bool) error {
	q := "DELETE FROM tasks WHERE id = $1"
	args := []any{id}
	if !viewerIsAdmin {
		q += " AND user_id = $2"
		args = append(args, viewerID)
	}
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Activity returns the change history for a task (most recent first).
func (r *Repository) Activity(ctx context.Context, taskID uuid.UUID) ([]models.Activity, error) {
	const q = `
		SELECT id, task_id, user_id, action, detail, created_at
		FROM task_activity WHERE task_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Activity, 0)
	for rows.Next() {
		var a models.Activity
		if err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.Action, &a.Detail, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// logActivity records a change. Failures are intentionally swallowed: the
// activity log is supplementary and must never fail the primary operation.
func (r *Repository) logActivity(ctx context.Context, taskID, userID uuid.UUID, action string, detail map[string]any) {
	const q = `INSERT INTO task_activity (task_id, user_id, action, detail) VALUES ($1, $2, $3, $4)`
	_, _ = r.pool.Exec(ctx, q, taskID, userID, action, detail)
}

func changedFields(p UpdateParams) map[string]any {
	changed := map[string]any{}
	if p.Title != nil {
		changed["title"] = *p.Title
	}
	if p.Description != nil {
		changed["description"] = *p.Description
	}
	if p.Status != nil {
		changed["status"] = *p.Status
	}
	if p.Priority != nil {
		changed["priority"] = *p.Priority
	}
	if p.ClearDue {
		changed["dueDate"] = nil
	} else if p.DueDate != nil {
		changed["dueDate"] = *p.DueDate
	}
	return changed
}
