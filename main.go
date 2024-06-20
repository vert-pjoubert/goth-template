package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

func initDB(config map[string]string) *xorm.Engine {
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

	return engine
}

func main() {
	config, err := LoadEnvConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load environment config: %v", err)
	}
	staticDir := "static"
	engine := initDB(config)

	authKey := os.Getenv("SESSION_AUTH_KEY")
	encKey := os.Getenv("SESSION_ENC_KEY")

	sessionManager, _ := auth.NewCookieSessionManager(authKey, encKey)
	dbStore := store.NewXormDbStore(engine)
	appStore := store.NewCachedAppStore(dbStore, sessionManager)

	authenticator, err := auth.NewOAuth2Authenticator(config, sessionManager, appStore)
	if err != nil {
		log.Fatalf("Failed to create OAuth2Authenticator: %v", err)
	}
	renderer := NewTemplRenderer()
	viewRenderer := NewViewRenderer(appStore)
	h := NewHandlers(authenticator, renderer, viewRenderer, sessionManager)

	http.Handle("/static/", http.StripPrefix("/static/", secureFileServer(http.Dir(staticDir))))
	http.HandleFunc("/", h.IndexHandler)
	http.HandleFunc("/view", authMiddleware(authenticator, viewRenderer.RenderView))
	http.HandleFunc("/layout", h.LayoutHandler)
	http.HandleFunc("/change-theme", h.ChangeThemeHandler)
	http.HandleFunc("/login", authenticator.LoginHandler)
	http.HandleFunc("/logout", authenticator.LogoutHandler)
	http.HandleFunc("/oauth2/callback", authenticator.CallbackHandler)

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
