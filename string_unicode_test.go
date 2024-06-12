package goja

import (
	"testing"
)

func TestStringUnicodeMemUsage(t *testing.T) {
	tests := []struct {
		name        string
		val         unicodeString
		expectedMem uint64
		errExpected error
	}{
		{
			name:        "should have a value of 0/SizeString given an empty string",
			val:         unicodeString{' '},
			expectedMem: SizeString, // string overhead
			errExpected: nil,
		},
		{
			name:        "should have a value given the length of the string",
			val:         unicodeString{' ', 'y', 'o'},
			expectedMem: 2 + SizeString, // length with string overhead
			errExpected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, err := tc.val.MemUsage(NewMemUsageContext(New(), 100, 100, 100, 100, 0.1, nil))
			if err != tc.errExpected {
				t.Fatalf("Unexpected error. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if err != nil && tc.errExpected != nil && err.Error() != tc.errExpected.Error() {
				t.Fatalf("Errors do not match. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if total != tc.expectedMem {
				t.Fatalf("Unexpected memory return. Actual: %v Expected: %v", total, tc.expectedMem)
			}
		})
	}
}
