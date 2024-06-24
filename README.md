### Go Templ + HTMX Website Template Documentation

This documentation provides an overview of the architecture and functionalities of the Go Templ + HTMX website template. It covers how we render pages, handle authentication, manage theme switching, and integrate with OAuth2 and SQLX for database interactions. The goal is to give a clear understanding of how the components work together without diving into large code samples.

#### 1. **Overview**

The website template leverages Go Templ for server-side rendering of HTML templates and HTMX for enhancing the user experience with AJAX capabilities. This combination allows for dynamic and interactive web pages while maintaining a clean and efficient server-side rendering process. Additionally, it supports OAuth2 authentication and integrates both XORM and SQLX for database interactions with MySQL and PostgreSQL.

```shell
go run render.go handlers.go main.go config.go interfaces.go
```

#### 2. **Page Rendering**

**Layout Template**

- The layout template (`Layout`) is the main structure of the web pages. It includes placeholders for the header, footer, sidebar, and content areas.
- The `@content` placeholder within the layout template dynamically inserts the main content of the page.
- HTMX attributes (`hx-get`, `hx-trigger`, `hx-swap`) are used within the layout to asynchronously load parts of the page like the header, footer, and sidebar.

**Example:**

- `hx-get` fetches the content for a specific part.
- `hx-trigger` defines when to load the content (e.g., on page load).
- `hx-swap` specifies how to update the content (e.g., `innerHTML`).

**Content Rendering**

- Specific content for each page is rendered using components defined in individual templates (e.g., `Home`, `Settings`).
- These components are passed to the layout template, which wraps them within the standard page structure.

#### 3. **Authentication**

**OpenID Connect Integration**

- The template supports authentication using Google or any other OAuth2-compliant provider via the OpenID connect protocol.
- The `OAuth2Authenticator` handles the login flow, including redirecting to the provider, handling callbacks, and managing sessions.
- No client-side java script is needed.

**Login Handler**

- The `LoginHandler` initiates the OAuth2 login flow.
- If the user is authenticated, a session cookie is set to manage the user's session.
- Unauthorized users are redirected to the login page.

**Authentication Check**

- The `IsAuthenticated` function checks if a valid session cookie is present.
- If not authenticated, certain routes respond with a `401 Unauthorized` status.

#### 4. **Database Integration**

**DbStore Interface**

- The `DbStore` interface provides basic CRUD operations for interacting with the database.
- The `XormDbStore` implementation uses XORM to manage database interactions.
- The `SQLXDbStore` implementation uses SQLX to manage database interactions.

**AppStore Interface**

- The `AppStore` interface wraps `DbStore` to add caching and application-specific logic.
- The `CachedAppStore` implementation uses an in-memory cache to enhance performance.

**Managing Application State**

- The `AppStore` interface includes methods for managing user sessions and retrieving user data with roles.
- The `GetUserWithRoleByEmail` method retrieves user information along with their role from the database.

#### 5. **Dynamic Content Loading**

**HTMX Integration**

- HTMX is used to dynamically update parts of the page without a full page reload.
- Navigation links and form submissions leverage HTMX to enhance interactivity and responsiveness.

**Handling Unauthorized Responses**

- The template includes an event listener for `htmx:responseError` to handle `401 Unauthorized` responses.
- When an unauthorized response is detected, the user is redirected to the login page.

#### 6. **Theme Switching**

**Settings Template**

- The `Settings` template provides buttons to switch between light and dark themes.
- When a theme button is clicked, a form submission updates the theme preference via the `ChangeThemeHandler`.

**CSS for Themes**

- Separate CSS files (`style-light.css` and `style-dark.css`) define the styles for light and dark themes.
- Conditional CSS classes are used to hide or show the appropriate theme button based on the current theme.

**Example:**

- The light theme shows the dark theme button and hides the light theme button.
- The dark theme shows the light theme button and hides the dark theme button.

#### 7. **Form Handling and AJAX**

**Form Submissions**

- Forms, like the theme change form, use standard POST submissions to update preferences.
- HTMX attributes are used to enhance forms where asynchronous behavior is needed.

**Event Listeners**

- Event listeners, such as for unauthorized responses, are added to handle specific scenarios dynamically.
- These listeners enhance user experience by providing immediate feedback and actions without requiring full page reloads.
