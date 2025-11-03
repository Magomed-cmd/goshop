package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"goshop/internal/core/domain/entities"
	errors2 "goshop/internal/core/domain/errors"
	"goshop/internal/core/mappers"
	"goshop/internal/core/ports/repositories"
	"goshop/internal/dto"
)

type CartService struct {
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
}

func NewCartService(cartRepo repositories.CartRepository, productRepo repositories.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *CartService) GetCart(ctx context.Context, userID int64) (*dto.CartResponse, error) {
	cart, err := s.cartRepo.GetUserCart(ctx, userID)
	if err != nil {
		if errors.Is(err, errors2.ErrCartNotFound) {
			cart, err = s.createCartForUser(ctx, userID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	response := mappers.ToCartResponse(cart)
	if response == nil {
		return nil, errors2.ErrCartNotFound
	}

	return response, nil
}

func (s *CartService) AddItem(ctx context.Context, userID int64, req *dto.AddToCartRequest) error {
	if req.Quantity <= 0 {
		return errors2.ErrInvalidQuantity
	}

	product, err := s.productRepo.GetProductByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, errors2.ErrProductNotFound) {
			return errors2.ErrProductNotFound
		}
		return err
	}

	if product.Stock < req.Quantity {
		return errors2.ErrInsufficientStock
	}

	cart, err := s.cartRepo.GetUserCart(ctx, userID)
	if err != nil {
		if errors.Is(err, errors2.ErrCartNotFound) {
			cart, err = s.createCartForUser(ctx, userID)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return s.cartRepo.AddItem(ctx, cart.ID, req.ProductID, req.Quantity)
}

func (s *CartService) UpdateItem(ctx context.Context, userID int64, productID int64, quantity int) error {
	if quantity <= 0 {
		return errors2.ErrInvalidQuantity
	}

	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return err
	}

	if product.Stock < quantity {
		return errors2.ErrInsufficientStock
	}

	cart, err := s.cartRepo.GetUserCart(ctx, userID)
	if err != nil {
		return err
	}

	return s.cartRepo.UpdateItem(ctx, cart.ID, productID, quantity)
}

func (s *CartService) RemoveItem(ctx context.Context, userID int64, productID int64) error {
	cart, err := s.cartRepo.GetUserCart(ctx, userID)
	if err != nil {
		return err
	}

	return s.cartRepo.RemoveItem(ctx, cart.ID, productID)
}

func (s *CartService) ClearCart(ctx context.Context, userID int64) error {
	cart, err := s.cartRepo.GetUserCart(ctx, userID)
	if err != nil {
		return err
	}

	return s.cartRepo.ClearCart(ctx, cart.ID)
}

func (s *CartService) createCartForUser(ctx context.Context, userID int64) (*entities.Cart, error) {
	cart := &entities.Cart{
		UUID:      uuid.New(),
		UserID:    userID,
		CreatedAt: time.Now(),
		Items:     []entities.CartItem{},
	}

	err := s.cartRepo.CreateCart(ctx, cart)
	if err != nil {
		return nil, err
	}

	return cart, nil
}
