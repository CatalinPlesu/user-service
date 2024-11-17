package jwts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/CatalinPlesu/user-service/model"
)

type RedisRepo struct {
	Client *redis.Client
}

var ErrNotExist = errors.New("user JWTs do not exist")
var ErrJWTNotFound = errors.New("JWT not found for the user")

func userJWTsKey(id uuid.UUID) string {
	return fmt.Sprintf("user_jwts:%s", id.String())
}

func (r *RedisRepo) Insert(ctx context.Context, userID uuid.UUID, jwt string) error {
	key := userJWTsKey(userID)

	var userJWTs model.UserJWTs
	value, err := r.Client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get user JWTs: %w", err)
	}

	if err == redis.Nil {
		userJWTs = model.UserJWTs{
			UserID: userID,
			JWTs:   []string{jwt},
		}
	} else {
		err = json.Unmarshal([]byte(value), &userJWTs)
		if err != nil {
			return fmt.Errorf("failed to decode user JWTs json: %w", err)
		}

		for _, existingJWT := range userJWTs.JWTs {
			if existingJWT == jwt {
				return nil
			}
		}

		userJWTs.JWTs = append(userJWTs.JWTs, jwt)
	}

	data, err := json.Marshal(userJWTs)
	if err != nil {
		return fmt.Errorf("failed to encode user JWTs: %w", err)
	}

	err = r.Client.Set(ctx, key, string(data), 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set user JWTs: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, userID uuid.UUID, jwts []string) error {
	key := userJWTsKey(userID)

	userJWTs := model.UserJWTs{
		UserID: userID,
		JWTs:   jwts,
	}

	data, err := json.Marshal(userJWTs)
	if err != nil {
		return fmt.Errorf("failed to encode user JWTs: %w", err)
	}

	err = r.Client.Set(ctx, key, string(data), 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set user JWTs: %w", err)
	}

	return nil
}
