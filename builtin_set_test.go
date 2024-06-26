package goja

import (
	"fmt"
	"strings"
	"testing"
)

func TestSetEvilIterator(t *testing.T) {
	const SCRIPT = `
	var o = {};
	o[Symbol.iterator] = function() {
		return {
			next: function() {
				if (!this.flag) {
					this.flag = true;
					return {};
				}
				return {done: true};
			}
		}
	}
	new Set(o);
	undefined;
	`
	testScript(SCRIPT, _undefined, t)
}

func ExampleRuntime_ExportTo_setToMap() {
	vm := New()
	s, err := vm.RunString(`
	new Set([1, 2, 3])
	`)
	if err != nil {
		panic(err)
	}
	m := make(map[int]struct{})
	err = vm.ExportTo(s, &m)
	if err != nil {
		panic(err)
	}
	fmt.Println(m)
	// Output: map[1:{} 2:{} 3:{}]
}

func ExampleRuntime_ExportTo_setToSlice() {
	vm := New()
	s, err := vm.RunString(`
	new Set([1, 2, 3])
	`)
	if err != nil {
		panic(err)
	}
	var a []int
	err = vm.ExportTo(s, &a)
	if err != nil {
		panic(err)
	}
	fmt.Println(a)
	// Output: [1 2 3]
}

func TestSetExportToSliceCircular(t *testing.T) {
	vm := New()
	s, err := vm.RunString(`
	let s = new Set();
	s.add(s);
	s;
	`)
	if err != nil {
		t.Fatal(err)
	}
	var a []Value
	err = vm.ExportTo(s, &a)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 1 {
		t.Fatalf("len: %d", len(a))
	}
	if a[0] != s {
		t.Fatalf("a: %v", a)
	}
}

func TestSetExportToArrayMismatchedLengths(t *testing.T) {
	vm := New()
	s, err := vm.RunString(`
	new Set([1, 2])
	`)
	if err != nil {
		panic(err)
	}
	var s1 [3]int
	err = vm.ExportTo(s, &s1)
	if err == nil {
		t.Fatal("expected error")
	}
	if msg := err.Error(); !strings.Contains(msg, "lengths mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetExportToNilMap(t *testing.T) {
	vm := New()
	var m map[int]interface{}
	res, err := vm.RunString("new Set([1])")
	if err != nil {
		t.Fatal(err)
	}
	err = vm.ExportTo(res, &m)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 1 {
		t.Fatal(m)
	}
	if _, exists := m[1]; !exists {
		t.Fatal(m)
	}
}

func TestSetExportToNonNilMap(t *testing.T) {
	vm := New()
	m := map[int]interface{}{
		2: true,
	}
	res, err := vm.RunString("new Set([1])")
	if err != nil {
		t.Fatal(err)
	}
	err = vm.ExportTo(res, &m)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 1 {
		t.Fatal(m)
	}
	if _, exists := m[1]; !exists {
		t.Fatal(m)
	}
}

func TestSetGetAdderGetIteratorOrder(t *testing.T) {
	const SCRIPT = `
	let getterCalled = 0;

	class S extends Set {
	    get add() {
	        getterCalled++;
	        return null;
	    }
	}

	let getIteratorCalled = 0;

	let iterable = {};
	iterable[Symbol.iterator] = () => {
	    getIteratorCalled++
	    return {
	        next: 1
	    };
	}

	let thrown = false;

	try {
	    new S(iterable);
	} catch (e) {
	    if (e instanceof TypeError) {
	        thrown = true;
	    } else {
	        throw e;
	    }
	}

	thrown && getterCalled === 1 && getIteratorCalled === 0;
	`
	testScript(SCRIPT, valueTrue, t)
}

func TestSetHasFloatVsInt(t *testing.T) {
	const SCRIPT = `const s = new Set()
	s.add(1);
	const hasFloat = s.has(1.0);
	const doesNotHaveFloat = s.has(1.3);

	s.add(2.0)
	const hasInt = s.has(2)

	hasFloat && hasInt && !doesNotHaveFloat`

	testScript(SCRIPT, valueTrue, t)
}

func createOrderedMapWithNilValues(size int) *orderedMap {
	ht := make(map[uint64]*mapEntry, 0)
	for i := 0; i < size; i += 1 {
		ht[uint64(i)] = &mapEntry{
			key:   nil,
			value: nil,
		}
		// These iter items are necessary for testing the mem usage
		// estimation since that's how we iterate through the map.
		if i > 0 {
			ht[uint64(i)].iterPrev = ht[uint64(i-1)]
			ht[uint64(i-1)].iterNext = ht[uint64(i)]
		}
	}
	return &orderedMap{
		size:      size,
		iterFirst: ht[uint64(0)],
		iterLast:  ht[uint64(size-1)],
		hashTable: ht,
	}
}

func TestSetObjectMemUsage(t *testing.T) {
	vm := New()

	tests := []struct {
		name        string
		mu          *MemUsageContext
		so          *setObject
		expectedMem uint64
		errExpected error
	}{
		{
			name: "mem below threshold",
			mu:   NewMemUsageContext(vm, 88, 5000, 50, 50, 0.1, TestNativeMemUsageChecker{}),
			so: &setObject{
				m: &orderedMap{
					hashTable: map[uint64]*mapEntry{
						1: {
							key:   vm.ToValue("key"),
							value: vm.ToValue("value"),
						},
					},
				},
			},
			// baseObject + (len(key) + overhead)  + (len(value) + overhead)
			expectedMem: SizeEmptyStruct + (3 + SizeString) + (5 + SizeString),
			errExpected: nil,
		},
		{
			name:        "mem is SizeEmptyStruct given a nil map object",
			mu:          NewMemUsageContext(vm, 88, 5000, 50, 50, 0.1, TestNativeMemUsageChecker{}),
			so:          nil,
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "mem way above threshold returns first crossing of threshold",
			mu:   NewMemUsageContext(vm, 88, 100, 50, 50, 0.1, TestNativeMemUsageChecker{}),
			so: &setObject{
				m: &orderedMap{
					hashTable: map[uint64]*mapEntry{
						1: {
							key:   vm.ToValue("key"),
							value: vm.ToValue("value"),
						},
						2: {
							key:   vm.ToValue("key"),
							value: vm.ToValue("value"),
						},
						3: {
							key:   vm.ToValue("key"),
							value: vm.ToValue("value"),
						},
						4: {
							key:   vm.ToValue("key"),
							value: vm.ToValue("value"),
						},
					},
				},
			},
			// baseObject
			expectedMem: SizeEmptyStruct +
				// len(key) + overhead (we reach the limit after 3)
				(3+SizeString)*3 +
				// len(value) + overhead (we reach the limit after 3)
				(5+SizeString)*3,
			errExpected: nil,
		},
		{
			name: "mem above estimate threshold and within memory limit returns correct mem usage",
			mu:   NewMemUsageContext(vm, 88, 100, 50, 5, 0.1, TestNativeMemUsageChecker{}),
			so: &setObject{
				m: createOrderedMap(vm, 20),
			},
			// baseObject
			expectedMem: SizeEmptyStruct +
				// len(key) + overhead (we reach the limit after 3)
				(3+SizeString)*20 +
				// len(value) + overhead (we reach the limit after 3)
				(5+SizeString)*20,
			errExpected: nil,
		},
		{
			name: "mem above estimate threshold and within memory limit and nil values returns correct mem usage",
			mu:   NewMemUsageContext(vm, 88, 100, 50, 1, 0.1, TestNativeMemUsageChecker{}),
			so: &setObject{
				m: createOrderedMapWithNilValues(3),
			},
			// baseObject
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name:        "mem is SizeEmptyStruct given a nil orderedMap object",
			mu:          NewMemUsageContext(vm, 88, 5000, 50, 50, 0.1, TestNativeMemUsageChecker{}),
			so:          &setObject{},
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, err := tc.so.MemUsage(tc.mu)
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
