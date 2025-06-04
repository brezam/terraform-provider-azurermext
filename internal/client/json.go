package client

// CosmosDBResponse

type CosmosDBResponse struct {
	ID         string              `json:"id"`
	Properties *CosmosDBProperties `json:"properties"`
}

type CosmosDBProperties struct {
	IpRules             []CosmosDBIpRule            `json:"ipRules"`
	PublicNetworkAccess cosmosDBPublicNetworkAccess `json:"publicNetworkAccess"`
}

type CosmosDBIpRule struct {
	IpAddressOrRange string `json:"ipAddressOrRange"`
}

type cosmosDBPublicNetworkAccess string

const (
	cosmosDBPublicNetworkAccessEnabled  cosmosDBPublicNetworkAccess = "Enabled"
	cosmosDBPublicNetworkAccessDisabled cosmosDBPublicNetworkAccess = "Disabled"
)

func (a cosmosDBPublicNetworkAccess) IsEnabled() bool {
	return a == cosmosDBPublicNetworkAccessEnabled
}

// PollResponse

type PollResponse struct {
	Status PollResponseStatus `json:"status"`
}

type PollResponseStatus string

const (
	PollResponseStatusSucceeded  PollResponseStatus = "Succeeded"
	PollResponseStatusFailed     PollResponseStatus = "Failed"
	PollResponseStatusInProgress PollResponseStatus = "InProgress"
	PollResponseStatusEnqueued   PollResponseStatus = "Enqueued"
	PollResponseStatusDequeued   PollResponseStatus = "Dequeued"
)

func (s PollResponseStatus) IsPending() bool {
	return s == PollResponseStatusInProgress || s == PollResponseStatusEnqueued || s == PollResponseStatusDequeued
}

func (s PollResponseStatus) IsSuccess() bool {
	return s == PollResponseStatusSucceeded
}
