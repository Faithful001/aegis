package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrOAuthProviderNotSupported = errors.New("unsupported oauth provider")
	ErrOAuthCodeExchangeFailed   = errors.New("failed to exchange authorization code")
	ErrOAuthProfileFetchFailed   = errors.New("failed to fetch user profile from provider")
	ErrOAuthAccountAlreadyLinked = errors.New("this account is already linked to another user")
	ErrCannotUnlinkOnlyLogin     = errors.New("cannot unlink your only sign-in method; please set a password first")
)

// Normalized profile data returned by OAuth providers.
type OAuthProfile struct {
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      string
}

// ExchangeGoogleCode exchanges an authorization code with Google for user info.
func ExchangeGoogleCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*OAuthProfile, error) {
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("google oauth client id or secret not configured")
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	// 1. Exchange code for access token
	form := url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthCodeExchangeFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: google token endpoint returned %d: %s", ErrOAuthCodeExchangeFailed, resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode google token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("%w: missing access_token from google", ErrOAuthCodeExchangeFailed)
	}

	// 2. Fetch user profile with access token
	userInfoReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build userinfo request: %w", err)
	}
	userInfoReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := httpClient.Do(userInfoReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthProfileFetchFailed, err)
	}
	defer userResp.Body.Close()

	userBody, err := io.ReadAll(userResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read userinfo response: %w", err)
	}

	if userResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: google userinfo endpoint returned %d: %s", ErrOAuthProfileFetchFailed, userResp.StatusCode, string(userBody))
	}

	var profileResp struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(userBody, &profileResp); err != nil {
		return nil, fmt.Errorf("failed to decode google profile: %w", err)
	}

	if profileResp.Email == "" || profileResp.Sub == "" {
		return nil, errors.New("incomplete google profile data (missing sub or email)")
	}

	return &OAuthProfile{
		ProviderUserID: profileResp.Sub,
		Email:          strings.ToLower(strings.TrimSpace(profileResp.Email)),
		Name:           profileResp.Name,
		AvatarURL:      profileResp.Picture,
	}, nil
}

// ExchangeGitHubCode exchanges an authorization code with GitHub for user info.
func ExchangeGitHubCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*OAuthProfile, error) {
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("github oauth client id or secret not configured")
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	// 1. Exchange code for access token
	payload := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	}
	jsonBody, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthCodeExchangeFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode github token response: %w", err)
	}
	if tokenResp.Error != "" || tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("%w: github error: %s (%s)", ErrOAuthCodeExchangeFailed, tokenResp.Error, tokenResp.ErrorDesc)
	}

	// 2. Fetch user profile with access token
	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build github user request: %w", err)
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userReq.Header.Set("User-Agent", "Aegis-App")
	userReq.Header.Set("Accept", "application/json")

	userResp, err := httpClient.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthProfileFetchFailed, err)
	}
	defer userResp.Body.Close()

	userBody, err := io.ReadAll(userResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read github user response: %w", err)
	}

	var githubUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(userBody, &githubUser); err != nil {
		return nil, fmt.Errorf("failed to decode github user: %w", err)
	}

	userEmail := strings.TrimSpace(githubUser.Email)

	// If email is private on GitHub, fetch from /user/emails
	if userEmail == "" {
		emailsReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
		if err == nil {
			emailsReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
			emailsReq.Header.Set("User-Agent", "Aegis-App")
			emailsReq.Header.Set("Accept", "application/json")

			if emailsResp, err := httpClient.Do(emailsReq); err == nil {
				defer emailsResp.Body.Close()
				emailsBody, _ := io.ReadAll(emailsResp.Body)
				var emailsList []struct {
					Email    string `json:"email"`
					Primary  bool   `json:"primary"`
					Verified bool   `json:"verified"`
				}
				if json.Unmarshal(emailsBody, &emailsList) == nil {
					for _, em := range emailsList {
						if em.Primary && em.Verified {
							userEmail = em.Email
							break
						}
					}
					if userEmail == "" && len(emailsList) > 0 {
						userEmail = emailsList[0].Email
					}
				}
			}
		}
	}

	if userEmail == "" {
		return nil, errors.New("unable to retrieve a valid email from github account")
	}

	userName := strings.TrimSpace(githubUser.Name)
	if userName == "" {
		userName = githubUser.Login
	}

	return &OAuthProfile{
		ProviderUserID: fmt.Sprintf("%d", githubUser.ID),
		Email:          strings.ToLower(userEmail),
		Name:           userName,
		AvatarURL:      githubUser.AvatarURL,
	}, nil
}
