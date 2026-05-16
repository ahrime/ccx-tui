package client

import "time"

type ChannelType string

const (
	ChannelTypeMessages  ChannelType = "messages"
	ChannelTypeResponses ChannelType = "responses"
	ChannelTypeChat      ChannelType = "chat"
	ChannelTypeGemini    ChannelType = "gemini"
)

var AllChannelTypes = []ChannelType{ChannelTypeMessages, ChannelTypeResponses, ChannelTypeChat, ChannelTypeGemini}

type UpstreamConfig struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	BaseURL           string            `json:"baseUrl"`
	BaseURLs          []string          `json:"baseUrls"`
	APIKeys           []string          `json:"apiKeys"`
	HistoricalAPIKeys []string          `json:"historicalApiKeys"`
	ServiceType       string            `json:"serviceType"`
	Status            string            `json:"status"`
	Priority          int               `json:"priority"`
	PromotionUntil    *time.Time        `json:"promotionUntil"`
	ModelMapping      map[string]string `json:"modelMapping"`
	SupportedModels   []string          `json:"supportedModels"`
	Description       string            `json:"description"`
	ProxyURL          string            `json:"proxyUrl"`
	CustomHeaders     map[string]string `json:"customHeaders"`
	ReasoningMapping  map[string]string `json:"reasoningMapping"`
	TestModels        []string          `json:"testModels"`
}

type ChannelMetrics struct {
	ChannelID     string  `json:"channelId"`
	TotalRequests int64   `json:"totalRequests"`
	SuccessCount  int64   `json:"successCount"`
	FailCount     int64   `json:"failCount"`
	SuccessRate   float64 `json:"successRate"`
	AvgLatencyMs  float64 `json:"avgLatencyMs"`
}

type GlobalStats struct {
	TotalRequests  int64 `json:"totalRequests"`
	TotalSuccess   int64 `json:"totalSuccess"`
	TotalFail      int64 `json:"totalFail"`
	ActiveChannels int   `json:"activeChannels"`
	TotalChannels  int   `json:"totalChannels"`
}

type ChannelStatusUpdate struct {
	Status string `json:"status"`
}

type PromotionUpdate struct {
	PromotionUntil *time.Time `json:"promotionUntil"`
}

type ReorderRequest struct {
	Order []string `json:"order"`
}

type APIKeyRequest struct {
	APIKey string `json:"apiKey"`
}
