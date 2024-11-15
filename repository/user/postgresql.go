package user

import (
	"context"
	"fmt"

	"github.com/CatalinPlesu/user-service/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PostgresRepo struct {
	DB *bun.DB
}

func NewPostgresRepo(db *bun.DB) *PostgresRepo {
	return &PostgresRepo{DB: db}
}

func (p *PostgresRepo) Migrate(ctx context.Context) error {
	_, err := p.DB.NewCreateTable().
		Model((*model.User)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	return nil
}

func (p *PostgresRepo) Insert(ctx context.Context, user model.User) error {
	_, err := p.DB.NewInsert().Model(&user).Exec(ctx)
	if err != nil {
		p.Migrate(ctx)
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (p *PostgresRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := p.DB.NewSelect().Model(&user).Where("user_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	return &user, nil
}

func (p *PostgresRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := p.DB.NewSelect().Model(&user).Where("username = ?", username).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	return &user, nil
}

func (p *PostgresRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	_, err := p.DB.NewDelete().Model((*model.User)(nil)).Where("user_id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (p *PostgresRepo) Update(ctx context.Context, user *model.User) error {
	_, err := p.DB.NewUpdate().Model(user).Where("user_id = ?", user.UserID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

type UserPage struct {
	Users  []model.User
	Cursor uint64
}

func (r *PostgresRepo) FindByDisplayName(ctx context.Context, displayName string) ([]model.User, error) {
	var users []model.User

	// Query the database for users
	query := r.DB.NewSelect().
		Model(&users).
		Where("display_name ILIKE ?", "%"+displayName+"%").
		Order("user_id ASC").
		Limit(10)

	// Execute the query
	err := query.Scan(ctx)
	if err != nil {
		return users, fmt.Errorf("failed to retrieve users: %w", err)
	}

	return users, nil
}

func (r *PostgresRepo) FindAll(ctx context.Context, page FindAllPage) (UserPage, error) {
	var users []model.User

	// Query the database for users
	query := r.DB.NewSelect().
		Model(&users).
		Order("user_id ASC").
		Limit(int(page.Size))

	// If a cursor is provided, only retrieve users with an ID greater than the cursor
	if page.Offset > 0 {
		query.Where("user_id > ?", page.Offset)
	}

	// Execute the query
	err := query.Scan(ctx)
	if err != nil {
		return UserPage{}, fmt.Errorf("failed to retrieve users: %w", err)
	}

	// If no users were found, return an empty result
	if len(users) == 0 {
		return UserPage{
			Users:  []model.User{},
			Cursor: 0,
		}, nil
	}

	return UserPage{
		Users:  users,
		Cursor: page.Size + 50,
	}, nil
}
