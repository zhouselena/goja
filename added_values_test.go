package goja

import (
	"testing"
)

func TestNumberEquality(t *testing.T) {
	vm := New()

	res, err := vm.RunString(`var a = new Number('5')
	a`)
	if err != nil {
		t.Fatal(err)
	}
	if !valueInt(5).Equals(res) {
		t.Fatal("values are not equal")
	}
}

func TestInt64StrictEqualsFloat(t *testing.T) {
	if !valueInt64(5).StrictEquals(valueFloat(5.0)) {
		t.Fatal("values are not equal")
	}

	if !valueInt64(0).StrictEquals(valueFloat(0.0)) {
		t.Fatal("values are not equal")
	}
}

func TestIntStringEquality(t *testing.T) {
	vm := New()

	res, err := vm.RunString(`"0"==0`)
	if err != nil {
		t.Fatal(err)
	}
	if !valueBool(true).Equals(res) {
		t.Fatal("values are not equal")
	}

	res, err = vm.RunString(`"0.0"===0`)
	if err != nil {
		t.Fatal(err)
	}
	if !valueBool(false).Equals(res) {
		t.Fatal("values should not be equal")
	}
}

func TestAddedValuesMemUsage(t *testing.T) {
	vm := New()

	for _, tc := range []struct {
		name        string
		val         MemUsageReporter
		expectedMem uint64
	}{
		{
			name:        "should have memory usage of SizeNumber given a non-empty valueNumber",
			val:         valueNumber{val: 0},
			expectedMem: SizeNumber,
		},
		{
			name:        "should have memory usage of SizeInt32 given a non-empty valueUInt32",
			val:         valueUInt32(1),
			expectedMem: SizeInt32,
		},
		{
			name:        "should have memory usage of SizeInt32 given a non-empty valueInt32",
			val:         valueInt32(1),
			expectedMem: SizeInt32,
		},
		{
			name:        "should have memory usage of SizeNumber given a non-empty valueInt64",
			val:         valueInt64(1),
			expectedMem: SizeNumber,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mem, err := tc.val.MemUsage(NewMemUsageContext(vm, 100, 100, 100, 100, nil))
			if err != nil {
				t.Fatalf("Unexpected error. Actual: %v Expected: nil", err)
			}
			if mem != tc.expectedMem {
				t.Fatalf("Unexpected memory return. Actual: %v Expected: %v", mem, tc.expectedMem)
			}
		})
	}
}
