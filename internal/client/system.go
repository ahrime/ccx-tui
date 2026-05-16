package client

import "context"

func (c *APIClient) PingChannel(ctx context.Context, chType ChannelType, id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/"+string(chType)+"/ping/"+id, &result)
	return result, err
}

func (c *APIClient) PingAll(ctx context.Context, chType ChannelType) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/"+string(chType)+"/ping", &result)
	return result, err
}

func (c *APIClient) GetDashboard(ctx context.Context, chType ChannelType) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/"+string(chType)+"/channels/dashboard", &result)
	return result, err
}

func (c *APIClient) GetChannelMetrics(ctx context.Context, chType ChannelType) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/"+string(chType)+"/channels/metrics", &result)
	return result, err
}

func (c *APIClient) GetMetricsHistory(ctx context.Context, chType ChannelType) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/"+string(chType)+"/channels/metrics/history", &result)
	return result, err
}

func (c *APIClient) GetGlobalStatsHistory(ctx context.Context, chType ChannelType) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/"+string(chType)+"/global/stats/history", &result)
	return result, err
}

func (c *APIClient) GetModelStatsHistory(ctx context.Context, chType ChannelType) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/"+string(chType)+"/models/stats/history", &result)
	return result, err
}

func (c *APIClient) GetSchedulerStats(ctx context.Context, chType ChannelType) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/"+string(chType)+"/channels/scheduler/stats", &result)
	return result, err
}

func (c *APIClient) GetChannelLogs(ctx context.Context, chType ChannelType, id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/"+string(chType)+"/channels/"+id+"/logs", &result)
	return result, err
}

func (c *APIClient) GetChannelModels(ctx context.Context, chType ChannelType, id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post(ctx, "/api/"+string(chType)+"/channels/"+id+"/models", nil, &result)
	return result, err
}

func (c *APIClient) StartCapabilityTest(ctx context.Context, chType ChannelType, id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post(ctx, "/api/"+string(chType)+"/channels/"+id+"/capability-test", nil, &result)
	return result, err
}

func (c *APIClient) GetCapabilityTestStatus(ctx context.Context, chType ChannelType, id, jobID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/api/"+string(chType)+"/channels/"+id+"/capability-test/"+jobID, &result)
	return result, err
}

func (c *APIClient) CancelCapabilityTest(ctx context.Context, chType ChannelType, id, jobID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Delete(ctx, "/api/"+string(chType)+"/channels/"+id+"/capability-test/"+jobID, &result)
	return result, err
}

func (c *APIClient) RetryCapabilityTest(ctx context.Context, chType ChannelType, id, jobID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post(ctx, "/api/"+string(chType)+"/channels/"+id+"/capability-test/"+jobID+"/retry", nil, &result)
	return result, err
}
