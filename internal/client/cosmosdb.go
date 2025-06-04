package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (c *Client) ReadCosmosDB(ctx context.Context, cosmosAccountId string) (_ *CosmosDBResponse, cErr error) {
	url := "https://management.azure.com" + cosmosAccountId + "?api-version=2025-04-15"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer captureErr(&cErr, resp.Body.Close)

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to read CosmosDB: " + resp.Status + " - " + string(respBody))
	}
	var body CosmosDBResponse
	err = json.Unmarshal(respBody, &body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func (c *Client) UpdateCosmosDBIpRulesAndPoll(ctx context.Context, cosmosAccountId string, ipRules []CosmosDBIpRule) (cErr error) {
	pollUrl := "https://management.azure.com" + cosmosAccountId + "?api-version=2025-04-15"
	body := CosmosDBResponse{
		Properties: &CosmosDBProperties{
			IpRules: ipRules,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PATCH", pollUrl, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	token, err := c.GetToken()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer captureErr(&cErr, resp.Body.Close)
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update CosmosDB IP rules: %s - %s", resp.Status, responseBody)
	}

	pollUrl = resp.Header.Get("Azure-Asyncoperation")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			finished, err := c.getPollUrl(ctx, pollUrl)
			if err != nil {
				return err
			}
			if finished {
				return nil
			}
		}
	}
}

// this function assumes the response is of the format `PollResponse`
func (c *Client) getPollUrl(ctx context.Context, pollUrl string) (finished bool, cErr error) {
	req, err := http.NewRequest("GET", pollUrl, nil)
	if err != nil {
		return false, err
	}
	// It's important we get keep using 'GetToken' in case the previous token expires.
	// The GetToken method already caches it properly so we're not "requesting" it each time.
	token, err := c.GetToken()
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer captureErr(&cErr, resp.Body.Close)
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	var respBody PollResponse
	err = json.Unmarshal(respBodyBytes, &respBody)
	if err != nil {
		return false, err
	}
	if respBody.Status.IsSuccess() {
		return true, nil
	} else if respBody.Status.IsPending() {
		return false, nil
	} else {
		return false, fmt.Errorf("CosmosDB IP rules update status %s, failed: ", respBodyBytes)
	}
}
