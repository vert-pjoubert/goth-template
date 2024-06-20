### README.md

# OAuth2Authenticator

The `OAuth2Authenticator` is a Go package that implements OAuth2 and OpenID Connect authentication for web applications. It provides methods for handling login, callback, and logout processes, as well as checking if a user is authenticated. The package uses session management and token caching to track user sessions and authorization levels efficiently.

## Features

- OAuth2 and OpenID Connect authentication
- Session management using Gorilla sessions
- Token caching with LRU (Least Recently Used) cache
- Secure random string generation for state parameters
- Handles token validation and refresh logic

## Usage

### Interface Overview

The `OAuth2Authenticator` struct implements the `IAuthenticator` interface, which provides the following methods:

- `LoginHandler(http.ResponseWriter, *http.Request)`: Handles the login process by generating a state parameter, saving it in the session, and redirecting the user to the OAuth2 provider's authorization URL.

- `CallbackHandler(http.ResponseWriter, *http.Request)`: Handles the callback from the OAuth2 provider, exchanges the authorization code for tokens, validates the ID token, and saves the user information and tokens in the session.

- `LogoutHandler(http.ResponseWriter, *http.Request)`: Handles the logout process by invalidating the session and redirecting the user to the OAuth2 provider's logout URL.

- `IsAuthenticated(http.ResponseWriter, *http.Request) (bool, error)`: Checks if the user is authenticated by validating the session and ID token. It also handles token refresh logic if the token is near expiry.

### How It Works

1. **Login Process**:
   - When a user initiates the login process, the `LoginHandler` generates a secure random state parameter and saves it in the session.
   - The user is redirected to the OAuth2 provider's authorization URL with the state parameter.
   
2. **Callback Process**:
   - After the user grants permission, the OAuth2 provider redirects back to your application with an authorization code and the state parameter.
   - The `CallbackHandler` exchanges the authorization code for access and ID tokens.
   - It validates the ID token using the OpenID Connect library and extracts user information (e.g., email).
   - The user information and tokens are saved in the session.
   
3. **Checking Authentication**:
   - The `IsAuthenticated` method retrieves the session and checks if it contains a valid ID token.
   - If the token is valid and not expired, the user is considered authenticated.
   - If the token is near expiry, the method refreshes the token using the refresh token and updates the session.
   
4. **Logout Process**:
   - The `LogoutHandler` invalidates the session and redirects the user to the OAuth2 provider's logout URL.

### Environment Variables

To configure the `OAuth2Authenticator`, you need to set the following environment variables in your `.env` file:

```
OAUTH2_CLIENT_ID=your-client-id
OAUTH2_CLIENT_SECRET=your-client-secret
OAUTH2_ISSUER_URL=https://accounts.google.com
OAUTH2_REDIRECT_URL=https://yourapp.com/oauth2/callback
OAUTH2_AUTH_URL=https://accounts.google.com/o/oauth2/auth
OAUTH2_TOKEN_URL=https://accounts.google.com/o/oauth2/token
OAUTH2_USERINFO_URL=https://www.googleapis.com/oauth2/v3/userinfo
OAUTH2_LOGOUT_URL=https://accounts.google.com/Logout
TOKEN_EXPIRATION_TIME_SECONDS=3600
SESSION_EXPIRATION_SECONDS=7200
```

### Notes

- Ensure that your application is served over HTTPS to secure the OAuth2 flow.
- Configure your OAuth2 provider with the correct redirect URL that matches the `OAUTH2_REDIRECT_URL` environment variable.
- The token cache uses an LRU (Least Recently Used) eviction strategy to efficiently manage memory usage and ensure that old tokens are removed when the cache reaches its maximum size.
