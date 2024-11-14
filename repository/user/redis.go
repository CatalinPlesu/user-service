package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"

	"github.com/CatalinPlesu/user-service/model"
)

type RedisRepo struct {
	Client *redis.Client
}

func userIDKey(id uuid.UUID) string {
	return fmt.Sprintf("user:%s", id.String())
}

func (r *RedisRepo) Insert(ctx context.Context, user model.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to encode user: %w", err)
	}

	key := userIDKey(user.UserID)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set user: %w", err)
	}

	if err := txn.SAdd(ctx, "users", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add user to set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

var ErrNotExist = errors.New("user does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	key := userIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.User{}, ErrNotExist
	} else if err != nil {
		return model.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	var user model.User
	err = json.Unmarshal([]byte(value), &user)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to decode user json: %w", err)
	}

	return user, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	key := userIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := txn.SRem(ctx, "users", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove user from set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, user model.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to encode user: %w", err)
	}

	key := userIDKey(user.UserID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Users  []model.User
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "users", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get user ids: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Users: []model.User{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get users: %w", err)
	}

	users := make([]model.User, len(xs))

	for i, x := range xs {
		x := x.(string)
		var user model.User

		err := json.Unmarshal([]byte(x), &user)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode user json: %w", err)
		}

		users[i] = user
	}

	return FindResult{
		Users: users,
		Cursor: cursor,
	}, nil
}
