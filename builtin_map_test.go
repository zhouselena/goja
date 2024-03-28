package goja

import (
	"fmt"
	"hash/maphash"
	"testing"
)

func TestMapEvilIterator(t *testing.T) {
	const SCRIPT = `
	'use strict';
	var o = {};

	function Iter(value) {
		this.value = value;
		this.idx = 0;
	}

	Iter.prototype.next = function() {
		var idx = this.idx;
		if (idx === 0) {
			this.idx++;
			return this.value;
		}
		return {done: true};
	}

	o[Symbol.iterator] = function() {
		return new Iter({});
	}

	assert.throws(TypeError, function() {
		new Map(o);
	});

	o[Symbol.iterator] = function() {
		return new Iter({value: []});
	}

	function t(prefix) {
		var m = new Map(o);
		assert.sameValue(1, m.size, prefix+": m.size");
		assert.sameValue(true, m.has(undefined), prefix+": m.has(undefined)");
		assert.sameValue(undefined, m.get(undefined), prefix+": m.get(undefined)");
	}

	t("standard adder");

	var count = 0;
	var origSet = Map.prototype.set;

	Map.prototype.set = function() {
		count++;
		origSet.apply(this, arguments);
	}

	t("custom adder");
	assert.sameValue(1, count, "count");

	undefined;
	`
	testScriptWithTestLib(SCRIPT, _undefined, t)
}

func TestMapExportToNilMap(t *testing.T) {
	vm := New()
	var m map[int]interface{}
	res, err := vm.RunString("new Map([[1, true]])")
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

func TestMapExportToNonNilMap(t *testing.T) {
	vm := New()
	m := map[int]interface{}{
		2: true,
	}
	res, err := vm.RunString("new Map([[1, true]])")
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

func TestMapGetAdderGetIteratorOrder(t *testing.T) {
	const SCRIPT = `
	let getterCalled = 0;

	class M extends Map {
	    get set() {
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
	    new M(iterable);
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

func ExampleObject_Export_map() {
	vm := New()
	m, err := vm.RunString(`
	new Map([[1, true], [2, false]]);
	`)
	if err != nil {
		panic(err)
	}
	exp := m.Export()
	fmt.Printf("%T, %v\n", exp, exp)
	// Output: [][2]interface {}, [[1 true] [2 false]]
}

func ExampleRuntime_ExportTo_mapToMap() {
	vm := New()
	m, err := vm.RunString(`
	new Map([[1, true], [2, false]]);
	`)
	if err != nil {
		panic(err)
	}
	exp := make(map[int]bool)
	err = vm.ExportTo(m, &exp)
	if err != nil {
		panic(err)
	}
	fmt.Println(exp)
	// Output: map[1:true 2:false]
}

func ExampleRuntime_ExportTo_mapToSlice() {
	vm := New()
	m, err := vm.RunString(`
	new Map([[1, true], [2, false]]);
	`)
	if err != nil {
		panic(err)
	}
	exp := make([][]interface{}, 0)
	err = vm.ExportTo(m, &exp)
	if err != nil {
		panic(err)
	}
	fmt.Println(exp)
	// Output: [[1 true] [2 false]]
}

func ExampleRuntime_ExportTo_mapToTypedSlice() {
	vm := New()
	m, err := vm.RunString(`
	new Map([[1, true], [2, false]]);
	`)
	if err != nil {
		panic(err)
	}
	exp := make([][2]interface{}, 0)
	err = vm.ExportTo(m, &exp)
	if err != nil {
		panic(err)
	}
	fmt.Println(exp)
	// Output: [[1 true] [2 false]]
}

func BenchmarkMapDelete(b *testing.B) {
	var key1 Value = asciiString("a")
	var key2 Value = asciiString("b")
	one := intToValue(1)
	two := intToValue(2)
	for i := 0; i < b.N; i++ {
		m := newOrderedMap(&maphash.Hash{})
		m.set(key1, one)
		m.set(key2, two)
		if !m.remove(key1) {
			b.Fatal("remove() returned false")
		}
	}
}

func BenchmarkMapDeleteJS(b *testing.B) {
	prg, err := Compile("test.js", `
	var m = new Map([['a',1], ['b', 2]]);
	
	var result = m.delete('a');

	if (!result || m.size !== 1) {
		throw new Error("Fail!");
	}
	`,
		false)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := New()
		_, err := vm.RunProgram(prg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestMapObjectMemUsage(t *testing.T) {
	vm := New()
	tests := []struct {
		name        string
		mu          *MemUsageContext
		mo          *mapObject
		expectedMem uint64
		errExpected error
	}{
		{
			name: "mem below threshold",
			mu:   NewMemUsageContext(vm, 88, 5000, 50, 50, TestNativeMemUsageChecker{}),
			mo: &mapObject{
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
			mu:          NewMemUsageContext(vm, 88, 5000, 50, 50, TestNativeMemUsageChecker{}),
			mo:          nil,
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "mem way above threshold returns first crossing of threshold",
			mu:   NewMemUsageContext(vm, 88, 100, 50, 50, TestNativeMemUsageChecker{}),
			mo: &mapObject{
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, err := tc.mo.MemUsage(tc.mu)
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
