package goja

import (
	"reflect"
	"testing"
)

func TestArray1(t *testing.T) {
	r := &Runtime{}
	a := r.newArray(nil)
	a.setOwnIdx(valueInt(0), asciiString("test"), true)
	if l := a.getStr("length", nil).ToInteger(); l != 1 {
		t.Fatalf("Unexpected length: %d", l)
	}
}

func TestArrayExportProps(t *testing.T) {
	vm := New()
	arr := vm.NewArray()
	err := arr.DefineDataProperty("0", vm.ToValue(true), FLAG_TRUE, FLAG_FALSE, FLAG_TRUE)
	if err != nil {
		t.Fatal(err)
	}
	actual := arr.Export()
	expected := []interface{}{true}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Expected: %#v, actual: %#v", expected, actual)
	}
}

func TestArrayCanonicalIndex(t *testing.T) {
	const SCRIPT = `
	var a = [];
	a["00"] = 1;
	a["01"] = 2;
	if (a[0] !== undefined) {
		throw new Error("a[0]");
	}
	`

	testScript(SCRIPT, _undefined, t)
}

func BenchmarkArrayGetStr(b *testing.B) {
	b.StopTimer()
	r := New()
	v := &Object{runtime: r}

	a := &arrayObject{
		baseObject: baseObject{
			val:        v,
			extensible: true,
		},
	}
	v.self = a

	a.init()

	v.setOwn(valueInt(0), asciiString("test"), false)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		a.getStr("0", nil)
	}

}

func BenchmarkArrayGet(b *testing.B) {
	b.StopTimer()
	r := New()
	v := &Object{runtime: r}

	a := &arrayObject{
		baseObject: baseObject{
			val:        v,
			extensible: true,
		},
	}
	v.self = a

	a.init()

	var idx Value = valueInt(0)

	v.setOwn(idx, asciiString("test"), false)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		v.get(idx, nil)
	}

}

func BenchmarkArrayPut(b *testing.B) {
	b.StopTimer()
	r := New()

	v := &Object{runtime: r}

	a := &arrayObject{
		baseObject: baseObject{
			val:        v,
			extensible: true,
		},
	}

	v.self = a

	a.init()

	var idx Value = valueInt(0)
	var val Value = asciiString("test")

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		v.setOwn(idx, val, false)
	}

}

func BenchmarkArraySetEmpty(b *testing.B) {
	r := New()
	_ = r.Get("Array").(*Object).Get("prototype").String() // materialise Array.prototype
	a := r.NewArray(0, 0)
	values := a.self.(*arrayObject).values
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values[0] = nil
		a.self.setOwnIdx(0, valueTrue, true)
	}
}

func TestArrayObjectMemUsage(t *testing.T) {
	vm := New()

	tests := []struct {
		name           string
		mu             *MemUsageContext
		ao             *arrayObject
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name: "mem below threshold given a nil slice of values",
			mu:   NewMemUsageContext(vm, 88, 5000, 50, 50, TestNativeMemUsageChecker{}),
			ao:   &arrayObject{},
			// array overhead + array baseObject
			expectedMem: SizeEmptyStruct + SizeEmptyStruct,
			// array overhead + array baseObject
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "mem below threshold given empty slice of values",
			mu:   NewMemUsageContext(vm, 88, 5000, 50, 50, TestNativeMemUsageChecker{}),
			ao:   &arrayObject{values: []Value{}},
			// array overhead + array baseObject + values slice overhead
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptySlice,
			// array overhead + array baseObject + values slice overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptySlice,
			errExpected:    nil,
		},
		{
			name: "mem way above threshold returns first crossing of threshold",
			mu:   NewMemUsageContext(vm, 88, 100, 50, 50, TestNativeMemUsageChecker{}),
			ao: &arrayObject{
				values: []Value{
					vm.ToValue("key0"),
					vm.ToValue("key1"),
					vm.ToValue("key2"),
					vm.ToValue("key3"),
					vm.ToValue("key4"),
					vm.ToValue("key5"),
				},
			},
			// array overhead + array baseObject + values slice overhead
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptySlice +
				// len("keyN") with string overhead * entries (at 4 we reach the limit)
				(4+SizeString)*4,
			// array overhead + array baseObject + values slice overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptySlice +
				// len("keyN") with string overhead * entries (at 4 we reach the limit)
				(4+SizeString)*4,
			errExpected: nil,
		},
		{
			name: "array limit function undefined throws error",
			mu: &MemUsageContext{
				visitTracker: visitTracker{
					objsVisited:    map[objectImpl]bool{},
					stashesVisited: map[*stash]bool{}},
				depthTracker: &depthTracker{
					curDepth: 0,
					maxDepth: 50,
				},
				NativeMemUsageChecker: &TestNativeMemUsageChecker{},
				memoryLimit:           50,
				ObjectPropsLenExceedsThreshold: func(objPropsLen int) bool {
					// number of obj props beyond which we should estimate mem usage
					return objPropsLen > 50
				},
			},
			ao: &arrayObject{
				values: []Value{vm._newString(newStringValue("key"), nil)},
			},
			expectedMem:    SizeEmptyStruct + SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct,
			errExpected:    errArrayLenExceedsThresholdNil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, newTotal, err := tc.ao.MemUsage(tc.mu)
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
