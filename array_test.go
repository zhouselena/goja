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
	tests := []struct {
		name        string
		mu          *MemUsageContext
		ao          *arrayObject
		expected    uint64
		errExpected error
	}{
		{
			name: "mem below threshold",
			mu:   NewMemUsageContext(New(), 88, 5000, 50, 50, TestNativeMemUsageChecker{}),
			ao: &arrayObject{
				values: []Value{
					New()._newString(newStringValue("key"), nil),
				},
			},
			expected:    41,
			errExpected: nil,
		},
		{
			name: "mem way above threshold returns first crossing of threshold",
			mu:   NewMemUsageContext(New(), 88, 100, 50, 50, TestNativeMemUsageChecker{}),
			ao: &arrayObject{
				values: []Value{
					New()._newString(newStringValue("key"), nil),
					New()._newString(newStringValue("key1"), nil),
					New()._newString(newStringValue("key2"), nil),
					New()._newString(newStringValue("key3"), nil),
					New()._newString(newStringValue("key4"), nil),
					New()._newString(newStringValue("key5"), nil),
				},
			},
			expected:    119,
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
				values: []Value{New()._newString(newStringValue("key"), nil)},
			},
			expected:    16,
			errExpected: errArrayLenExceedsThresholdNil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, err := tc.ao.MemUsage(tc.mu)
			if err == nil && tc.errExpected != nil || err != nil && tc.errExpected == nil {
				t.Fatalf("Unexpected error. Actual: %v Expected; %v", err, tc.errExpected)
			}
			if err != nil && tc.errExpected != nil && err.Error() != tc.errExpected.Error() {
				t.Fatalf("Errors do not match. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if total != tc.expected {
				t.Fatalf("Unexpected memory return. Actual: %v Expected: %v", total, tc.expected)
			}
		})
	}
}
