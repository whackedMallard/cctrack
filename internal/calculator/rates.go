package calculator

type ModelRates struct {
	Family              string
	InputPerMToken      float64
	OutputPerMToken     float64
	CacheReadPerMToken  float64
	CacheWritePerMToken float64
}

// Rates maps model family prefixes to their pricing.
// GetRates() does a prefix match and returns the first hit, so more-specific
// prefixes MUST appear before less-specific ones.
// e.g. "claude-opus-4-6" before "claude-opus-4-5" before "claude-opus-4".
//
// Pricing source: https://platform.claude.com/docs/en/about-claude/pricing
// Cache read = 0.1x base input; cache write (5-min) = 1.25x base input.
var Rates = []ModelRates{
	// -- Opus 4.6 ($5 / $25) --
	{Family: "claude-opus-4-6", InputPerMToken: 5.00, OutputPerMToken: 25.00, CacheReadPerMToken: 0.50, CacheWritePerMToken: 6.25},

	// -- Opus 4.5 ($5 / $25) --
	{Family: "claude-opus-4-5", InputPerMToken: 5.00, OutputPerMToken: 25.00, CacheReadPerMToken: 0.50, CacheWritePerMToken: 6.25},

	// -- Opus 4.1 ($15 / $75) --
	{Family: "claude-opus-4-1", InputPerMToken: 15.00, OutputPerMToken: 75.00, CacheReadPerMToken: 1.50, CacheWritePerMToken: 18.75},

	// -- Opus 4.0 ($15 / $75) -- catches "claude-opus-4-0*" and dated slugs like "claude-opus-4-20250514"
	{Family: "claude-opus-4", InputPerMToken: 15.00, OutputPerMToken: 75.00, CacheReadPerMToken: 1.50, CacheWritePerMToken: 18.75},

	// -- Sonnet 4.6 ($3 / $15) --
	{Family: "claude-sonnet-4-6", InputPerMToken: 3.00, OutputPerMToken: 15.00, CacheReadPerMToken: 0.30, CacheWritePerMToken: 3.75},

	// -- Sonnet 4.5 ($3 / $15) --
	{Family: "claude-sonnet-4-5", InputPerMToken: 3.00, OutputPerMToken: 15.00, CacheReadPerMToken: 0.30, CacheWritePerMToken: 3.75},

	// -- Sonnet 4.0 ($3 / $15) -- catches "claude-sonnet-4-0*" and dated slugs like "claude-sonnet-4-20250514"
	{Family: "claude-sonnet-4", InputPerMToken: 3.00, OutputPerMToken: 15.00, CacheReadPerMToken: 0.30, CacheWritePerMToken: 3.75},

	// -- Haiku 4.5 ($1 / $5) --
	{Family: "claude-haiku-4-5", InputPerMToken: 1.00, OutputPerMToken: 5.00, CacheReadPerMToken: 0.10, CacheWritePerMToken: 1.25},

	// -- Claude 3.5 Sonnet ($3 / $15) -- also covers "claude-3-5-sonnet-*" model IDs
	{Family: "claude-3-5-sonnet", InputPerMToken: 3.00, OutputPerMToken: 15.00, CacheReadPerMToken: 0.30, CacheWritePerMToken: 3.75},

	// -- Claude 3.5 Haiku ($0.80 / $4) --
	{Family: "claude-3-5-haiku", InputPerMToken: 0.80, OutputPerMToken: 4.00, CacheReadPerMToken: 0.08, CacheWritePerMToken: 1.00},

	// -- Claude 3 Opus ($15 / $75) --
	{Family: "claude-3-opus", InputPerMToken: 15.00, OutputPerMToken: 75.00, CacheReadPerMToken: 1.50, CacheWritePerMToken: 18.75},

	// -- Claude 3 Haiku ($0.25 / $1.25) --
	{Family: "claude-3-haiku", InputPerMToken: 0.25, OutputPerMToken: 1.25, CacheReadPerMToken: 0.03, CacheWritePerMToken: 0.30},
}

func GetRates(model string) *ModelRates {
	for i := range Rates {
		if len(model) >= len(Rates[i].Family) && model[:len(Rates[i].Family)] == Rates[i].Family {
			return &Rates[i]
		}
	}
	// Fallback: default to sonnet 4.0 rates (index 6) as the most common model
	return &Rates[6]
}
