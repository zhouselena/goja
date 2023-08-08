package goja

import (
	"testing"
)

func TestDestructMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            *destructKeyedSource
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should have a value given by the wrapped value",
			val:            &destructKeyedSource{wrapped: valueInt(99)},
			expectedMem:    SizeInt, // wrapped value mem
			expectedNewMem: SizeInt, // wrapped value mem
			errExpected:    nil,
		},
		{
			name:           "should have a value of 0 given a nil wrapped value",
			val:            &destructKeyedSource{},
			expectedMem:    0,
			expectedNewMem: 0,
			errExpected:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, newTotal, err := tc.val.MemUsage(NewMemUsageContext(New(), 100, 100, 100, 100, nil))
			if err != tc.errExpected {
				t.Fatalf("Unexpected error. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if err != nil && tc.errExpected != nil && err.Error() != tc.errExpected.Error() {
				t.Fatalf("Errors do not match. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if total != tc.expectedMem {
				t.Fatalf("Unexpected memory return. Actual: %v Expected: %v", total, tc.expectedMem)
			}
			if newTotal != tc.expectedNewMem {
				t.Fatalf("Unexpected new memory return. Actual: %v Expected: %v", newTotal, tc.expectedNewMem)
			}
		})
	}
}
