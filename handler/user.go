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
	"github.com/CatalinPlesu/user-service/repository/jwts"
	"github.com/CatalinPlesu/user-service/repository/user"
)

type User struct {
	RdRepo   *jwts.RedisRepo
	PgRepo   *user.PostgresRepo
	RabbitMQ *messaging.RabbitMQ
}

func (h *User) Register(w http.ResponseWriter, r *http.Request) {
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

	err := h.PgRepo.Insert(r.Context(), user)
	if err != nil {
		fmt.Println("failed to insert user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID := user.UserID
	jwt, err := jwts.GenerateJWT(userID)
	if err != nil {
		fmt.Println("failed to generate jwt:", err)
		return
	}
	err = h.RdRepo.Insert(r.Context(), user.UserID, jwt)
	if err != nil {
		fmt.Println("failed to insert user jwt:", err)
		return
	}

	err = h.RabbitMQ.PublishLoginRegisterMessage("user_id_jwt", userID, jwt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish to RabbitMQ: %v", err), http.StatusInternalServerError)
		return
	}

	response := struct {
		User    model.User `json:"user"`
		UserJWT string     `json:"jwt"`
	}{
		User:    user,
		UserJWT: jwt,
	}

	res, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

func (h *User) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.PgRepo.FindByUsername(r.Context(), body.Username)
	if errors.Is(err, user.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find user by username:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if body.Password == u.Password {
		fmt.Println("Succes login")
	} else {

		fmt.Println("fail login")
	}

	if err := json.NewEncoder(w).Encode(u); err != nil {
		fmt.Println("failed to marshal user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID := u.UserID
	jwt, err := jwts.GenerateJWT(userID)
	if err != nil {
		fmt.Println("failed to generate jwt:", err)
		return
	}

	err = h.RdRepo.Insert(r.Context(), u.UserID, jwt)
	if err != nil {
		fmt.Println("failed to insert user jwt:", err)
		return
	}

	err = h.RabbitMQ.PublishLoginRegisterMessage("user_id_jwt", userID, jwt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish to RabbitMQ: %v", err), http.StatusInternalServerError)
		return
	}

	response := struct {
		User    model.User `json:"user"`
		UserJWT string     `json:"jwt"`
	}{
		User:    *u,
		UserJWT: jwt,
	}

	res, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

func (h *User) Auth(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		JWT      string `json:"jwt"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.PgRepo.FindByUsername(r.Context(), body.Username)
	if errors.Is(err, user.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find user by username:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID := u.UserID
	claims, err := jwts.ValidateJWT(body.JWT)
	if err != nil {
		fmt.Println("bad jwt jwt:", err)
		return
	}

	if claims.UserID != userID {
		fmt.Println("bad jwt no acces or expierd")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
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
	res, err := h.PgRepo.FindAll(r.Context(), user.FindAllPage{
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

	u, err := h.PgRepo.FindByID(r.Context(), userID)
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

func (h *User) GetByDisplayName(w http.ResponseWriter, r *http.Request) {
	displayNameParam := chi.URLParam(r, "displayname")

	res, err := h.PgRepo.FindByDisplayName(r.Context(), displayNameParam)
	if err != nil {
		fmt.Println("failed to find all users:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(res)
	if err != nil {
		fmt.Println("failed to marshal users:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (h *User) GetByUsername(w http.ResponseWriter, r *http.Request) {
	usernameParam := chi.URLParam(r, "username")

	u, err := h.PgRepo.FindByUsername(r.Context(), usernameParam)
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

	theUser, err := h.PgRepo.FindByID(r.Context(), userID)
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

	err = h.PgRepo.Update(r.Context(), theUser)
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

	err = h.PgRepo.DeleteByID(r.Context(), userID)
	if errors.Is(err, user.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to delete user by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
