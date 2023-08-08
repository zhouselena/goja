package goja

import (
	"testing"
)

func TestObjectLazyMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            *lazyObject
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should have a value of SizeEmptyStruct given a nil lazy object",
			val:            nil,
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given an empty lazy object",
			val:            &lazyObject{},
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given a base dynamic array with an empty val",
			val:            &lazyObject{val: &Object{}},
			expectedMem:    SizeEmptyStruct + SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct,
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
