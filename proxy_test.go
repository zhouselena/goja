package goja

import (
	"testing"

	"github.com/dop251/goja/unistring"
)

func TestProxyMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            *proxyObject
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should have a value of SizeEmptyStruct given a nil proxy object",
			val:            nil,
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should account for base object overhead given an empty proxy object",
			val:  &proxyObject{},
			// proxy overhead + baseObject overhead
			expectedMem: SizeEmptyStruct + SizeEmptyStruct,
			// proxy overhead + baseObject overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should account for baseObject and target overhead given a proxy object with empty target",
			val:  &proxyObject{target: &Object{}},
			// proxy overhead + baseObject overhead + target overhead
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct,
			// proxy overhead + baseObject overhead + target overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should account for baseObjet overhead and target given a proxy object with a non-empty target",
			val: &proxyObject{
				target: &Object{
					self: &baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
				},
			},
			// proxy overhead + baseObject overhead + target overhead + key/value pair
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct + (4 + SizeInt),
			// proxy overhead + baseObject overhead + target overhead + key/value pair with string overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct + (4 + SizeString + SizeInt),
			errExpected:    nil,
		},
		{
			name: "should account for baseObject overhead given a base dynamic array with an empty handler",
			val:  &proxyObject{handler: &jsProxyHandler{handler: &Object{}}},
			// proxy overhead + baseObject overhead + target overhead
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct,
			// proxy overhead + baseObject overhead + target overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should account for baseObject overhead and handler given a base dynamic array with a non-empty handler",
			val: &proxyObject{
				handler: &jsProxyHandler{
					handler: &Object{
						self: &baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
					},
				},
			},
			// proxy overhead + baseObject overhead + target overhead + key/value pair
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct + (4 + SizeInt),
			// proxy overhead + baseObject overhead + target overhead + key/value pair with string overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct + (4 + SizeString + SizeInt),
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

func TestJSProxyHandlerMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            *jsProxyHandler
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should have a value of SizeEmptyStruct given a nil proxy handler",
			val:            nil,
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given an empty proxy handler",
			val:            &jsProxyHandler{},
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should have a value of SizeEmptyStruct given an empty proxy handler",
			val: &jsProxyHandler{
				handler: &Object{
					self: &baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
				},
			},
			// baseObject overhead + key/value pair
			expectedMem: SizeEmptyStruct + (4 + SizeInt),
			// baseObject overhead + key/value pair with string overhead
			expectedNewMem: SizeEmptyStruct + (4 + SizeString + SizeInt),
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
