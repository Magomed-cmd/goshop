# GoShop Architecture

## Overview

GoShop follows **Clean Architecture** combined with **Hexagonal Architecture** (Ports & Adapters) patterns. This approach ensures:

- Clear separation of concerns
- Easy testing and maintenance
- Easy to swap implementations
- Domain logic is independent of frameworks

## Layered Structure

### 1. Presentation Layer (adapters/input/http)

**Responsibility**: Handle HTTP requests/responses and convert them to/from domain objects.

```
HTTP Client
    ↓
[HTTP Handler]
    ↓
[Validation & Auth Middleware]
    ↓
[DTO Conversion]
    ↓
[Application Service]
```

**Components**:
- **HTTP Handlers** (`*_handler.go`): Accept HTTP requests, validate input, call services
- **Mappers** (`mappers/*.go`): Convert between DTOs and domain entities
- **Middleware**: Authentication, logging, CORS, error handling
- **Error Handler**: Converts domain errors to HTTP responses
- **DTOs**: Request/Response objects for API contracts

**Key Files**:
- `adapters/input/http/user_handler.go`
- `adapters/input/http/product_handler.go`
- `adapters/input/http/order_handler.go`
- `middleware/auth.go`

**Example Flow**:
```go
// HTTP Handler
func (h *UserHandler) Register(c *gin.Context) {
    var req dto.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // Validation error
        return
    }
    
    // Call service
    user, err := h.userService.Register(c.Request.Context(), req)
    if err != nil {
        // Handle error
        return
    }
    
    // Convert to DTO and respond
    c.JSON(200, toUserResponse(user))
}
```

---

### 2. Application Layer (core/services)

**Responsibility**: Implement use cases and business orchestration.

```
[Service]
    ↓
[Validation]
    ↓
[Business Logic]
    ↓
[Repository Calls]
    ↓
[Mapper to Domain]
```

**Components**:
- **Services**: Implement use cases (e.g., UserService.Register, ProductService.GetById)
- **Transaction Management**: Coordinate multi-step operations

**Key Responsibilities**:
1. Orchestrate domain entities
2. Call repositories and other services
3. Handle transaction boundaries
4. Implement business logic and validation

**Example**:
```go
func (s *UserService) Register(ctx context.Context, req dto.RegisterRequest) (*User, error) {
    // Validation
    if err := validateEmail(req.Email); err != nil {
        return nil, ErrInvalidEmail
    }
    
    // Check if user exists
    exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, ErrUserAlreadyExists
    }
    
    // Create domain entity
    user := entities.NewUser(req.Email, req.Password, req.Name)
    
    // Save
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    return user, nil
}
```

**Key Files**:
- `core/services/user.go`
- `core/services/product.go`
- `core/services/order.go`

---

### 3. Domain Layer (core/domain)

**Responsibility**: Contain pure business logic and rules.

```
[Entity]
    ↓
[Value Objects]
    ↓
[Business Rules]
    ↓
[Domain Errors]
```

**Components**:

#### Entities
- **Definition**: Objects with identity, lifecycle, and state changes
- **Examples**: User, Product, Order, Cart
- **Characteristics**: 
  - Have unique ID
  - Can be created, modified, deleted
  - Contain business rules

```go
// core/domain/entities/user.go
type User struct {
    ID       int64
    UUID     string
    Email    string
    Password string // hashed
    Name     *string
    Role     Role
    CreatedAt time.Time
}

// Domain method with business rule
func (u *User) ChangePassword(oldPwd, newPwd string) error {
    if !u.VerifyPassword(oldPwd) {
        return ErrInvalidPassword
    }
    u.Password = hashPassword(newPwd)
    return nil
}
```

#### Value Objects (VO)
- **Definition**: Objects without identity, immutable
- **Examples**: Money, Email, Address, Rating
- **Characteristics**:
  - No identity (equality by value)
  - Immutable
  - Encapsulate domain concept

```go
// core/domain/vo/email.go
type Email struct {
    value string
}

func NewEmail(val string) (Email, error) {
    if !isValidEmail(val) {
        return Email{}, ErrInvalidEmail
    }
    return Email{value: val}, nil
}

func (e Email) String() string { return e.value }
```

#### Domain Errors
- Custom errors specific to business rules
- Contain semantic meaning
- Can be caught and handled appropriately

```go
// core/domain/errors/errors.go
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrInvalidPassword   = errors.New("invalid password")
    ErrProductOutOfStock = errors.New("product out of stock")
)
```

**Key Files**:
- `core/domain/entities/user.go`
- `core/domain/entities/product.go`
- `core/domain/entities/order.go`
- `core/domain/vo/email.go`
- `core/domain/errors/errors.go`

---

### 4. Ports & Interfaces (core/ports)

**Responsibility**: Define contracts for external dependencies.

```
[Domain]
    ↓
[Ports (Interfaces)]
    ↓
[Infrastructure Implementations]
```

**Why Ports?**
- Domain doesn't depend on infrastructure
- Easy to mock for testing
- Can swap implementations without changing domain code

**Types of Ports**:

#### Repository Ports
```go
// core/ports/repositories/user.go
type UserRepository interface {
    Create(ctx context.Context, user *entities.User) error
    GetByID(ctx context.Context, id int64) (*entities.User, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    Update(ctx context.Context, user *entities.User) error
    Delete(ctx context.Context, id int64) error
}
```

#### Cache Ports
```go
// core/ports/cache/product.go
type ProductCache interface {
    Get(ctx context.Context, key string) (*entities.Product, error)
    Set(ctx context.Context, key string, product *entities.Product, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

#### Storage Ports
```go
// core/ports/storage/file.go
type FileStorage interface {
    Upload(ctx context.Context, file *File) (url string, err error)
    Download(ctx context.Context, path string) ([]byte, error)
    Delete(ctx context.Context, path string) error
}
```

#### Transaction Ports
```go
// core/ports/transaction/runner.go
type TransactionRunner interface {
    WithTx(ctx context.Context, fn func(context.Context) error) error
}
```

**Key Files**:
- `core/ports/repositories/user.go`
- `core/ports/repositories/product.go`
- `core/ports/cache/product.go`
- `core/ports/storage/file.go`
- `core/ports/transaction/runner.go`

---

### 5. Infrastructure Layer (adapters/output + infrastructure)

**Responsibility**: Implement concrete details - database, cache, storage, etc.

```
[Port Interface]
    ↓
[Implementation]
    ↓
[Database/Cache/Storage]
```

**Components**:

#### Database Adapters
```go
// adapters/output/database/user_repository.go
type UserRepository struct {
    db *pgx.Conn
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
    query := `INSERT INTO users (uuid, email, password, name, role) VALUES ($1, $2, $3, $4, $5)`
    _, err := r.db.Exec(ctx, query, user.UUID, user.Email, user.Password, user.Name, user.Role)
    return err
}
```

#### Cache Adapters
```go
// adapters/output/cache/product_cache.go
type ProductCache struct {
    redis *redis.Client
}

func (c *ProductCache) Get(ctx context.Context, key string) (*entities.Product, error) {
    val, err := c.redis.Get(ctx, key).Result()
    // Unmarshal from JSON
    // Return product
}
```

#### Storage Adapters
```go
// adapters/output/storage/s3_storage.go
type S3Storage struct {
    client *s3.Client
}

func (s *S3Storage) Upload(ctx context.Context, file *File) (string, error) {
    // Upload to S3
    // Return URL
}
```

**Key Files**:
- `adapters/output/database/postgres/*`
- `adapters/output/cache/redis.go`
- `adapters/output/storage/s3.go`
- `infrastructure/database/postgres/connection.go`
- `infrastructure/database/redis/connection.go`

---

## Data Flow Example: Create Order

### Request Path

```
1. HTTP Request
   POST /orders
   {
     "address_id": 123
   }
        ↓
2. HTTP Handler (order_handler.go)
   - Validate request
   - Extract user from context
        ↓
3. Service (order.go)
   - Get user's cart
   - Get address
   - Create order domain entity
   - Transaction start
   - Save order to repository
   - Clear cart
   - Transaction commit
        ↓
4. Repository (database/order_repository.go)
   - Execute INSERT query
   - Return created order
        ↓
5. Mapper (mappers/order.go)
   - Convert entity to DTO
        ↓
6. HTTP Response
   {
     "id": 1,
     "uuid": "abc-123",
     "status": "pending",
     ...
   }
```

### Code Example

```go
// Presentation Layer
func (h *OrderHandler) Create(c *gin.Context) {
    var req dto.CreateOrderRequest
    c.ShouldBindJSON(&req)
    
    userID := c.GetInt64("user_id")
    order, err := h.orderService.CreateOrder(c.Request.Context(), userID, req)
    
    c.JSON(201, toOrderResponse(order))
}

// Application Layer
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, req dto.CreateOrderRequest) (*entities.Order, error) {
    // Get cart
    cart, err := s.cartRepo.GetByUserID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Create entity
    order := entities.NewOrder(userID, cart.Items, req.AddressID)
    
    // Transaction
    return order, s.txRunner.WithTx(ctx, func(txCtx context.Context) error {
        // Save order
        if err := s.orderRepo.Create(txCtx, order); err != nil {
            return err
        }
        // Clear cart
        return s.cartRepo.Clear(txCtx, userID)
    })
}

// Domain Layer
func NewOrder(userID int64, items []CartItem, addressID *int64) *Order {
    return &Order{
        ID:        0,
        UserID:    userID,
        AddressID: addressID,
        Status:    "pending",
        Items:     items,
        TotalPrice: calculateTotal(items),
        CreatedAt: time.Now(),
    }
}

// Infrastructure Layer
func (r *OrderRepository) Create(ctx context.Context, order *entities.Order) error {
    query := `INSERT INTO orders (user_id, address_id, total_price, status) VALUES ($1, $2, $3, $4) RETURNING id`
    return r.db.QueryRow(ctx, query, order.UserID, order.AddressID, order.TotalPrice, order.Status).Scan(&order.ID)
}
```

---

## Dependency Direction

```
Presentation → Application → Domain ← Infrastructure
                    ↑           ↑
                 [Ports]--------┘
```

**Rules**:
1. ✅ Presentation depends on Application
2. ✅ Application depends on Domain
3. ✅ Infrastructure implements Ports (Domain interfaces)
4. ❌ Domain never depends on Application or Infrastructure
5. ❌ Application never depends on Infrastructure directly

---

## Testing Strategy

### Unit Tests
- Test domain entities and value objects
- Test services with mocked repositories
- Test error handling

```go
func TestOrderCreation(t *testing.T) {
    order := entities.NewOrder(1, items, nil)
    
    assert.Equal(t, "pending", order.Status)
    assert.Equal(t, expectedTotal, order.TotalPrice)
}
```

### Integration Tests
- Test repositories with real database
- Test service layer end-to-end

```go
func TestCreateOrder(t *testing.T) {
    service := NewOrderService(mockRepo, txRunner)
    order, err := service.CreateOrder(ctx, userID, req)
    
    assert.NoError(t, err)
    assert.NotNil(t, order)
}
```

### Handler Tests
- Mock services
- Test HTTP interface

```go
func TestCreateOrderHandler(t *testing.T) {
    mockService := &MockOrderService{}
    handler := NewOrderHandler(mockService)
    
    // Test request/response
}
```

---

## Key Design Patterns Used

### 1. Repository Pattern
Abstracts data access logic behind interfaces.

### 2. Service Layer Pattern
Coordinates use cases and domain logic.

### 3. Mapper Pattern
Converts between DTOs and domain entities.

### 4. Dependency Injection
Services receive dependencies via constructor.

### 5. Ports & Adapters
Infrastructure implements domain interfaces.

### 6. Error Handling
Domain defines errors, presentation translates to HTTP status codes.

### 7. Context Usage
Passed through all layers for cancellation and timeouts.

---

## Configuration & Bootstrapping

**Entry Point**: `cmd/goshop/main.go`

```
1. Load Configuration (config/)
2. Initialize Logger (logger/)
3. Connect to Databases (infrastructure/)
4. Create Repositories (adapters/output/)
5. Create Services (core/services/)
6. Create HTTP Handlers (adapters/input/http/)
7. Setup Routes (adapters/input/http/)
8. Start Server
```

---

## Caching Strategy

### Product Cache
- Cache products by ID
- TTL: 1 hour
- Invalidate on update

### Category Cache
- Cache categories list
- TTL: 2 hours
- Invalidate on update

### Review Cache
- Cache review stats
- TTL: 30 minutes
- Invalidate on new review

---

## Transaction Management

### Unit of Work Pattern

```go
// Start transaction
txRunner.WithTx(ctx, func(txCtx context.Context) error {
    // Multiple repository operations
    userRepo.Create(txCtx, user)
    cartRepo.Clear(txCtx, userID)
    // All succeed or all rollback
})
```

---

## Future Improvements

- [ ] Add event sourcing for order changes
- [ ] Implement CQRS for read models
- [ ] Add background job processing (orders, notifications)
- [ ] Implement API versioning
- [ ] Add rate limiting
- [ ] Add GraphQL endpoint alongside REST

---

## Resources

- [Clean Architecture - Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture - Alistair Cockburn](https://alistair.cockburn.us/hexagonal-architecture/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Domain-Driven Design - Eric Evans](https://www.domainlanguage.com/ddd/)
