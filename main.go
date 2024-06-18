package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/store"
	"github.com/vert-pjoubert/goth-template/store/models"
	"xorm.io/xorm"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func initDB() *xorm.Engine {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

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
	engine := initDB()
	sessionStore := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	dbStore := store.NewXormDbStore(engine)
	appStore := store.NewCachedAppStore(dbStore, sessionStore)
	authenticator := auth.NewOAuth2Authenticator(sessionStore)
	renderer := NewTemplRenderer()
	viewRenderer := NewViewRenderer(appStore)
	h := NewHandlers(authenticator, renderer, viewRenderer)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", h.IndexHandler)
	http.HandleFunc("/view", authMiddleware(authenticator, viewRenderer.RenderView))
	http.HandleFunc("/layout", h.LayoutHandler)
	http.HandleFunc("/change-theme", h.ChangeThemeHandler)
	http.HandleFunc("/login", authenticator.LoginHandler)
	http.HandleFunc("/logout", authenticator.LogoutHandler)
	http.HandleFunc("/oauth2/callback", authenticator.CallbackHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func authMiddleware(auth auth.Authenticator, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	}
}
