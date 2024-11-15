package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/CatalinPlesu/user-service/messaging"
	"github.com/CatalinPlesu/user-service/model"
	"github.com/CatalinPlesu/user-service/repository/user"
)

type User struct {
	// Repo *user.RedisRepo
	Repo     *user.PostgresRepo
	RabbitMQ *messaging.RabbitMQ
}

func (h *User) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	user := model.User{
		UserID:      uuid.New(),
		Username:    body.Username,
		DisplayName: body.DisplayName,
		Email:       body.Email,
		Password:    body.Password,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	err := h.Repo.Insert(r.Context(), user)
	if err != nil {
		fmt.Println("failed to insert user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID := user.UserID
	jwt := "jwt-token"

	err = h.RabbitMQ.PublishLoginRegisterMessage("user_login_register", userID, jwt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish to RabbitMQ: %v", err), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(user)
	if err != nil {
		fmt.Println("failed to marshal user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

func (h *User) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 50
	res, err := h.Repo.FindAll(r.Context(), user.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		fmt.Println("failed to find all users:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []model.User `json:"items"`
		Next  uint64       `json:"next,omitempty"`
	}
	response.Items = res.Users
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal users:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (h *User) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	userID, err := uuid.Parse(idParam) // Parse as UUID
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.Repo.FindByID(r.Context(), userID)
	if errors.Is(err, user.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find user by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(u); err != nil {
		fmt.Println("failed to marshal user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *User) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username    *string `json:"username,omitempty"`
		DisplayName *string `json:"display_name,omitempty"`
		Email       *string `json:"email,omitempty"`
		Password    *string `json:"password,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")

	userID, err := uuid.Parse(idParam) // Parse as UUID
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	theUser, err := h.Repo.FindByID(r.Context(), userID)
	if errors.Is(err, user.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find user by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	if body.Username != nil {
		theUser.Username = *body.Username
	}
	if body.DisplayName != nil {
		theUser.DisplayName = *body.DisplayName
	}
	if body.Email != nil {
		theUser.Email = *body.Email
	}
	if body.Password != nil {
		theUser.Password = *body.Password
	}
	theUser.UpdatedAt = &now

	err = h.Repo.Update(r.Context(), theUser)
	if err != nil {
		fmt.Println("failed to update user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(theUser); err != nil {
		fmt.Println("failed to marshal user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *User) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	userID, err := uuid.Parse(idParam) // Parse as UUID
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Repo.DeleteByID(r.Context(), userID)
	if errors.Is(err, user.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to delete user by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// func (r *RedisRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
// 	// Implement the logic to query by username in Redis
// }
//
// func (r *RedisRepo) FindByDisplayName(ctx context.Context, displayName string) (*model.User, error) {
// 	// Implement the logic to query by display name in Redis
// }
