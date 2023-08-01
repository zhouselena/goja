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
