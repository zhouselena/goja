package goja

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestFuncProto(t *testing.T) {
	const SCRIPT = `
	"use strict";
	function A() {}
	A.__proto__ = Object;
	A.prototype = {};

	function B() {}
	B.__proto__ = Object.create(null);
	var thrown = false;
	try {
		delete B.prototype;
	} catch (e) {
		thrown = e instanceof TypeError;
	}
	thrown;
	`
	testScript(SCRIPT, valueTrue, t)
}

func TestFuncPrototypeRedefine(t *testing.T) {
	const SCRIPT = `
	let thrown = false;
	try {
		Object.defineProperty(function() {}, "prototype", {
			set: function(_value) {},
		});
	} catch (e) {
		if (e instanceof TypeError) {
			thrown = true;
		} else {
			throw e;
		}
	}
	thrown;
	`

	testScript(SCRIPT, valueTrue, t)
}

func TestFuncExport(t *testing.T) {
	vm := New()
	typ := reflect.TypeOf((func(FunctionCall) Value)(nil))

	f := func(expr string, t *testing.T) {
		v, err := vm.RunString(expr)
		if err != nil {
			t.Fatal(err)
		}
		if actualTyp := v.ExportType(); actualTyp != typ {
			t.Fatalf("Invalid export type: %v", actualTyp)
		}
		ev := v.Export()
		if actualTyp := reflect.TypeOf(ev); actualTyp != typ {
			t.Fatalf("Invalid export value: %v", ev)
		}
	}

	t.Run("regular function", func(t *testing.T) {
		f("(function() {})", t)
	})

	t.Run("arrow function", func(t *testing.T) {
		f("(()=>{})", t)
	})

	t.Run("method", func(t *testing.T) {
		f("({m() {}}).m", t)
	})

	t.Run("class", func(t *testing.T) {
		f("(class {})", t)
	})
}

func TestFuncWrapUnwrap(t *testing.T) {
	vm := New()
	f := func(a int, b string) bool {
		return a > 0 && b != ""
	}
	var f1 func(int, string) bool
	v := vm.ToValue(f)
	if et := v.ExportType(); et != reflect.TypeOf(f1) {
		t.Fatal(et)
	}
	err := vm.ExportTo(v, &f1)
	if err != nil {
		t.Fatal(err)
	}
	if !f1(1, "a") {
		t.Fatal("not true")
	}
}

func TestWrappedFunc(t *testing.T) {
	vm := New()
	f := func(a int, b string) bool {
		return a > 0 && b != ""
	}
	vm.Set("f", f)
	const SCRIPT = `
	assert.sameValue(typeof f, "function");
	const s = f.toString()
	assert(s.endsWith("TestWrappedFunc.func1() { [native code] }"), s);
	assert(f(1, "a"));
	assert(!f(0, ""));
	`
	vm.testScriptWithTestLib(SCRIPT, _undefined, t)
}

func TestWrappedFuncErrorPassthrough(t *testing.T) {
	vm := New()
	e := errors.New("test")
	f := func(a int) error {
		if a > 0 {
			return e
		}
		return nil
	}

	var f1 func(a int64) error
	err := vm.ExportTo(vm.ToValue(f), &f1)
	if err != nil {
		t.Fatal(err)
	}
	if err := f1(1); err != e {
		t.Fatal(err)
	}
}

func ExampleAssertConstructor() {
	vm := New()
	res, err := vm.RunString(`
		(class C {
			constructor(x) {
				this.x = x;
			}
		})
	`)
	if err != nil {
		panic(err)
	}
	if ctor, ok := AssertConstructor(res); ok {
		obj, err := ctor(nil, vm.ToValue("Test"))
		if err != nil {
			panic(err)
		}
		fmt.Print(obj.Get("x"))
	} else {
		panic("Not a constructor")
	}
	// Output: Test
}

func TestNativeFuncObjectMemUsage(t *testing.T) {
	tests := []struct {
		name        string
		val         *nativeFuncObject
		expectedMem uint64
		errExpected error
	}{
		{
			name:        "should have a value given by the wrapped value",
			val:         &nativeFuncObject{},
			expectedMem: SizeEmptyStruct, // baseFuncObject
			errExpected: nil,
		},
		{
			name:        "should have a value of SizeEmptyStruct given a nil nativeFuncObject",
			val:         nil,
			expectedMem: SizeEmptyStruct,
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

func TestFuncObjectMemUsage(t *testing.T) {
	tests := []struct {
		name        string
		val         *funcObject
		expectedMem uint64
		errExpected error
	}{
		{
			name:        "should have a value of SizeEmptyStruct given a nil funcObject",
			val:         nil,
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name:        "should have the correct value given an empty funcObject",
			val:         &funcObject{},
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a baseJSfuncObject with values in stash",
			val: &funcObject{
				baseJsFuncObject: baseJsFuncObject{
					stash: &stash{
						values: []Value{valueInt(0)},
					},
				},
			},
			// baseJsFuncObject + value in baseJsFuncObject stash + []Value
			expectedMem: SizeEmptyStruct + SizeInt + SizeEmptySlice,
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

func TestBaseJsFuncObjectMemUsage(t *testing.T) {
	tests := []struct {
		name        string
		val         *baseJsFuncObject
		expectedMem uint64
		errExpected error
	}{
		{
			name:        "should have a value of SizeEmptyStruct given a nil baseJsFuncObject",
			val:         nil,
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name:        "should have a value of SizeEmptyStruct given an empty baseJsFuncObject",
			val:         &baseJsFuncObject{},
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a baseJsFuncObject with values in stash",
			val: &baseJsFuncObject{
				stash: &stash{
					values: []Value{valueInt(0)},
				},
			},
			// baseJsFuncObject + value in baseJsFuncObject stash + []Value
			expectedMem: SizeEmptyStruct + SizeInt + SizeEmptySlice,
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

func TestClassFuncObjectMemUsage(t *testing.T) {
	tests := []struct {
		name        string
		val         *classFuncObject
		expectedMem uint64
		errExpected error
	}{
		{
			name:        "should have a value of SizeEmptyStruct given a nil classFuncObject",
			val:         nil,
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name:        "should have a value of SizeEmptyStruct given an empty classFuncObject",
			val:         &classFuncObject{},
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a classFuncObject with valid baseJSFuncObject",
			val: &classFuncObject{
				baseJsFuncObject: baseJsFuncObject{
					stash: &stash{
						values: []Value{valueInt(0)},
					},
				},
			},
			// baseJsFuncObject + value baseJsFuncObject in stash + []Value
			expectedMem: SizeEmptyStruct + SizeInt + SizeEmptySlice,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a classFuncObject with valid initFields",
			val: &classFuncObject{
				initFields: &Program{
					values: []Value{valueInt(0)},
				},
			},
			// baseJsFuncObject + value in Program
			expectedMem: SizeEmptyStruct + SizeInt,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a classFuncObject with valid computedKeys",
			val: &classFuncObject{
				computedKeys: []Value{valueInt(0)},
			},
			// baseJsFuncObject + value in computedKeys
			expectedMem: SizeEmptyStruct + SizeInt,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a classFuncObject with valid privateMethods",
			val: &classFuncObject{
				privateMethods: []Value{valueInt(0)},
			},
			// baseJsFuncObject + value in privateMethods
			expectedMem: SizeEmptyStruct + SizeInt,
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

func TestMethodFuncObjectMemUsage(t *testing.T) {
	tests := []struct {
		name        string
		val         *methodFuncObject
		expectedMem uint64
		errExpected error
	}{
		{
			name:        "should have a value of SizeEmptyStruct given a nil methodFuncObject",
			val:         nil,
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should have a value of SizeEmptyStruct given an empty methodFuncObject",
			val:  &methodFuncObject{},
			// methodFuncObject + nil Object
			expectedMem: SizeEmptyStruct + SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a methodFuncObject with values in stash",
			val: &methodFuncObject{
				baseJsFuncObject: baseJsFuncObject{
					stash: &stash{
						values: []Value{valueInt(0)},
					},
				},
			},
			// methodFuncObject + nil Object + value in baseJsFuncObject stash + []Value
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + SizeInt + SizeEmptySlice,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a methodFuncObject with nil homeObject",
			val: &methodFuncObject{
				homeObject: nil,
			},
			// methodFuncObject + nil Object
			expectedMem: SizeEmptyStruct + SizeEmptyStruct,
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

func TestArrowFuncObjectMemUsage(t *testing.T) {
	tests := []struct {
		name        string
		val         *arrowFuncObject
		expectedMem uint64
		errExpected error
	}{
		{
			name:        "should have a value of SizeEmptyStruct given a nil arrowFuncObject",
			val:         nil,
			expectedMem: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should have a value of SizeEmptyStruct given an empty arrowFuncObject",
			val:  &arrowFuncObject{},
			// arrowFuncObject + nil Object
			expectedMem: SizeEmptyStruct + SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a arrowFuncObject with values in stash",
			val: &arrowFuncObject{
				baseJsFuncObject: baseJsFuncObject{
					stash: &stash{
						values: []Value{valueInt(0)},
					},
				},
			},
			// arrowFuncObject + nil Object + value in baseJsFuncObject stash + []Value
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + SizeInt + SizeEmptySlice,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a arrowFuncObject with nil funcObj",
			val: &arrowFuncObject{
				funcObj: nil,
			},
			// arrowFuncObject + nil Object
			expectedMem: SizeEmptyStruct + SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should have the correct value given a valid newTarget",
			val: &arrowFuncObject{
				newTarget: valueInt(0),
			},
			// arrowFuncObject + nil Object + valueInt 0
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + SizeInt,
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
