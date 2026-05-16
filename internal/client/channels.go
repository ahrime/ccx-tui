package client

import "context"

func (c *APIClient) ListChannels(ctx context.Context, chType ChannelType) ([]UpstreamConfig, error) {
	var wrapper struct {
		Channels []UpstreamConfig `json:"channels"`
	}
	err := c.Get(ctx, "/api/"+string(chType)+"/channels", &wrapper)
	return wrapper.Channels, err
}

func (c *APIClient) AddChannel(ctx context.Context, chType ChannelType, ch UpstreamConfig) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post(ctx, "/api/"+string(chType)+"/channels", ch, &result)
	return result, err
}

func (c *APIClient) UpdateChannel(ctx context.Context, chType ChannelType, id string, ch UpstreamConfig) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Put(ctx, "/api/"+string(chType)+"/channels/"+id, ch, &result)
	return result, err
}

func (c *APIClient) DeleteChannel(ctx context.Context, chType ChannelType, id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Delete(ctx, "/api/"+string(chType)+"/channels/"+id, &result)
	return result, err
}

func (c *APIClient) SetChannelStatus(ctx context.Context, chType ChannelType, id, status string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Patch(ctx, "/api/"+string(chType)+"/channels/"+id+"/status", ChannelStatusUpdate{Status: status}, &result)
	return result, err
}

func (c *APIClient) ReorderChannels(ctx context.Context, chType ChannelType, order []string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post(ctx, "/api/"+string(chType)+"/channels/reorder", ReorderRequest{Order: order}, &result)
	return result, err
}

func (c *APIClient) AddAPIKey(ctx context.Context, chType ChannelType, id, apiKey string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post(ctx, "/api/"+string(chType)+"/channels/"+id+"/keys", APIKeyRequest{APIKey: apiKey}, &result)
	return result, err
}

func (c *APIClient) DeleteAPIKey(ctx context.Context, chType ChannelType, id, apiKey string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Delete(ctx, "/api/"+string(chType)+"/channels/"+id+"/keys/"+apiKey, &result)
	return result, err
}

func (c *APIClient) MoveKeyToTop(ctx context.Context, chType ChannelType, id, apiKey string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post(ctx, "/api/"+string(chType)+"/channels/"+id+"/keys/"+apiKey+"/top", nil, &result)
	return result, err
}

func (c *APIClient) MoveKeyToBottom(ctx context.Context, chType ChannelType, id, apiKey string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post(ctx, "/api/"+string(chType)+"/channels/"+id+"/keys/"+apiKey+"/bottom", nil, &result)
	return result, err
}

func (c *APIClient) SetPromotion(ctx context.Context, chType ChannelType, id string, promo PromotionUpdate) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post(ctx, "/api/"+string(chType)+"/channels/"+id+"/promotion", promo, &result)
	return result, err
}

func (c *APIClient) ResumeChannel(ctx context.Context, chType ChannelType, id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post(ctx, "/api/"+string(chType)+"/channels/"+id+"/resume", nil, &result)
	return result, err
}
