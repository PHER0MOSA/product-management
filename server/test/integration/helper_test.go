package integration_test

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/gorilla/mux"

	infrapg "products-manage-server/internal/infrastructure/postgres"
	"products-manage-server/internal/product/application"
	"products-manage-server/internal/product/handler"
	productpg "products-manage-server/internal/product/infrastructure/postgres"
)

type productJSON struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

type errorJSON struct {
	Message string `json:"message"`
}

func TestMain(m *testing.M) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := infrapg.Open(databaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	if err := setupTestDB(db); err != nil {
		log.Fatalf("setup test db: %v", err)
	}

	os.Exit(m.Run())
}

// setupTestDB はスキーマ作成とテスト用シードデータ投入を行う。
// migrate 不要。空の DB からテストできる。
func setupTestDB(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id    BIGSERIAL PRIMARY KEY,
			name  VARCHAR(20) NOT NULL,
			price BIGINT NOT NULL CHECK (price >= 0)
		)
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`TRUNCATE products RESTART IDENTITY`); err != nil {
		return err
	}

	_, err := db.Exec(`
		INSERT INTO products (name, price) VALUES
			('商品A', 1000),
			('商品B', 2000),
			('商品C', 3000)
	`)
	return err
}

func setupRouter(t *testing.T) http.Handler {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := infrapg.Open(databaseURL)
	if err != nil {
		t.Fatalf("database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return newRouter(db)
}

func newRouter(db *sql.DB) http.Handler {
	repo := productpg.NewProductRepository(db)
	svc := application.NewProductService(repo)
	h := handler.NewProductHandler(svc)

	r := mux.NewRouter()
	h.Register(r)
	return r
}
