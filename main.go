package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/store"
	"github.com/vert-pjoubert/goth-template/store/models"
	"xorm.io/xorm"
)

func init() {
	if _, err := LoadEnvConfig(".env"); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}
	gob.Register(time.Time{})

}

func initDB(config map[string]string) store.DbStore {
	dbType := config["DB_TYPE"]

	switch dbType {
	case "sqlx":
		return initSqlxDB(config)
	case "xorm":
		return initXormDB(config)
	default:
		log.Fatalf("Unknown DB_TYPE: %s", dbType)
		return nil
	}
}

func initSqlxDB(config map[string]string) *store.SqlxDbStore {
	dbUser := config["DB_USER"]
	dbPassword := config["DB_PASSWORD"]
	dbHost := config["DB_HOST"]
	dbPort := config["DB_PORT"]
	dbName := config["DB_NAME"]

	dataSourceName := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sqlx.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to create sqlx.DB: %v", err)
	}

	// Ensure the database schema is in sync with the models
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		role_id INT,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS roles (
		id SERIAL PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		description TEXT,
		permissions JSONB
	)`)
	if err != nil {
		log.Fatalf("Failed to create roles table: %v", err)
	}

	return store.NewSqlxDbStore(db)
}

func initXormDB(config map[string]string) *store.XormDbStore {
	dbUser := config["DB_USER"]
	dbPassword := config["DB_PASSWORD"]
	dbHost := config["DB_HOST"]
	dbPort := config["DB_PORT"]
	dbName := config["DB_NAME"]

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	engine, err := xorm.NewEngine("mysql", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to create XORM engine: %v", err)
	}

	err = engine.Sync2(new(models.User), new(models.Role))
	if err != nil {
		log.Fatalf("Failed to sync database schema: %v", err)
	}

	return store.NewXormDbStore(engine)
}

func main() {
	// Load configuration
	config, err := LoadEnvConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load environment config: %v", err)
	}

	// Initialize database
	dbStore := initDB(config)

	// Initialize session manager
	authKey := os.Getenv("SESSION_AUTH_KEY")
	encKey := os.Getenv("SESSION_ENC_KEY")
	sessionManager, _ := auth.NewCookieSessionManager(authKey, encKey)

	// Create cached app store
	appStore := store.NewCachedAppStore(dbStore, sessionManager)

	// Initialize OAuth2 authenticator
	authenticator, err := auth.NewOAuth2Authenticator(config, sessionManager, appStore)
	if err != nil {
		log.Fatalf("Failed to create OAuth2Authenticator: %v", err)
	}

	// Initialize renderers
	renderer := NewTemplRenderer()
	viewRenderer := NewViewRenderer(appStore)

	// Register views
	h := NewHandlers(authenticator, renderer, viewRenderer, sessionManager)
	viewRenderer.RegisterView("settings", h.SettingsViewHandler, []string{"admin", "user"}, []string{"read"})
	viewRenderer.RegisterView("servers", h.ServersViewHandler, []string{"admin"}, []string{"read"})
	viewRenderer.RegisterView("events", h.EventsViewHandler, []string{"admin", "user"}, []string{"read"})
	viewRenderer.RegisterView("protected", h.ProtectedViewHandler, []string{"admin"}, []string{"admin_only"})

	// Set up HTTP routes
	http.Handle("/static/", http.StripPrefix("/static/", secureFileServer(http.Dir("static"))))
	http.HandleFunc("/", h.IndexHandler)
	http.HandleFunc("/view", authMiddleware(authenticator, viewRenderer.RenderView))
	http.HandleFunc("/layout", h.LayoutHandler)
	http.HandleFunc("/change-theme", h.ChangeThemeHandler)
	http.HandleFunc("/login", h.LoginHandler)
	http.HandleFunc("/logout", authenticator.LogoutHandler)
	http.HandleFunc("/oauth2/callback", authenticator.CallbackHandler)

	// Start HTTP server
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func authMiddleware(auth IAuthenticator, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authenticated, err := auth.IsAuthenticated(w, r)
		if err != nil {
			log.Printf("Authentication error: %v", err)
		}
		if !authenticated {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	}
}
