package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type authToken struct {
	token     string
	expiresAt int64
}

type Client struct {
	lock         sync.Mutex
	authToken    authToken
	clientId     string
	clientSecret string
	tenantId     string
}

func New(clientId, clientSecret, tenantId string) *Client {
	return &Client{sync.Mutex{}, authToken{}, clientId, clientSecret, tenantId}
}

func (c *Client) GetToken() (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if time.Now().Unix()+60 > c.authToken.expiresAt {
		err := c.refreshAuthToken()
		if err != nil {
			return "", err
		}
	}
	return c.authToken.token, nil
}

func (c *Client) refreshAuthToken() (cErr error) {
	reqBody := url.Values{}
	reqBody.Set("grant_type", "client_credentials")
	reqBody.Set("client_id", c.clientId)
	reqBody.Set("client_secret", c.clientSecret)
	reqBody.Set("scope", "https://management.core.windows.net//.default")
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", c.tenantId),
		strings.NewReader(reqBody.Encode()),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer captureErr(&cErr, resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to request token. Response status: %d. Request query params: %s", resp.StatusCode, req.URL.RawQuery)
	}

	var tokenResponse struct {
		Token     string `json:"access_token"`
		ExpiresIn int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return err
	}

	c.authToken.token = tokenResponse.Token
	c.authToken.expiresAt = time.Now().Unix() + tokenResponse.ExpiresIn
	return nil
}
