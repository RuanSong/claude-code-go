package oauth

import (
	"net/url"
	"testing"
	"time"
)

func TestNewOAuth(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AuthURL:      "https://auth.example.com",
		TokenURL:     "https://auth.example.com/token",
		RedirectURI:  "http://localhost:8080/callback",
		Scopes:       []string{"read", "write"},
	}

	oauth := NewOAuth(config)

	if oauth == nil {
		t.Fatal("NewOAuth() returned nil")
	}

	if oauth.config != config {
		t.Error("NewOAuth() did not set config correctly")
	}

	if oauth.httpClient == nil {
		t.Error("NewOAuth() did not set httpClient")
	}

	if oauth.token != nil {
		t.Error("NewOAuth() should not have a token initially")
	}
}

func TestOAuth_BuildAuthURL(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AuthURL:      "https://auth.example.com/authorize",
		TokenURL:     "https://auth.example.com/token",
		RedirectURI:  "http://localhost:8080/callback",
		Scopes:       []string{"read", "write"},
	}

	oauth := NewOAuth(config)

	authURL, err := oauth.BuildAuthURL("test-state")
	if err != nil {
		t.Fatalf("BuildAuthURL() error = %v", err)
	}

	if authURL == "" {
		t.Fatal("BuildAuthURL() returned empty URL")
	}

	// Parse the URL and verify components
	parsed, err := url.ParseRequestURI(authURL)
	if err != nil {
		t.Fatalf("BuildAuthURL() returned invalid URL: %v", err)
	}

	if parsed.Query().Get("client_id") != "test-client-id" {
		t.Error("BuildAuthURL() URL missing or incorrect client_id")
	}

	if parsed.Query().Get("redirect_uri") != "http://localhost:8080/callback" {
		t.Error("BuildAuthURL() URL missing or incorrect redirect_uri")
	}

	if parsed.Query().Get("state") != "test-state" {
		t.Error("BuildAuthURL() URL missing or incorrect state")
	}

	if parsed.Query().Get("response_type") != "code" {
		t.Error("BuildAuthURL() URL missing or incorrect response_type")
	}
}

func TestOAuth_BuildAuthURL_NoScopes(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AuthURL:      "https://auth.example.com/authorize",
		TokenURL:     "https://auth.example.com/token",
		RedirectURI:  "http://localhost:8080/callback",
		Scopes:       []string{},
	}

	oauth := NewOAuth(config)

	url, err := oauth.BuildAuthURL("test-state")
	if err != nil {
		t.Fatalf("BuildAuthURL() error = %v", err)
	}

	if url == "" {
		t.Fatal("BuildAuthURL() returned empty URL")
	}
}

func TestOAuth_BuildAuthURL_InvalidAuthURL(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AuthURL:      "://invalid-url",
		TokenURL:     "https://auth.example.com/token",
		RedirectURI:  "http://localhost:8080/callback",
		Scopes:       []string{},
	}

	oauth := NewOAuth(config)

	_, err := oauth.BuildAuthURL("test-state")
	if err == nil {
		t.Error("BuildAuthURL() expected error for invalid URL")
	}
}

func TestOAuth_GetToken(t *testing.T) {
	oauth := NewOAuth(&Config{})

	// Initially no token
	token := oauth.GetToken()
	if token != nil {
		t.Error("GetToken() should return nil when no token is set")
	}

	// Set a token
	testToken := &Token{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}
	oauth.SetToken(testToken)

	// Now should return the token
	retrieved := oauth.GetToken()
	if retrieved == nil {
		t.Fatal("GetToken() returned nil after SetToken()")
	}

	if retrieved.AccessToken != "test-access-token" {
		t.Error("GetToken() returned wrong token")
	}
}

func TestOAuth_IsTokenExpired(t *testing.T) {
	oauth := NewOAuth(&Config{})

	// No token should be expired
	if !oauth.IsTokenExpired() {
		t.Error("IsTokenExpired() should return true when no token is set")
	}

	// Set an unexpired token (expires in 1 hour)
	oauth.SetToken(&Token{
		AccessToken: "test-token",
		ExpiresIn:   3600,
		CreatedAt:   time.Now(),
	})

	if oauth.IsTokenExpired() {
		t.Error("IsTokenExpired() should return false for unexpired token")
	}

	// Set an expired token
	oauth.SetToken(&Token{
		AccessToken: "test-token",
		ExpiresIn:   -1, // Already expired
		CreatedAt:   time.Now().Add(-2 * time.Hour),
	})

	if !oauth.IsTokenExpired() {
		t.Error("IsTokenExpired() should return true for expired token")
	}
}

func TestOAuth_SetToken(t *testing.T) {
	oauth := NewOAuth(&Config{})

	token := &Token{
		AccessToken:  "new-access-token",
		TokenType:    "Bearer",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
	}

	oauth.SetToken(token)

	retrieved := oauth.GetToken()
	if retrieved.AccessToken != "new-access-token" {
		t.Error("SetToken() did not store token correctly")
	}
}

func TestOAuth_ClearToken(t *testing.T) {
	oauth := NewOAuth(&Config{})

	// Set a token
	oauth.SetToken(&Token{
		AccessToken: "test-token",
	})

	// Clear it
	oauth.ClearToken()

	if oauth.GetToken() != nil {
		t.Error("ClearToken() did not clear token")
	}

	if !oauth.IsTokenExpired() {
		t.Error("IsTokenExpired() should return true after clearing token")
	}
}

func TestNewTokenManager(t *testing.T) {
	manager := NewTokenManager()

	if manager == nil {
		t.Fatal("NewTokenManager() returned nil")
	}

	if manager.tokens == nil {
		t.Error("NewTokenManager() did not initialize tokens map")
	}
}

func TestTokenManager_Store(t *testing.T) {
	manager := NewTokenManager()

	token := &Token{
		AccessToken: "test-token",
	}

	manager.Store("provider1", token)

	retrieved, ok := manager.Get("provider1")
	if !ok {
		t.Fatal("Store() did not store token, Get() returned ok = false")
	}

	if retrieved.AccessToken != "test-token" {
		t.Error("Store()/Get() round-trip failed")
	}
}

func TestTokenManager_Get_NotFound(t *testing.T) {
	manager := NewTokenManager()

	_, ok := manager.Get("non-existent")
	if ok {
		t.Error("Get() should return ok = false for non-existent provider")
	}
}

func TestTokenManager_Remove(t *testing.T) {
	manager := NewTokenManager()

	manager.Store("provider1", &Token{AccessToken: "test"})

	manager.Remove("provider1")

	_, ok := manager.Get("provider1")
	if ok {
		t.Error("Remove() did not remove token")
	}
}

func TestTokenManager_ListProviders(t *testing.T) {
	manager := NewTokenManager()

	manager.Store("provider1", &Token{AccessToken: "test1"})
	manager.Store("provider2", &Token{AccessToken: "test2"})
	manager.Store("provider3", &Token{AccessToken: "test3"})

	providers := manager.ListProviders()

	if len(providers) != 3 {
		t.Errorf("ListProviders() returned %d providers, want 3", len(providers))
	}
}

func TestTokenManager_ListProviders_Empty(t *testing.T) {
	manager := NewTokenManager()

	providers := manager.ListProviders()

	if len(providers) != 0 {
		t.Errorf("ListProviders() returned %d providers, want 0", len(providers))
	}
}

func TestTokenManager_ConcurrentAccess(t *testing.T) {
	manager := NewTokenManager()

	done := make(chan bool)

	// Concurrent store
	for i := 0; i < 10; i++ {
		go func(id int) {
			manager.Store(string(rune('0'+id)), &Token{AccessToken: "token"})
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	providers := manager.ListProviders()
	if len(providers) != 10 {
		t.Errorf("ListProviders() after concurrent stores returned %d, want 10", len(providers))
	}
}

func TestToken_Fields(t *testing.T) {
	token := &Token{
		AccessToken:  "access-token-value",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "refresh-token-value",
		Scope:        "read write",
		CreatedAt:    time.Now(),
	}

	if token.AccessToken != "access-token-value" {
		t.Error("Token.AccessToken not set correctly")
	}

	if token.TokenType != "Bearer" {
		t.Error("Token.TokenType not set correctly")
	}

	if token.ExpiresIn != 3600 {
		t.Error("Token.ExpiresIn not set correctly")
	}

	if token.RefreshToken != "refresh-token-value" {
		t.Error("Token.RefreshToken not set correctly")
	}

	if token.Scope != "read write" {
		t.Error("Token.Scope not set correctly")
	}
}

func TestConfig_Fields(t *testing.T) {
	config := &Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		AuthURL:      "https://auth.example.com",
		TokenURL:     "https://auth.example.com/token",
		RedirectURI:  "http://localhost:8080",
		Scopes:       []string{"scope1", "scope2"},
	}

	if config.ClientID != "client-id" {
		t.Error("Config.ClientID not set correctly")
	}

	if config.ClientSecret != "client-secret" {
		t.Error("Config.ClientSecret not set correctly")
	}

	if len(config.Scopes) != 2 {
		t.Errorf("Config.Scopes has %d items, want 2", len(config.Scopes))
	}
}
