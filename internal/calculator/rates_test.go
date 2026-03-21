package calculator

import (
	"math"
	"testing"
)

// almostEqual checks float64 equality within a small epsilon to avoid
// floating-point comparison issues in cost calculations.
func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// ---------- Rate matching per model family ----------

func TestGetRates_Opus46(t *testing.T) {
	// Exact alias
	r := GetRates("claude-opus-4-6")
	assertRates(t, r, "claude-opus-4-6", 5.00, 25.00, 0.50, 6.25)

	// Dated slug — the bug that motivated this fix
	r = GetRates("claude-opus-4-6-20250514")
	assertRates(t, r, "claude-opus-4-6", 5.00, 25.00, 0.50, 6.25)
}

func TestGetRates_Opus45(t *testing.T) {
	r := GetRates("claude-opus-4-5")
	assertRates(t, r, "claude-opus-4-5", 5.00, 25.00, 0.50, 6.25)

	r = GetRates("claude-opus-4-5-20251101")
	assertRates(t, r, "claude-opus-4-5", 5.00, 25.00, 0.50, 6.25)
}

func TestGetRates_Opus41(t *testing.T) {
	r := GetRates("claude-opus-4-1")
	assertRates(t, r, "claude-opus-4-1", 15.00, 75.00, 1.50, 18.75)

	r = GetRates("claude-opus-4-1-20250805")
	assertRates(t, r, "claude-opus-4-1", 15.00, 75.00, 1.50, 18.75)
}

func TestGetRates_Opus40(t *testing.T) {
	// Alias "claude-opus-4-0" starts with "claude-opus-4" prefix
	r := GetRates("claude-opus-4-0")
	assertRates(t, r, "claude-opus-4", 15.00, 75.00, 1.50, 18.75)

	// Dated slug
	r = GetRates("claude-opus-4-20250514")
	assertRates(t, r, "claude-opus-4", 15.00, 75.00, 1.50, 18.75)
}

func TestGetRates_Sonnet46(t *testing.T) {
	r := GetRates("claude-sonnet-4-6")
	assertRates(t, r, "claude-sonnet-4-6", 3.00, 15.00, 0.30, 3.75)

	r = GetRates("claude-sonnet-4-6-20260101")
	assertRates(t, r, "claude-sonnet-4-6", 3.00, 15.00, 0.30, 3.75)
}

func TestGetRates_Sonnet45(t *testing.T) {
	r := GetRates("claude-sonnet-4-5")
	assertRates(t, r, "claude-sonnet-4-5", 3.00, 15.00, 0.30, 3.75)

	r = GetRates("claude-sonnet-4-5-20250929")
	assertRates(t, r, "claude-sonnet-4-5", 3.00, 15.00, 0.30, 3.75)
}

func TestGetRates_Sonnet40(t *testing.T) {
	r := GetRates("claude-sonnet-4-0")
	assertRates(t, r, "claude-sonnet-4", 3.00, 15.00, 0.30, 3.75)

	r = GetRates("claude-sonnet-4-20250514")
	assertRates(t, r, "claude-sonnet-4", 3.00, 15.00, 0.30, 3.75)
}

func TestGetRates_Haiku45(t *testing.T) {
	r := GetRates("claude-haiku-4-5")
	assertRates(t, r, "claude-haiku-4-5", 1.00, 5.00, 0.10, 1.25)

	r = GetRates("claude-haiku-4-5-20251001")
	assertRates(t, r, "claude-haiku-4-5", 1.00, 5.00, 0.10, 1.25)
}

func TestGetRates_Claude35Sonnet(t *testing.T) {
	r := GetRates("claude-3-5-sonnet-20241022")
	assertRates(t, r, "claude-3-5-sonnet", 3.00, 15.00, 0.30, 3.75)

	r = GetRates("claude-3-5-sonnet-latest")
	assertRates(t, r, "claude-3-5-sonnet", 3.00, 15.00, 0.30, 3.75)
}

func TestGetRates_Claude35Haiku(t *testing.T) {
	r := GetRates("claude-3-5-haiku-20241022")
	assertRates(t, r, "claude-3-5-haiku", 0.80, 4.00, 0.08, 1.00)

	r = GetRates("claude-3-5-haiku-latest")
	assertRates(t, r, "claude-3-5-haiku", 0.80, 4.00, 0.08, 1.00)
}

func TestGetRates_Claude3Opus(t *testing.T) {
	r := GetRates("claude-3-opus-20240229")
	assertRates(t, r, "claude-3-opus", 15.00, 75.00, 1.50, 18.75)
}

func TestGetRates_Claude3Haiku(t *testing.T) {
	r := GetRates("claude-3-haiku-20240307")
	assertRates(t, r, "claude-3-haiku", 0.25, 1.25, 0.03, 0.30)
}

// ---------- Ordering: specific prefixes beat general ones ----------

func TestGetRates_OrderingOpus46BeforeOpus4(t *testing.T) {
	// This is THE bug: "claude-opus-4-6-20250514" must get $5/$25 (Opus 4.6),
	// NOT $15/$75 (Opus 4.0). The old 3-entry table matched "claude-opus-4"
	// for all Opus models regardless of version.
	r := GetRates("claude-opus-4-6-20250514")
	if r.InputPerMToken != 5.00 {
		t.Fatalf("claude-opus-4-6-20250514: want input $5.00, got $%.2f (matched family %q)", r.InputPerMToken, r.Family)
	}
	if r.OutputPerMToken != 25.00 {
		t.Fatalf("claude-opus-4-6-20250514: want output $25.00, got $%.2f (matched family %q)", r.OutputPerMToken, r.Family)
	}
}

func TestGetRates_OrderingOpus45BeforeOpus4(t *testing.T) {
	r := GetRates("claude-opus-4-5-20251101")
	if r.InputPerMToken != 5.00 {
		t.Fatalf("claude-opus-4-5-20251101: want input $5.00, got $%.2f (matched family %q)", r.InputPerMToken, r.Family)
	}
}

func TestGetRates_OrderingOpus41BeforeOpus4(t *testing.T) {
	r := GetRates("claude-opus-4-1-20250805")
	if r.InputPerMToken != 15.00 {
		t.Fatalf("claude-opus-4-1-20250805: want input $15.00, got $%.2f (matched family %q)", r.InputPerMToken, r.Family)
	}
}

func TestGetRates_OrderingSonnet46BeforeSonnet4(t *testing.T) {
	r := GetRates("claude-sonnet-4-6-20260101")
	if r.Family != "claude-sonnet-4-6" {
		t.Fatalf("claude-sonnet-4-6-20260101: want family claude-sonnet-4-6, got %q", r.Family)
	}
}

// ---------- Unknown model fallback ----------

func TestGetRates_UnknownModelFallsBackToSonnet(t *testing.T) {
	r := GetRates("some-unknown-model-v2")
	// Fallback should return sonnet 4.0 rates ($3/$15)
	if r.InputPerMToken != 3.00 {
		t.Fatalf("unknown model: want fallback input $3.00, got $%.2f", r.InputPerMToken)
	}
	if r.OutputPerMToken != 15.00 {
		t.Fatalf("unknown model: want fallback output $15.00, got $%.2f", r.OutputPerMToken)
	}
}

func TestGetRates_EmptyStringFallsBackToSonnet(t *testing.T) {
	r := GetRates("")
	if r.InputPerMToken != 3.00 {
		t.Fatalf("empty model: want fallback input $3.00, got $%.2f", r.InputPerMToken)
	}
}

// ---------- Calculate() cost accuracy ----------

func TestCalculate_Opus46_KnownTokens(t *testing.T) {
	// 1M input, 500k output, 2M cache read, 100k cache write
	cb := Calculate("claude-opus-4-6-20250514", TokenUsage{
		InputTokens:      1_000_000,
		OutputTokens:     500_000,
		CacheReadTokens:  2_000_000,
		CacheWriteTokens: 100_000,
	})
	// Expected: input=1*5=5, output=0.5*25=12.5, cacheRead=2*0.5=1, cacheWrite=0.1*6.25=0.625
	wantInput := 5.00
	wantOutput := 12.50
	wantCacheRead := 1.00
	wantCacheWrite := 0.625
	wantTotal := wantInput + wantOutput + wantCacheRead + wantCacheWrite

	if !almostEqual(cb.InputCost, wantInput) {
		t.Errorf("InputCost: want %.4f, got %.4f", wantInput, cb.InputCost)
	}
	if !almostEqual(cb.OutputCost, wantOutput) {
		t.Errorf("OutputCost: want %.4f, got %.4f", wantOutput, cb.OutputCost)
	}
	if !almostEqual(cb.CacheReadCost, wantCacheRead) {
		t.Errorf("CacheReadCost: want %.4f, got %.4f", wantCacheRead, cb.CacheReadCost)
	}
	if !almostEqual(cb.CacheWriteCost, wantCacheWrite) {
		t.Errorf("CacheWriteCost: want %.4f, got %.4f", wantCacheWrite, cb.CacheWriteCost)
	}
	if !almostEqual(cb.TotalCost, wantTotal) {
		t.Errorf("TotalCost: want %.4f, got %.4f", wantTotal, cb.TotalCost)
	}
}

func TestCalculate_Sonnet40_KnownTokens(t *testing.T) {
	cb := Calculate("claude-sonnet-4-20250514", TokenUsage{
		InputTokens:      2_000_000,
		OutputTokens:     1_000_000,
		CacheReadTokens:  500_000,
		CacheWriteTokens: 200_000,
	})
	// Expected: input=2*3=6, output=1*15=15, cacheRead=0.5*0.30=0.15, cacheWrite=0.2*3.75=0.75
	wantTotal := 6.00 + 15.00 + 0.15 + 0.75
	if !almostEqual(cb.TotalCost, wantTotal) {
		t.Errorf("TotalCost: want %.4f, got %.4f", wantTotal, cb.TotalCost)
	}
}

func TestCalculate_Haiku45_KnownTokens(t *testing.T) {
	cb := Calculate("claude-haiku-4-5-20251001", TokenUsage{
		InputTokens:  10_000_000,
		OutputTokens: 5_000_000,
	})
	// Expected: input=10*1=10, output=5*5=25, cacheRead=0, cacheWrite=0
	wantTotal := 35.00
	if !almostEqual(cb.TotalCost, wantTotal) {
		t.Errorf("TotalCost: want %.4f, got %.4f", wantTotal, cb.TotalCost)
	}
}

func TestCalculate_ZeroTokens(t *testing.T) {
	cb := Calculate("claude-opus-4-6", TokenUsage{})
	if cb.TotalCost != 0 {
		t.Errorf("zero tokens: want total 0, got %.4f", cb.TotalCost)
	}
}

// ---------- Table-driven test covering every rate entry ----------

func TestGetRates_AllFamilies(t *testing.T) {
	tests := []struct {
		model           string
		wantFamily      string
		wantInput       float64
		wantOutput      float64
		wantCacheRead   float64
		wantCacheWrite  float64
	}{
		{"claude-opus-4-6", "claude-opus-4-6", 5.00, 25.00, 0.50, 6.25},
		{"claude-opus-4-5", "claude-opus-4-5", 5.00, 25.00, 0.50, 6.25},
		{"claude-opus-4-1", "claude-opus-4-1", 15.00, 75.00, 1.50, 18.75},
		{"claude-opus-4-0", "claude-opus-4", 15.00, 75.00, 1.50, 18.75},
		{"claude-sonnet-4-6", "claude-sonnet-4-6", 3.00, 15.00, 0.30, 3.75},
		{"claude-sonnet-4-5", "claude-sonnet-4-5", 3.00, 15.00, 0.30, 3.75},
		{"claude-sonnet-4-0", "claude-sonnet-4", 3.00, 15.00, 0.30, 3.75},
		{"claude-haiku-4-5", "claude-haiku-4-5", 1.00, 5.00, 0.10, 1.25},
		{"claude-3-5-sonnet-20241022", "claude-3-5-sonnet", 3.00, 15.00, 0.30, 3.75},
		{"claude-3-5-haiku-20241022", "claude-3-5-haiku", 0.80, 4.00, 0.08, 1.00},
		{"claude-3-opus-20240229", "claude-3-opus", 15.00, 75.00, 1.50, 18.75},
		{"claude-3-haiku-20240307", "claude-3-haiku", 0.25, 1.25, 0.03, 0.30},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			r := GetRates(tt.model)
			assertRates(t, r, tt.wantFamily, tt.wantInput, tt.wantOutput, tt.wantCacheRead, tt.wantCacheWrite)
		})
	}
}

// ---------- helper ----------

func assertRates(t *testing.T, r *ModelRates, wantFamily string, wantInput, wantOutput, wantCacheRead, wantCacheWrite float64) {
	t.Helper()
	if r.Family != wantFamily {
		t.Errorf("Family: want %q, got %q", wantFamily, r.Family)
	}
	if !almostEqual(r.InputPerMToken, wantInput) {
		t.Errorf("InputPerMToken: want %.2f, got %.2f (family %q)", wantInput, r.InputPerMToken, r.Family)
	}
	if !almostEqual(r.OutputPerMToken, wantOutput) {
		t.Errorf("OutputPerMToken: want %.2f, got %.2f (family %q)", wantOutput, r.OutputPerMToken, r.Family)
	}
	if !almostEqual(r.CacheReadPerMToken, wantCacheRead) {
		t.Errorf("CacheReadPerMToken: want %.2f, got %.2f (family %q)", wantCacheRead, r.CacheReadPerMToken, r.Family)
	}
	if !almostEqual(r.CacheWritePerMToken, wantCacheWrite) {
		t.Errorf("CacheWritePerMToken: want %.2f, got %.2f (family %q)", wantCacheWrite, r.CacheWritePerMToken, r.Family)
	}
}
