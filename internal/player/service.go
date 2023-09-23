package player

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/xid"
)

type Service interface {
	Create(ctx context.Context, email, name string) (Player, error)
	Read(ctx context.Context, id xid.ID) (Player, error)
	Update(ctx context.Context, id xid.ID, email, name string) (Player, error)
	Delete(ctx context.Context, id xid.ID) error
	List(ctx context.Context) ([]Player, error)
	Filter(ctx context.Context, req FilterRequest, offset, limit uint) (total uint, pp []Player, err error)
}

type service struct {
	serviceName string
	storage     Storage
}

func NewService(serviceName string, storage Storage) Service {
	return &service{
		serviceName: serviceName,
		storage:     storage,
	}
}

func (c *service) Create(ctx context.Context, email, name string) (Player, error) {
	now := time.Now().UTC()
	p := Player{
		ID:        xid.New(),
		Version:   xid.New(),
		Email:     email,
		Name:      name,
		UpdatedAt: now,
		CreatedAt: now,
	}

	p, err := c.storage.Insert(ctx, p)
	if err != nil {
		return p, fmt.Errorf("insert player: %w", err)
	}

	return p, nil
}

func (c *service) Read(ctx context.Context, id xid.ID) (Player, error) {
	return c.storage.GetByID(ctx, id)
}

func (c *service) Update(ctx context.Context, id xid.ID, email, name string) (Player, error) {
	oldP, err := c.Read(ctx, id)
	if err != nil {
		return oldP, err
	}

	newP := oldP
	newP.Email = email
	newP.Name = name
	newP.Version = xid.New()
	newP.UpdatedAt = time.Now().UTC()

	p, err := c.storage.Replace(ctx, oldP, newP)
	if err != nil {
		return p, fmt.Errorf("replace player: %w", err)
	}

	return p, nil
}

func (c *service) Delete(ctx context.Context, id xid.ID) error {
	err := c.storage.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("delete player: %w", err)
	}

	return nil
}

func (c *service) List(ctx context.Context) ([]Player, error) {
	return c.storage.All(ctx)
}

func (c *service) Filter(
	ctx context.Context,
	req FilterRequest,
	offset,
	limit uint,
) (total uint, pp []Player, err error) {
	return c.storage.Filter(ctx, req, offset, limit)
}
