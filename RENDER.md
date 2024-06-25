### How to Create, Register, and Use Views with `Templ`

### Introduction

This guide will walk you through the steps to create, register, and use views in your application using the `ViewRenderer` and `AppStore` components of this framework, and how to render templates using the `templ` renderer.

### Step 1: Set Up Your Application

First, ensure you have your application set up with the necessary dependencies, including the `ViewRenderer`, `AppStore`, and necessary routes.

```go
package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/store"
	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	database, err := sqlx.Connect("postgres", "user=youruser dbname=yourdb sslmode=disable")
	if err != nil {
		panic(err)
	}

	// Initialize session manager
	sessionManager := auth.NewSessionManager([]byte("your-secret-key"))

	// Initialize app store
	appStore := store.NewAppStore(database, sessionManager)

	// Initialize view renderer
	viewRenderer := NewViewRenderer(appStore)

	// Register views
	viewRenderer.RegisterView("dashboard", DashboardViewHandler, []string{"admin", "user"}, []string{"view_dashboard"})
	viewRenderer.RegisterView("settings", SettingsViewHandler, []string{"admin"}, []string{"edit_settings"})

	// Set up routes
	router := mux.NewRouter()
	router.HandleFunc("/view", viewRenderer.RenderView).Methods("GET")

	// Start server
	http.ListenAndServe(":8080", router)
}
```

### Step 2: Define View Handlers

Create view handlers that will generate the content to be rendered using the `templ` template renderer.

```go
package main

import (
	"net/http"
	"github.com/vert-pjoubert/goth-template/models"
	"github.com/vert-pjoubert/goth-template/templates"
	"github.com/a-h/templ"
)

// TemplRenderer struct for rendering templates with a layout
type TemplRenderer struct{}

// NewTemplRenderer creates a new TemplRenderer
func NewTemplRenderer() *TemplRenderer {
	return &TemplRenderer{}
}

// RenderWithLayout renders the content with the layout
func (r *TemplRenderer) RenderWithLayout(w http.ResponseWriter, content templ.Component, req *http.Request) {
	theme := getTheme(req)
	layout := templates.Layout(content, theme)
	layout.Render(context.Background(), w)
}

// getTheme retrieves the theme from the request cookies
func getTheme(r *http.Request) string {
	cookie, err := r.Cookie("theme")
	if err != nil {
		return "light" // default theme
	}
	return cookie.Value
}

// DashboardViewHandler handles the dashboard view
func DashboardViewHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	renderer := NewTemplRenderer()
	content := templates.Dashboard(user)
	renderer.RenderWithLayout(w, content, r)
}

// SettingsViewHandler handles the settings view
func SettingsViewHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	renderer := NewTemplRenderer()
	content := templates.Settings(user)
	renderer.RenderWithLayout(w, content, r)
}
```

### Step 3: Create Templates

Define the templates using `templ` for your views.

```go
package templates

import (
	"context"
	"net/http"
	"github.com/a-h/templ"
	"github.com/vert-pjoubert/goth-template/models"
)

// Layout defines the layout template
func Layout(content templ.Component, theme string) templ.Component {
	return templ.Div(
		templ.If(theme == "dark",
			templ.Attr("class", "dark-theme"),
		),
		templ.Attr("class", "main-layout"),
		content,
	)
}

// Dashboard generates the dashboard content
func Dashboard(user *models.User) templ.Component {
	return templ.H1(
		templ.Text("Welcome to the Dashboard, "+user.Name),
	)
}

// Settings generates the settings content
func Settings(user *models.User) templ.Component {
	return templ.H1(
		templ.Text("Settings Page for "+user.Name),
	)
}
```

### Step 4: Register Views

Register your views with the `ViewRenderer`.

**Step-by-Step Instructions:**

1. **Initialize the ViewRenderer**:
   ```go
   viewRenderer := NewViewRenderer(appStore)
   ```

2. **Register the Dashboard View**:
   ```go
   viewRenderer.RegisterView(
       "dashboard",                      // View name
       DashboardViewHandler,             // Handler function
       []string{"admin", "user"},        // Required roles
       []string{"view_dashboard"},       // Required permissions
   )
   ```

3. **Register the Settings View**:
   ```go
   viewRenderer.RegisterView(
       "settings",                       // View name
       SettingsViewHandler,              // Handler function
       []string{"admin"},                // Required roles
       []string{"edit_settings"},        // Required permissions
   )
   ```

4. **Complete the Setup**:
   ```go
   func main() {
       // ... previous setup code ...

       // Initialize view renderer
       viewRenderer := NewViewRenderer(appStore)

       // Register views
       viewRenderer.RegisterView("dashboard", DashboardViewHandler, []string{"admin", "user"}, []string{"view_dashboard"})
       viewRenderer.RegisterView("settings", SettingsViewHandler, []string{"admin"}, []string{"edit_settings"})

       // Set up routes
       router := mux.NewRouter()
       router.HandleFunc("/view", viewRenderer.RenderView).Methods("GET")

       // Start server
       http.ListenAndServe(":8080", router)
   }
   ```

### Step 5: Set Up Routes

Configure the router to handle requests to the registered views.

```go
func main() {
	// ... previous setup code ...

	// Set up routes
	router := mux.NewRouter()
	router.HandleFunc("/view", viewRenderer.RenderView).Methods("GET")

	// Start server
	http.ListenAndServe(":8080", router)
}
```

### Step 6: Start the Server

Launch the HTTP server to serve your application.

```go
func main() {
	// ... previous setup code ...

	// Start server
	http.ListenAndServe(":8080", router)
}
```

### Example

Visit `http://localhost:8080/view?view=dashboard` to see the dashboard view.

Visit `http://localhost:8080/view?view=settings` to see the settings view.

Make sure you are logged in and have the necessary roles and permissions to access these views.

---

### Summary

- **Set Up Your Application**: Initialize the database, session manager, and app store.
- **Define View Handlers**: Create handlers that generate the content for each view using `templ`.
- **Create Templates**: Define templates using `templ`.
- **Register Views**: Use the `ViewRenderer` to register views with their handlers, required roles, and permissions.
- **Set Up Routes**: Configure the router to handle view requests.
- **Start the Server**: Launch the HTTP server to serve your application.