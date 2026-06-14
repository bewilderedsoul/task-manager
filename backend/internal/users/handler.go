package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"taskmanager/internal/auth"
	"taskmanager/internal/httpx"
	"taskmanager/internal/middleware"
	"taskmanager/internal/models"
	"taskmanager/internal/validate"
)

type Handler struct {
	repo *Repository
	jwt  *auth.Manager
}

func NewHandler(repo *Repository, jwt *auth.Manager) *Handler {
	return &Handler{repo: repo, jwt: jwt}
}

type credentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// Signup creates a new account and returns a JWT.
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if errs := validate.Struct(in); errs != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation failed", errs)
		return
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not process password")
		return
	}

	user, err := h.repo.Create(r.Context(), in.Email, hash, models.RoleUser)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			httpx.Error(w, http.StatusConflict, "an account with this email already exists")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "could not create account")
		return
	}

	h.issueToken(w, http.StatusCreated, user)
}

// Login verifies credentials and returns a JWT.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	user, err := h.repo.GetByEmail(r.Context(), in.Email)
	// Return the same error for "no such user" and "wrong password" so the
	// endpoint can't be used to enumerate which emails are registered.
	if err != nil || !auth.CheckPassword(user.PasswordHash, in.Password) {
		httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	h.issueToken(w, http.StatusOK, user)
}

// Me returns the currently authenticated user. Used by the frontend on load to
// restore session state after a refresh.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	user, err := h.repo.getByID(r.Context(), authUser.ID.String())
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *Handler) issueToken(w http.ResponseWriter, status int, user *models.User) {
	token, err := h.jwt.Generate(user.ID, user.Role)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	httpx.JSON(w, status, authResponse{Token: token, User: user})
}

// getByID is a small helper kept on the repository for the Me endpoint.
func (r *Repository) getByID(ctx context.Context, id string) (*models.User, error) {
	const q = `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users WHERE id = $1`
	var u models.User
	if err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &u, nil
}
