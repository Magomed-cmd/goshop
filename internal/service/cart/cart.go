package cart

import (
	"context"
	"errors"
	errors2 "goshop/internal/domain/errors"
	"time"

	"goshop/internal/domain/entities"
	"goshop/internal/dto"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CartRepository interface {
	GetUserCart(ctx context.Context, userID int64) (*entities.Cart, error)
	CreateCart(ctx context.Context, cart *entities.Cart) error
	AddItem(ctx context.Context, cartID int64, productID int64, quantity int) error
	UpdateItem(ctx context.Context, cartID int64, productID int64, quantity int) error
	RemoveItem(ctx context.Context, cartID int64, productID int64) error
	ClearCart(ctx context.Context, cartID int64) error
}

type ProductRepository interface {
	GetProductByID(ctx context.Context, id int64) (*entities.Product, error)
}

type CartService struct {
	cartRepo    CartRepository
	productRepo ProductRepository
}

func NewCartService(cartRepo CartRepository, productRepo ProductRepository) *CartService {
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

	items := cart.Items
	itemsForResponse := make([]dto.CartItemResponse, len(items))

	totalPrice := decimal.Zero
	totalItems := 0

	for i, item := range items {
		priceDecimal := item.Product.Price
		subtotalDecimal := priceDecimal.Mul(decimal.NewFromInt(int64(item.Quantity)))

		itemsForResponse[i] = dto.CartItemResponse{
			ProductID:   item.Product.ID,
			ProductName: item.Product.Name,
			Quantity:    item.Quantity,
			Price:       priceDecimal.StringFixed(2),
			Subtotal:    subtotalDecimal.StringFixed(2),
		}

		totalPrice = totalPrice.Add(subtotalDecimal)
		totalItems += item.Quantity
	}

	response := &dto.CartResponse{
		ID:         cart.ID,
		Items:      itemsForResponse,
		TotalPrice: totalPrice.StringFixed(2),
		TotalItems: totalItems,
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
