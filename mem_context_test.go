package goja

import "testing"

func TestMemUsageLimitExceeded(t *testing.T) {
	tests := []struct {
		name     string
		memUsage uint64
		mu       *MemUsageContext
		expected bool
	}{
		{
			name:     "did not exceed returns false",
			memUsage: 12,
			mu: &MemUsageContext{
				memoryLimit: 50,
			},
			expected: false,
		},
		{
			name:     "memory exceeds threshold returns true",
			memUsage: 700,
			mu: &MemUsageContext{
				memoryLimit: 50,
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.mu.MemUsageLimitExceeded(tc.memUsage)
			if actual != tc.expected {
				t.Fatalf("ACTUAL: %v EXPECTED: %v", actual, tc.expected)
			}
		})
	}
}

func TestComputeSampleStep(t *testing.T) {
	tests := []struct {
		name       string
		totalItems int
		sampleRate float64
		expected   int
	}{
		{"should compute sample size given 10 items and 10% rate", 100, 0.1, 10},
		{"should compute sample size given 10 items and 20% rate", 100, 0.2, 5},
		{"should compute sample size given 10 items and 50% rate", 100, 0.5, 2},
		{"should compute sample size of 5 given 10 items and 70% rate", 100, 0.7, 2},
		{"should compute sample size of 5 given 10 items and 100% rate", 100, 1, 2},
		{"should compute sample size of 5 given 10 items and 150% rate", 100, 1.5, 2},
		{"should compute sample size of 1 given 0 items and 50% rate", 0, 0.5, 1},
		{"should compute sample size of 1 given 10 items and 0% rate", 100, 0, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := computeSampleStep(tc.totalItems, tc.sampleRate)
			if actual != tc.expected {
				t.Fatalf("ACTUAL: %v EXPECTED: %v", actual, tc.expected)
			}
		})
	}
}
