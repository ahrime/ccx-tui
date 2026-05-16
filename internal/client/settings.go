package client

import "context"

func (c *APIClient) GetFuzzyMode(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/settings/fuzzy-mode", &result)
	return result, err
}

func (c *APIClient) SetFuzzyMode(ctx context.Context, enabled bool) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Put(ctx, "/api/settings/fuzzy-mode", map[string]bool{"enabled": enabled}, &result)
	return result, err
}

func (c *APIClient) GetStripBillingHeader(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/settings/strip-billing-header", &result)
	return result, err
}

func (c *APIClient) SetStripBillingHeader(ctx context.Context, enabled bool) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Put(ctx, "/api/settings/strip-billing-header", map[string]bool{"enabled": enabled}, &result)
	return result, err
}
