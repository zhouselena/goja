package goja

import (
	"math"
	"testing"

	"github.com/dop251/goja/parser"
	"github.com/dop251/goja/unistring"
)

func TestTaggedTemplateArgExport(t *testing.T) {
	vm := New()
	vm.Set("f", func(v Value) {
		v.Export()
	})
	vm.RunString("f`test`")
}

func TestVM1(t *testing.T) {
	r := &Runtime{}
	r.init()

	vm := r.vm

	vm.prg = &Program{
		values: []Value{valueInt(2), valueInt(3), asciiString("test")},
		code: []instruction{
			&bindGlobal{vars: []unistring.String{"v"}},
			newObject,
			setGlobal("v"),
			loadVal(2),
			loadVal(1),
			loadVal(0),
			add,
			setElem,
			pop,
			loadDynamic("v"),
			halt,
		},
	}

	vm.run()

	rv := vm.pop()

	if obj, ok := rv.(*Object); ok {
		if v := obj.self.getStr("test", nil).ToInteger(); v != 5 {
			t.Fatalf("Unexpected property value: %v", v)
		}
	} else {
		t.Fatalf("Unexpected result: %v", rv)
	}

}

func TestEvalVar(t *testing.T) {
	const SCRIPT = `
	function test() {
		var a;
		return eval("var a = 'yes'; var z = 'no'; a;") === "yes" && a === "yes";
	}
	test();
	`

	testScript(SCRIPT, valueTrue, t)
}

func TestResolveMixedStack1(t *testing.T) {
	const SCRIPT = `
	function test(arg) {
		var a = 1;
		var scope = {};
		(function() {return arg})(); // move arguments to stash
		with (scope) {
			a++; // resolveMixedStack1 here
			return a + arg;
		}
	}
	test(40);
	`

	testScript(SCRIPT, valueInt(42), t)
}

func TestNewArrayFromIterClosed(t *testing.T) {
	const SCRIPT = `
	const [a, ...other] = [];
	assert.sameValue(a, undefined);
	assert(Array.isArray(other));
	assert.sameValue(other.length, 0);
	`
	testScriptWithTestLib(SCRIPT, _undefined, t)
}

func BenchmarkVmNOP2(b *testing.B) {
	prg := []func(*vm){
		//loadVal(0).exec,
		//loadVal(1).exec,
		//add.exec,
		jump(1).exec,
		halt.exec,
	}

	r := &Runtime{}
	r.init()

	vm := r.vm
	vm.prg = &Program{
		values: []Value{intToValue(2), intToValue(3)},
	}

	for i := 0; i < b.N; i++ {
		vm.halt = false
		vm.pc = 0
		for !vm.halt {
			prg[vm.pc](vm)
		}
		//vm.sp--
		/*r := vm.pop()
		if r.ToInteger() != 5 {
			b.Fatalf("Unexpected result: %+v", r)
		}
		if vm.sp != 0 {
			b.Fatalf("Unexpected sp: %d", vm.sp)
		}*/
	}
}

func BenchmarkVmNOP(b *testing.B) {
	r := &Runtime{}
	r.init()

	vm := r.vm
	vm.prg = &Program{
		code: []instruction{
			jump(1),
			//jump(1),
			halt,
		},
	}

	for i := 0; i < b.N; i++ {
		vm.pc = 0
		vm.run()
	}

}

func BenchmarkVm1(b *testing.B) {
	r := &Runtime{}
	r.init()

	vm := r.vm

	//ins1 := loadVal1(0)
	//ins2 := loadVal1(1)

	vm.prg = &Program{
		values: []Value{valueInt(2), valueInt(3)},
		code: []instruction{
			loadVal(0),
			loadVal(1),
			add,
			halt,
		},
	}

	for i := 0; i < b.N; i++ {
		vm.pc = 0
		vm.run()
		r := vm.pop()
		if r.ToInteger() != 5 {
			b.Fatalf("Unexpected result: %+v", r)
		}
		if vm.sp != 0 {
			b.Fatalf("Unexpected sp: %d", vm.sp)
		}
	}
}

func BenchmarkFib(b *testing.B) {
	const TEST_FIB = `
function fib(n) {
if (n < 2) return n;
return fib(n - 2) + fib(n - 1);
}

fib(35);
`
	b.StopTimer()
	prg, err := parser.ParseFile(nil, "test.js", TEST_FIB, 0)
	if err != nil {
		b.Fatal(err)
	}

	c := newCompiler()
	c.compile(prg, false, true, nil)
	c.p.dumpCode(b.Logf)

	r := &Runtime{}
	r.init()

	vm := r.vm

	var expectedResult Value = valueInt(9227465)

	b.StartTimer()

	vm.prg = c.p
	vm.run()
	v := vm.result

	b.Logf("stack size: %d", len(vm.stack))
	b.Logf("stashAllocs: %d", vm.stashAllocs)

	if !v.SameAs(expectedResult) {
		b.Fatalf("Result: %+v, expected: %+v", v, expectedResult)
	}

}

func BenchmarkEmptyLoop(b *testing.B) {
	const SCRIPT = `
	function f() {
		for (var i = 0; i < 100; i++) {
		}
	}
	f()
	`
	b.StopTimer()
	vm := New()
	prg := MustCompile("test.js", SCRIPT, false)
	// prg.dumpCode(log.Printf)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		vm.RunProgram(prg)
	}
}

func BenchmarkVMAdd(b *testing.B) {
	vm := &vm{}
	vm.stack = append(vm.stack, nil, nil)
	vm.sp = len(vm.stack)

	var v1 Value = valueInt(3)
	var v2 Value = valueInt(5)

	for i := 0; i < b.N; i++ {
		vm.stack[0] = v1
		vm.stack[1] = v2
		add.exec(vm)
		vm.sp++
	}
}

func BenchmarkFuncCall(b *testing.B) {
	const SCRIPT = `
	function f(a, b, c, d) {
	}
	`

	b.StopTimer()

	vm := New()
	prg := MustCompile("test.js", SCRIPT, false)

	vm.RunProgram(prg)
	if f, ok := AssertFunction(vm.Get("f")); ok {
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			f(nil, nil, intToValue(1), intToValue(2), intToValue(3), intToValue(4), intToValue(5), intToValue(6))
		}
	} else {
		b.Fatal("f is not a function")
	}
}

func BenchmarkAssertInt(b *testing.B) {
	v := intToValue(42)
	for i := 0; i < b.N; i++ {
		if i, ok := v.(valueInt); !ok || int64(i) != 42 {
			b.Fatal()
		}
	}
}

func TestIntToValue(t *testing.T) {
	for _, tc := range []struct {
		i        int64
		expected Value
	}{
		{
			intCacheMinValue - 1,
			valueInt(intCacheMinValue - 1),
		},
		{
			intCacheMinValue,
			valueInt(intCacheMinValue),
		},
		{
			intCacheMinValue + 1,
			valueInt(intCacheMinValue + 1),
		},
		{
			-1,
			valueInt(-1),
		},
		{
			0,
			valueInt(0),
		},
		{
			1,
			valueInt(1),
		},
		{
			intCacheMaxValue - 1,
			valueInt(intCacheMaxValue - 1),
		},
		{
			intCacheMaxValue,
			valueInt(intCacheMaxValue),
		},
		{
			intCacheMaxValue + 1,
			valueInt(intCacheMaxValue + 1),
		},
	} {
		actual := intToValue(tc.i)
		if tc.expected != actual {
			t.Fatalf("%v is not equal to %v", actual, tc.expected)
		}
	}
}

func TestInt64ToValue(t *testing.T) {
	for _, tc := range []struct {
		i        int64
		expected Value
	}{
		{
			9223372036854775807,
			valueInt64(9223372036854775807),
		},
		{
			math.MaxInt64,
			valueInt64(math.MaxInt64),
		},
	} {
		actual := int64ToValue(tc.i)
		if tc.expected != actual {
			t.Fatalf("%v is not equal to %v", actual, tc.expected)
		}
	}
}

func TestFloatToValue(t *testing.T) {
	for _, tc := range []struct {
		f             float64
		expectedValue Value
	}{
		{
			0.0,
			valueFloat(0),
		},
		{
			2.0000,
			valueFloat(2),
		},
		{
			2.001,
			valueFloat(2.001),
		},
		{
			1.234000,
			valueFloat(1.234),
		},
	} {
		actual := floatToValue(tc.f)
		if tc.expectedValue != actual {
			t.Fatalf("%v is not equal to %v", actual, tc.expectedValue)
		}
	}
}

func TestValueStackMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            valueStack
		memLimit       uint64
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should account for no memory usage given an empty value stack",
			val:            []Value{},
			memLimit:       100,
			expectedMem:    0,
			expectedNewMem: 0,
			errExpected:    nil,
		},
		{
			name:           "should account for no memory usage given a value stack with nil",
			val:            []Value{nil},
			memLimit:       100,
			expectedMem:    0,
			expectedNewMem: 0,
			errExpected:    nil,
		},
		{
			name:     "should account for each value given a non-empty value stack",
			val:      []Value{valueInt(99)},
			memLimit: 100,
			// value
			expectedMem: SizeInt,
			// value
			expectedNewMem: SizeInt,
			errExpected:    nil,
		},
		{
			name:     "should exit early given value stack over the memory limit",
			val:      []Value{valueInt(99), valueInt(99), valueInt(99), valueInt(99)},
			memLimit: 0,
			// value
			expectedMem: SizeInt,
			// value
			expectedNewMem: SizeInt,
			errExpected:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, newTotal, err := tc.val.MemUsage(NewMemUsageContext(New(), 100, tc.memLimit, 100, 100, nil))
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

func TestVMContextMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            *vmContext
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should have a value of SizeEmptyStruct given a nil vmContext",
			val:            nil,
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given an empty vmContext",
			val:            &vmContext{},
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should account for newTarget given a vmContext with non-empty newTarget",
			val:  &vmContext{newTarget: valueInt(99)},
			// vmContext overhead + newTarget value
			expectedMem: SizeEmptyStruct + SizeInt,
			// vmContext overhead + newTarget value
			expectedNewMem: SizeEmptyStruct + SizeInt,
			errExpected:    nil,
		},
		{
			name: "should account for stash given a vmContext with non-empty stash",
			val:  &vmContext{stash: &stash{values: []Value{valueInt(99)}}},
			// vmContext overhead + stash value
			expectedMem: SizeEmptyStruct + SizeInt,
			// vmContext overhead + stash value
			expectedNewMem: SizeEmptyStruct + SizeInt,
			errExpected:    nil,
		},
		{
			name: "should account for stash given a vmContext with non-empty stash",
			val: &vmContext{prg: &Program{
				values: []Value{valueInt(99)},
			}},
			// vmContext overhead + prg value
			expectedMem: SizeEmptyStruct + SizeInt,
			// vmContext overhead + prg value
			expectedNewMem: SizeEmptyStruct + SizeInt,
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

func TestStashMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            *stash
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should have a value of 0 given a nil stash",
			val:            nil,
			expectedMem:    0,
			expectedNewMem: 0,
			errExpected:    nil,
		},
		{
			name:           "should have a value of 0 given an empty stash",
			val:            &stash{},
			expectedMem:    0,
			expectedNewMem: 0,
			errExpected:    nil,
		},
		{
			name: "should account for obj given a stash with non-empty obj",
			val: &stash{
				obj: &Object{
					self: &baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
				},
			},
			// baseObject overhead + obj value with string overhead
			expectedMem: SizeEmptyStruct + (4 + SizeString + SizeInt),
			// baseObject overhead + obj value with string overhead
			expectedNewMem: SizeEmptyStruct + (4 + SizeString + SizeInt),
			errExpected:    nil,
		},
		{
			name: "should account for values given a stash with non-empty values",
			val:  &stash{values: []Value{valueInt(99)}},
			// value
			expectedMem: SizeInt,
			// value
			expectedNewMem: SizeInt,
			errExpected:    nil,
		},
		{
			name: "should account for outer given a stash with non-empty outer",
			val:  &stash{outer: &stash{values: []Value{valueInt(99)}}},
			// outer stash value
			expectedMem: SizeInt,
			// outer stash value
			expectedNewMem: SizeInt,
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
