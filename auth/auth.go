package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
	"ec2-restart-manager/config"
)

var debug = false

var oauthConfig *oauth2.Config
var groupID string

// InitializeAuth sets up the OAuth configuration using the loaded config.
func InitializeAuth(cfg *config.Config) {
    groupID = cfg.AzureAD.GroupID
    oauthConfig = &oauth2.Config{
        ClientID:     cfg.AzureAD.ClientID,
        ClientSecret: os.Getenv("AZURE_AD_CLIENT_SECRET"), // Ensure this is set as an environment variable
        RedirectURL:  cfg.AzureAD.RedirectURL,
        Endpoint:     microsoft.AzureADEndpoint(cfg.AzureAD.TenantID),
        Scopes:       []string{"openid", "profile", "User.Read", "GroupMember.Read.All"},
    }
}

// AuthMiddleware ensures that requests are authenticated.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil || cookie.Value == "" {
			if debug {
				fmt.Println("Token not found or invalid, redirecting to login")
			}
			
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Create an OAuth2 token object with the value from the cookie
		token := &oauth2.Token{
			AccessToken: cookie.Value,
			TokenType:   "Bearer",
		}

		// Validate the token by making a request to the API
		client := oauthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Printf("Token validation failed or token expired: %v\n", err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		defer resp.Body.Close()
		next.ServeHTTP(w, r) // Call the next handler in the chain if authentication is successful
	})
}

// LoginHandler redirects users to the Azure login page.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

// CallbackHandler handles the callback from Azure after authentication.
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token.AccessToken,
		Path:     "/",
		Expires:  token.Expiry,
		HttpOnly: true,
		Secure:   false, // Set to true in production to ensure it's only sent over HTTPS
	})

	http.Redirect(w, r, "/", http.StatusFound)
}
