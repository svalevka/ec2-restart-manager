package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"


	"ec2-restart-manager/config"
	"ec2-restart-manager/utils"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

var oauthConfig *oauth2.Config
var groupID string

// InitializeAuth sets up the OAuth configuration using the loaded config.
func InitializeAuth(cfg *config.EnvConfig) {
	groupID = cfg.AzureAD.GroupID
	oauthConfig = &oauth2.Config{
		ClientID:     cfg.AzureAD.ClientID,
		ClientSecret: os.Getenv("AZURE_AD_CLIENT_SECRET"), // Ensure this is set as an environment variable
		RedirectURL:  cfg.AzureAD.RedirectURL,
		Endpoint:     microsoft.AzureADEndpoint(cfg.AzureAD.TenantID),
		Scopes:       []string{"openid", "profile", "User.Read", "GroupMember.Read.All"},
	}
}

// Session store for server-side session management (maps session ID to user name)
var SessionStore = make(map[string]string)

func PrintSessionStore() {
    log.Println("Current SessionStore contents:")
    for sessionID, userName := range SessionStore {
        log.Printf("SessionID: %s, UserName: %s\n", sessionID, userName)
    }
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" || SessionStore[cookie.Value] == "" {
			// Redirect to login if the session ID is missing or invalid
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// If the session is valid, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}


// LoginHandler redirects users to the Azure AD login page.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    // Delete the session from SessionStore
    if cookie, err := r.Cookie("session_id"); err == nil {
        delete(SessionStore, cookie.Value)
    }

    // Clear the session_id cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
        Expires:  time.Unix(0, 0),
        HttpOnly: true,
        Secure:   false,
    })

	// Use the current request host to build the redirect URL
	redirectURL := fmt.Sprintf("http://%s", r.Host)

    // Redirect to Azure AD logout URL
    azureLogoutURL := "https://login.microsoftonline.com/common/oauth2/logout"
    http.Redirect(w, r, fmt.Sprintf("%s?post_logout_redirect_uri=%s", azureLogoutURL, redirectURL), http.StatusFound)
}


// CallbackHandler handles the callback from Azure after authentication.
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get the authorization code from query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	// Exchange the code for an access token
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

    // **Check if the user is in the required group**
	if !isUserInGroup(token) {
		// Redirect to access denied page
		http.Redirect(w, r, "/access_denied", http.StatusFound)
		return
	}

	// Use the access token to fetch user profile info from Microsoft Graph
	client := oauthConfig.Client(context.Background(), token)
	userInfo, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil || userInfo.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer userInfo.Body.Close()

	// Decode the user info to get the display name
	var profile struct {
		DisplayName string `json:"displayName"`
	}
	if err := json.NewDecoder(userInfo.Body).Decode(&profile); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Generate a unique session ID and store it with the user's display name
	sessionID := uuid.NewString()
	SessionStore[sessionID] = profile.DisplayName

	if utils.Debug {	
		PrintSessionStore()
	}

	// Set the session ID in a secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production for HTTPS
		Expires:  token.Expiry,
	})

	// Redirect the user to the home page or dashboard
	http.Redirect(w, r, "/", http.StatusFound)
}

// isUserInGroup checks if the user is a member of the specified AD group.
func isUserInGroup(token *oauth2.Token) bool {
	client := oauthConfig.Client(context.Background(), token)
	url := "https://graph.microsoft.com/v1.0/me/memberOf"

	for {
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("Failed to fetch group memberships: %v\n", err)
			return false
		}
		defer resp.Body.Close()

		var groups struct {
			Value []struct {
				ID string `json:"id"`
			} `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
			fmt.Printf("Failed to decode group memberships response: %v\n", err)
			return false
		}

		// Check if the user is a member of the specified group
		for _, group := range groups.Value {
			if group.ID == groupID {
				return true
			}
		}

		// If there's a next link, continue fetching the next page
		if groups.NextLink == "" {
			break
		}
		url = groups.NextLink
	}

	return false
}

// outputUserGroups outputs the list of group IDs the user is a member of to the console.
func outputUserGroups(token *oauth2.Token) {
	client := oauthConfig.Client(context.Background(), token)
	url := "https://graph.microsoft.com/v1.0/me/memberOf"

	fmt.Println("User is a member of the following groups:")

	for {
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("Failed to fetch group memberships: %v\n", err)
			return
		}
		defer resp.Body.Close()

		var groups struct {
			Value []struct {
				ID string `json:"id"`
			} `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
			fmt.Printf("Failed to decode group memberships response: %v\n", err)
			return
		}

		// Output the group IDs to the console
		for _, group := range groups.Value {
			fmt.Printf("Group ID: %s\n", group.ID)
		}

		// If there's a next link, continue fetching the next page
		if groups.NextLink == "" {
			break
		}
		url = groups.NextLink
	}
}


func IsUserLoggedIn(r *http.Request) bool {
    cookie, err := r.Cookie("session_id")
    if err != nil {
        return false // No session cookie found
    }
    sessionID := cookie.Value
    _, loggedIn := SessionStore[sessionID] // Check if session ID exists in the store
    return loggedIn
}