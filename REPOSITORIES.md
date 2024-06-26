### Step-by-Step Guide for Creating, Registering, and Using a Product Repository

### Step 1: Define Your Product Model

**/repository/models/product.go**
```go
package models

import "time"

// Product represents a product with details and timestamps
type Product struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Price       float64   `db:"price" json:"price"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
```

### Step 2: Create Repository Interface

**/repository/interface.go**
```go
package repository

import "github.com/vert-pjoubert/goth-template/repository/models"

// ProductRepository defines methods for product operations
type ProductRepository interface {
	GetProductByID(id int64) (*models.Product, error)
	CreateProduct(product *models.Product) error
}
```

### Step 3: Implement Repository Methods

**/repository/impl_product_repository.go**
```go
package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/repository/models"
)

// SQLProductRepository is an implementation of ProductRepository interface
type SQLProductRepository struct {
	db *sqlx.DB
}

// NewSQLProductRepository creates a new SQLProductRepository
func NewSQLProductRepository(db *sqlx.DB) *SQLProductRepository {
	return &SQLProductRepository{db: db}
}

// GetProductByID retrieves a product by ID
func (repo *SQLProductRepository) GetProductByID(id int64) (*models.Product, error) {
	product := &models.Product{}
	err := repo.db.Get(product, "SELECT * FROM products WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// CreateProduct creates a new product
func (repo *SQLProductRepository) CreateProduct(product *models.Product) error {
	query := `INSERT INTO products (name, description, price, created_at, updated_at) VALUES (:name, :description, :price, :created_at, :updated_at)`
	_, err := repo.db.NamedExec(query, product)
	return err
}
```

### Step 4: AppStore with Repository Registration

**store/appstore.go**
```go
package store

import (
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/repository"
)

// AppStore struct integrates repository management.
type AppStore struct {
	Database     *sqlx.DB
	Repositories map[reflect.Type]interface{}
}

// NewAppStore initializes a new AppStore.
func NewAppStore(database *sqlx.DB) *AppStore {
	return &AppStore{
		Database:     database,
		Repositories: make(map[reflect.Type]interface{}),
	}
}

// RegisterRepo adds a new repository to the store.
func (store *AppStore) RegisterRepo(repo interface{}) {
	repoType := reflect.TypeOf(repo).Elem()
	store.Repositories[repoType] = repo
}

// GetRepo retrieves a repository by type.
func (store *AppStore) GetRepo[T any]() (T, error) {
	repoType := reflect.TypeOf((*T)(nil)).Elem()
	if repo, exists := store.Repositories[repoType]; exists {
		return repo.(T), nil
	}
	var zero T
	return zero, fmt.Errorf("repository not registered: %s", repoType.Name())
}
```

### Step 5: ViewRenderer and Example Usage

**viewrenderer/viewrenderer.go**
```go
package viewrenderer

import (
	"net/http"

	"github.com/vert-pjoubert/goth-template/store"
	"github.com/vert-pjoubert/goth-template/repository/models"
)

// ViewRenderer handles rendering views with access to AppStore.
type ViewRenderer struct {
	AppStore *store.AppStore
}

// NewViewRenderer creates a new ViewRenderer.
func NewViewRenderer(appStore *store.AppStore) *ViewRenderer {
	return &ViewRenderer{
		AppStore: appStore,
	}
}

// ProductDetailsHandler handles displaying product details.
func (vr *ViewRenderer) ProductDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Example product ID
	productID := int64(1)

	// Get the product repository
	productRepo, err := vr.AppStore.GetRepo[repository.ProductRepository]()
	if err != nil {
		http.Error(w, "Error getting product repository", http.StatusInternalServerError)
		return
	}

	// Fetch the product by ID
	product, err := productRepo.GetProductByID(productID)
	if err != nil {
		http.Error(w, "Error fetching product", http.StatusInternalServerError)
		return
	}

	// Render the product details (simplified)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
```

### Step 6: Main Application Setup

**main.go**
```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/repository"
	"github.com/vert-pjoubert/goth-template/store"
	"github.com/vert-pjoubert/goth-template/viewrenderer"
	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	database, err := sqlx.Connect("postgres", "user=youruser dbname=yourdb sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize app store
	appStore := store.NewAppStore(database)

	// Register product repository
	productRepo := repository.NewSQLProductRepository(database)
	appStore.RegisterRepo(productRepo)

	// Initialize ViewRenderer
	viewRenderer := viewrenderer.NewViewRenderer(appStore)

	// Set up routes
	router := mux.NewRouter()
	router.HandleFunc("/product/details", viewRenderer.ProductDetailsHandler).Methods("GET")

	// Start server
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
```

### Step-by-Step Instructions

1. **Define Your Product Model**: Create the `Product` model in `/repository/models`.
2. **Create Repository Interface**: Define the `ProductRepository` interface in `/repository`.
3. **Implement Repository Methods**: Implement the `SQLProductRepository` in `/repository`.
4. **AppStore with Repository Registration**: Implement `AppStore` to register and retrieve repositories.
5. **ViewRenderer with ProductDetailsHandler**: Implement `ViewRenderer` to handle product detail rendering.
6. **Main Application Setup**: Demonstrate how to set up the application, register repositories, and handle routes.

---