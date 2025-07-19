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

	"github.com/hashicorp/terraform-plugin-log/tflog"
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
		if resp.StatusCode == http.StatusNotFound {
			return nil, NewNotFoundError(cosmosAccountId)
		}
		return nil, errors.New("failed to read CosmosDB: " + resp.Status + " - " + string(respBody))
	}
	var body CosmosDBResponse
	err = json.Unmarshal(respBody, &body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func (c *Client) UpdateCosmosDBIpRulesAndPoll(ctx context.Context, cosmosAccountId string, rules []string) (cErr error) {
	pollUrl := "https://management.azure.com" + cosmosAccountId + "?api-version=2025-04-15"
	cosmosDBIPRules := make([]CosmosDBIpRule, len(rules))
	for i, ip := range rules {
		cosmosDBIPRules[i] = CosmosDBIpRule{IpAddressOrRange: ip}
	}
	body := CosmosDBResponse{Properties: &CosmosDBProperties{IpRules: cosmosDBIPRules}}
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
	tflog.Info(ctx, fmt.Sprintf("Updating IP rules to: %v", rules))
	tflog.Debug(ctx, "PATCH Request "+pollUrl)
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
	tflog.Debug(ctx, "Async operation url: "+pollUrl)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second): // Since Go 1.23 this isn't a memory leak anymore.
			finished, err := c.poll(ctx, pollUrl)
			if err != nil {
				return err
			}
			if finished {
				return nil
			}
		}
	}
}

func (c *Client) poll(ctx context.Context, pollUrl string) (finished bool, cErr error) {
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
	tflog.Debug(ctx, "Async operation response: "+string(respBody.Status))
	if respBody.Status.IsSuccess() {
		return true, nil
	} else if respBody.Status.IsPending() {
		return false, nil
	} else {
		return false, fmt.Errorf("updating IP failed. Poll status: %s", respBodyBytes)
	}
}
