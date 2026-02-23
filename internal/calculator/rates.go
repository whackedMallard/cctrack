package calculator

type ModelRates struct {
	Family              string
	InputPerMToken      float64
	OutputPerMToken     float64
	CacheReadPerMToken  float64
	CacheWritePerMToken float64
}

// Rates maps model family prefixes to their pricing.
// Models are matched by prefix: "claude-haiku-4-5-20251001" → "claude-haiku-4"
var Rates = []ModelRates{
	{
		Family:              "claude-opus-4",
		InputPerMToken:      15.00,
		OutputPerMToken:     75.00,
		CacheReadPerMToken:  1.50,
		CacheWritePerMToken: 18.75,
	},
	{
		Family:              "claude-sonnet-4",
		InputPerMToken:      3.00,
		OutputPerMToken:     15.00,
		CacheReadPerMToken:  0.30,
		CacheWritePerMToken: 3.75,
	},
	{
		Family:              "claude-haiku-4",
		InputPerMToken:      0.80,
		OutputPerMToken:     4.00,
		CacheReadPerMToken:  0.08,
		CacheWritePerMToken: 1.00,
	},
}

func GetRates(model string) *ModelRates {
	for i := range Rates {
		if len(model) >= len(Rates[i].Family) && model[:len(Rates[i].Family)] == Rates[i].Family {
			return &Rates[i]
		}
	}
	// Fallback: try matching shorter prefixes for unknown models
	// Default to sonnet rates as the most common
	return &Rates[1]
}
