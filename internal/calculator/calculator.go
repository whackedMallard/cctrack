package calculator

type TokenUsage struct {
	InputTokens      int64
	OutputTokens     int64
	CacheReadTokens  int64
	CacheWriteTokens int64
}

type CostBreakdown struct {
	InputCost      float64
	OutputCost     float64
	CacheReadCost  float64
	CacheWriteCost float64
	TotalCost      float64
}

func Calculate(model string, usage TokenUsage) CostBreakdown {
	rates := GetRates(model)
	cb := CostBreakdown{
		InputCost:      float64(usage.InputTokens) / 1_000_000 * rates.InputPerMToken,
		OutputCost:     float64(usage.OutputTokens) / 1_000_000 * rates.OutputPerMToken,
		CacheReadCost:  float64(usage.CacheReadTokens) / 1_000_000 * rates.CacheReadPerMToken,
		CacheWriteCost: float64(usage.CacheWriteTokens) / 1_000_000 * rates.CacheWritePerMToken,
	}
	cb.TotalCost = cb.InputCost + cb.OutputCost + cb.CacheReadCost + cb.CacheWriteCost
	return cb
}

func (u TokenUsage) Total() int64 {
	return u.InputTokens + u.OutputTokens + u.CacheReadTokens + u.CacheWriteTokens
}
