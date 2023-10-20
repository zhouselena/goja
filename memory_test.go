package goja

import (
	"fmt"
	"testing"
)

const (
	testNativeValueMemUsage = 100
	memUsageLimit           = uint64(10000)
	arrLenThreshold         = 1_000
	objPropsLenThreshold    = 1_000
)

type TestNativeValue struct {
}

type TestNativeMemUsageChecker struct {
}

func (muc TestNativeMemUsageChecker) NativeMemUsage(value interface{}) (uint64, bool) {
	switch value.(type) {
	case TestNativeValue:
		return testNativeValueMemUsage, true
	default:
		return 0, false
	}
}

func TestMemCheck(t *testing.T) {
	// This is the sum of property names allocated at each new (empty) scope
	var emptyFunctionScopeOverhead uint64 = 8
	var functionStackOverhead uint64 = 34

	for _, tc := range []struct {
		desc                string
		script              string
		expectedSizeDiff    uint64
		expectedNewSizeDiff uint64
	}{
		{
			desc: "number",
			script: `x = []
			x.push(0)
			checkMem()
			x.push(0)
			checkMem()`,
			expectedSizeDiff:    SizeNumber,
			expectedNewSizeDiff: SizeNumber,
		},
		{
			desc: "boolean",
			script: `x = []
			x.push(true)
			checkMem()
			x.push(true)
			checkMem()`,
			expectedSizeDiff:    SizeBool,
			expectedNewSizeDiff: SizeBool,
		},
		{
			desc: "null",
			script: `x = []
			x.push(null)
			checkMem()
			x.push(null)
			checkMem()`,
			expectedSizeDiff:    SizeEmptyStruct,
			expectedNewSizeDiff: SizeEmptyStruct,
		},
		{
			desc: "undefined",
			script: `x = []
			x.push(undefined)
			checkMem()
			x.push(undefined)
			checkMem()`,
			expectedSizeDiff:    SizeEmptyStruct,
			expectedNewSizeDiff: SizeEmptyStruct,
		},
		{
			desc: "string",
			script: `x = []
			x.push("12345")
			checkMem()
			x.push("12345")
			checkMem()`,
			expectedSizeDiff:    5 + SizeString,
			expectedNewSizeDiff: 5 + SizeString,
		},
		{
			desc: "string with multi-byte characters",
			script: `x = []
			x.push("\u2318")
			checkMem()
			x.push("\u2318")
			checkMem()`,
			expectedSizeDiff:    3 + SizeString, // single char with 3-byte width
			expectedNewSizeDiff: 3 + SizeString,
		},
		{
			desc: "nested objects",
			script: `y = []
			y.push(null)
			checkMem()
			y.push({"a":10, "b":"1234", "c":{}})
			checkMem()`,
			expectedSizeDiff: SizeEmptyStruct + SizeEmptyStruct + // outer object + reference to its prototype
				(1 + SizeString) + SizeNumber + // "a" and number 10
				(1 + SizeString) + (4 + SizeString) + // "b" and string "1234"
				(1 + SizeString) + SizeEmptyStruct + SizeEmptyStruct + //  "c" (object + prototype reference)
				SizeEmptyStruct, // stack difference from popping null(8) and then adding outer obj(8) + "c" obj (8)
			expectedNewSizeDiff: SizeEmptyStruct + SizeEmptyStruct + // outer object + reference to its prototype
				(1 + SizeString) + SizeNumber + // "a" and number 10
				(1 + SizeString) + (4 + SizeString) + // "b" and string "1234"
				(1 + SizeString) + SizeEmptyStruct + SizeEmptyStruct + //  "c" (object + prototype reference)
				SizeEmptyStruct, // stack difference from popping null(8) and then adding outer obj(8) + "c" obj (8),
		},
		{
			desc: "array of numbers",
			script: `y = []
			var i = 0;
			y.push([]);
			checkMem();
			for(i=0;i<20;i++){
				y[0].push(i);
			};
			checkMem()`,
			// Array overhead,
			// size of property values,
			expectedSizeDiff:    SizeEmptyStruct + 20*SizeNumber,
			expectedNewSizeDiff: SizeEmptyStruct + 20*SizeNumber,
		},
		{
			desc: "overhead of a single new scope",
			script: `checkMem();
			(function(){
				checkMem();
			})();`, // over
			expectedSizeDiff: emptyFunctionScopeOverhead +
				functionStackOverhead + // anonymous function on stack
				SizeString + // empty string for anon function name
				SizeString + SizeString + // overhead of "name" and "length" props on function object
				SizeEmptyStruct, // undefined return on stack,
			expectedNewSizeDiff: emptyFunctionScopeOverhead +
				functionStackOverhead + // anonymous function on stack
				SizeString + // empty string for anon function name
				SizeString + SizeString + // overhead of "name" and "length" props on function object
				SizeEmptyStruct, // undefined return on stack,
		},
		{
			desc: "previous function scopes should not affect the current memory",
			script: `checkMem();
			(function(){
			})();
			checkMem();`,
			expectedSizeDiff: 0 +
				SizeEmptyStruct, // undefined return value on stack
			expectedNewSizeDiff: 0 +
				SizeEmptyStruct, // undefined return value on stack
		},
		{
			desc: "overhead of each scope is equivalent regardless of depth",
			script: `checkMem();
			(function(){
				(function(){
					(function(){
						(function(){
							(function(){
								(function(){
									checkMem();
								})();
							})();
						})();
					})();
				})();
			})();`,
			expectedSizeDiff: (6 * functionStackOverhead) + // anonymous functions on stack
				(6 * SizeString) + // empty string for anon function name
				(6 * (SizeString + SizeString)) + // overhead of "name" and "length" props on function object
				(6 * SizeEmptyStruct) + // undefined return value for each function on stack
				(6 * emptyFunctionScopeOverhead),
			expectedNewSizeDiff: (6 * functionStackOverhead) + // anonymous functions on stack
				(6 * SizeString) + // empty string for anon function name
				(6 * (SizeString + SizeString)) + // overhead of "name" and "length" props on function object
				(6 * SizeEmptyStruct) + // undefined return value for each function on stack
				(6 * emptyFunctionScopeOverhead),
		},
		{
			desc: "values attached to lexical scope in a function",
			script: `checkMem();
			(function(){
				var zzzx = 10;
				checkMem();
			})();`,
			// function overhead plus the number value of the "zzzx" property and its string name
			expectedSizeDiff: emptyFunctionScopeOverhead + SizeNumber + functionStackOverhead +
				SizeString + // empty string for anon function name
				SizeString + SizeString + // overhead of "name" and "length" props on function object
				SizeEmptyStruct + // undefined return value on stack
				SizeNumber, // number 10 on stack,
			// function overhead plus the number value of the "zzzx" property and its string name
			expectedNewSizeDiff: emptyFunctionScopeOverhead + SizeNumber + functionStackOverhead +
				SizeString + // empty string for anon function name
				SizeString + SizeString + // overhead of "name" and "length" props on function object
				SizeEmptyStruct + // undefined return value on stack
				SizeNumber, // number 10 on stack,
		},
		{
			desc: "cyclical data structure",
			script: // cyclical data structure does not recurse infinitely
			// and does not artificially inflate mem count. The only change in mem
			// between the two checks is for the new property names for "y" and "x".
			`var zzza = {}
			 var zzzb = {}
			 checkMem();
			 zzza.y = zzzb
			 zzzb.x = zzza
			 checkMem()`,
			expectedSizeDiff: (1 + SizeString) + SizeEmptyStruct + // "x" property name + references to object
				(1 + SizeString) + SizeEmptyStruct, // "y" property names + references to object,
			expectedNewSizeDiff: (1 + SizeString) + SizeEmptyStruct + // "x" property name + references to object
				(1 + SizeString) + SizeEmptyStruct, // "y" property names + references to object,
		},
		{
			desc: "sparse array with arrayObject",
			script: `x = []
			x[1] = "abcd";
			checkMem()
			x[10] = "abc";
			checkMem()`,
			expectedSizeDiff:    2 + SizeString, // 3 -> "abc" added to global memory | -1 difference on stack between "abc" and  "abcd"
			expectedNewSizeDiff: 2 + SizeString, // 3 -> "abc" added to global memory | -1 difference on stack between "abc" and  "abcd",
		},
		{
			desc: "sparse array with sparseArrayObject",
			script: `x = []
			x[5000] = "abcd";
			checkMem()
			x[5001] = "abc";
			checkMem()`,
			expectedSizeDiff: SizeInt32 +
				2 + SizeString, // 3 -> "abc" added to global memory | -1 difference on stack between "abc" and  "abcd"
			expectedNewSizeDiff: SizeInt32 +
				2 + SizeString, // 3 -> "abc" added to global memory | -1 difference on stack between "abc" and  "abcd"
		},
		{
			desc: "array with non-numeric keys",
			script: `x = []
			x[0] = 3;
			checkMem()
			x["a"] = "abc";
			x[1] = 3;
			checkMem()
			`,
			expectedSizeDiff: 1 + SizeString + // len("a")
				3 + SizeString + // len("abc")
				SizeNumber, // number 3,
			expectedNewSizeDiff: 1 + SizeString + // len("a")
				3 + SizeString + // len("abc")
				SizeNumber, // number 3,
		},
		{
			desc: "reference to array",
			script: `x = []
			x[1] = "abcd";
			x[10] = "abc";
			checkMem()
			y = x;
			checkMem()`,
			// len("y") + reference to array
			expectedSizeDiff: (1 + SizeString) + SizeEmptyStruct,
			// len("y") + reference to array
			expectedNewSizeDiff: (1 + SizeString) + SizeEmptyStruct,
		},
		{
			desc: "reference to sparse array",
			script: `x = []
			x[5000] = "abcb";
			x[5001] = "abc";
			checkMem()
			y = x;
			// len("y") + reference to array
			checkMem()`,
			expectedSizeDiff:    (1 + SizeString) + SizeEmptyStruct,
			expectedNewSizeDiff: (1 + SizeString) + SizeEmptyStruct,
		},
		{
			desc: "Date object",
			script: `
			d1 = new Date();
			checkMem();
			d2 = new Date();
			checkMem()
			`,
			// len("d2") + size of msec + reference to visited base object + base object prototype reference
			expectedSizeDiff: (2 + SizeString) + SizeNumber + SizeEmptyStruct + SizeEmptyStruct,
			// len("d2") + size of msec + reference to visited base object + base object prototype reference
			expectedNewSizeDiff: (2 + SizeString) + SizeNumber + SizeEmptyStruct + SizeEmptyStruct,
		},
		{
			desc: "Empty object",
			script: `
			checkMem();
			o = {}
			checkMem()
			`,
			// len("o") + object's starting size + reference to prototype
			expectedSizeDiff: (1 + SizeString) + SizeEmptyStruct + SizeNumber,
			// len("o") + object's starting size + reference to prototype
			expectedNewSizeDiff: (1 + SizeString) + SizeEmptyStruct + SizeNumber,
		},
		{
			desc: "Map",
			script: `var m = new Map();
			m.set("first", 1);
			checkMem();
			m.set("abc", {"a":10, "b":"1234"});
			checkMem();`,
			expectedSizeDiff: 3 + SizeString + // "abc"
				SizeEmptyStruct + SizeEmptyStruct + // outer object + reference to its prototype
				(1 + SizeString) + SizeNumber + // "a" and number
				(1 + SizeString) + (4 + SizeString) + // "b" and "1234" string
				// stack difference in going from
				//	[..other, first, 1]
				//  to
				//	[..other, abc, [object Object], 1234]
				18,
			expectedNewSizeDiff: 3 + SizeString + // "abc"
				SizeEmptyStruct + SizeEmptyStruct + // outer object + reference to its prototype
				(1 + SizeString) + SizeNumber + // "a" and number
				(1 + SizeString) + (4 + SizeString) + // "b" and "1234" string
				// stack difference in going from
				//	[..other, first, 1]
				//  to
				//	[..other, abc, [object Object], 1234]
				18,
		},
		{
			desc: "Proxy",
			script: `var target = {
				message1: "hello",
				message2: "everyone"
			};

			var handler = {
				get: function(target, prop, receiver) {
					return "world";
				}
			};
			var proxy1 = new Proxy(target, handler);

			checkMem();
			proxy2 = new Proxy(target, handler);
			checkMem();
			`,
			expectedSizeDiff: (6 + SizeString) + // "proxy2"
				SizeEmptyStruct + // proxy overhead
				SizeEmptyStruct + SizeEmptyStruct + // base object + prototype
				SizeEmptyStruct + // target object reference
				SizeEmptyStruct, // handler object reference,
			expectedNewSizeDiff: (6 + SizeString) + // "proxy2"
				SizeEmptyStruct + // proxy overhead
				SizeEmptyStruct + SizeEmptyStruct + // base object + prototype
				SizeEmptyStruct + // target object reference
				SizeEmptyStruct, // handler object reference,
		},
		{
			desc: "String",
			script: `str1 = new String("hi")

			checkMem();
			str2 = new String("hello")
			checkMem();
			`,
			expectedSizeDiff: (4 + SizeString) + // "str2"
				(5 + SizeString) + // "hello"
				SizeEmptyStruct + SizeEmptyStruct + // base object + prototype
				(6 + SizeString) + SizeNumber, // "length" + number,
			expectedNewSizeDiff: (4 + SizeString) + // "str2"
				(5 + SizeString) + // "hello"
				SizeEmptyStruct + SizeEmptyStruct + // base object + prototype
				(6 + SizeString) + SizeNumber, // "length" + number,
		},
		{
			desc: "Typed array",
			script: `var ta = new Uint8Array(1);
			checkMem();
			ta2 = new Uint8Array([1, 2, 3, 4]);
			checkMem();
			`,
			expectedSizeDiff: (3 + SizeString) + // "ta2"
				SizeEmptyStruct + // typed array overhead
				SizeEmptyStruct + SizeEmptyStruct + // base object + prototype
				4 + SizeEmptyStruct + SizeEmptyStruct + // array buffer data +  base object + prototype
				SizeEmptyStruct + // default constructor
				SizeInt, // last element (4) on stack,
			expectedNewSizeDiff: (3 + SizeString) + // "ta2"
				SizeEmptyStruct + // typed array overhead
				SizeEmptyStruct + SizeEmptyStruct + // base object + prototype
				4 + SizeEmptyStruct + SizeEmptyStruct + // array buffer data +  base object + prototype
				SizeEmptyStruct + // default constructor
				SizeInt, // last element (4) on stack,
		},
		{
			desc: "ArrayBuffer",
			script: `var buffer = new ArrayBuffer(16);

			checkMem();
			buffer2 = new ArrayBuffer(16);
			checkMem();`,
			expectedSizeDiff: (7 + SizeString) + // "buffer2"
				16 + // data size
				SizeEmptyStruct + SizeEmptyStruct, // base object + prototype,
			expectedNewSizeDiff: (7 + SizeString) + // "buffer2"
				16 + // data size
				SizeEmptyStruct + SizeEmptyStruct, // base object + prototype,
		},
		{
			desc: "DataView",
			script: `var buffer = new ArrayBuffer(16);
			var view = new DataView(buffer, 0);
			var buffer2 = new ArrayBuffer(16);

			checkMem();
			view2 = new DataView(buffer2, 0);
			checkMem();`,
			expectedSizeDiff: (5 + SizeString) + // "view2"
				SizeEmptyStruct + // DataView overhead
				SizeEmptyStruct + SizeEmptyStruct + // base object + prototype
				SizeEmptyStruct, // array buffer reference,
			expectedNewSizeDiff: (5 + SizeString) + // "view2"
				SizeEmptyStruct + // DataView overhead
				SizeEmptyStruct + SizeEmptyStruct + // base object + prototype
				SizeEmptyStruct, // array buffer reference,
		},
		{
			desc: "Number",
			script: `num1 = new Number("1")

			checkMem();
			num2 = new Number("2")
			checkMem();`,
			expectedSizeDiff: (4 + SizeString) + // "num2"
				SizeNumber +
				SizeEmptyStruct + SizeEmptyStruct, // base object + prototype,
			expectedNewSizeDiff: (4 + SizeString) + // "num2"
				SizeNumber +
				SizeEmptyStruct + SizeEmptyStruct, // base object + prototype,
		},
		// TODO(REALMC-10739) add a test that calls Error.captureStackTrace when it is implemented)
		{
			desc: "stash",
			script: `
			// With the new template-object feature, objects are now not initialized on vm start but instead initialized
			// when actually being used. To account for this difference we create a New Error error so that this memory
			// usage is already included in the initial checkMem call.
			const err = new Error();
			checkMem();
			try {
				throw new Error("abc");
			} catch(e) {
				checkMem();
			}
			`,
			expectedSizeDiff: (7 + SizeString) + (3 + SizeString) + // Error "message" field + len("abc")
				(4 + SizeString) + (5 + SizeString) + // Error "name" field + len("Error")
				SizeEmptyStruct + SizeEmptyStruct, // base object + prototype,
			expectedNewSizeDiff: (7 + SizeString) + (3 + SizeString) + // Error "message" field + len("abc")
				(4 + SizeString) + (5 + SizeString) + // Error "name" field + len("Error")
				SizeEmptyStruct + SizeEmptyStruct, // base object + prototype,
		},
		{
			desc: "Native value",
			script: `checkMem();
			nv = new MyNativeVal()
			checkMem();`,
			expectedSizeDiff: testNativeValueMemUsage +
				(2 + SizeString), // "nv",
			expectedNewSizeDiff: testNativeValueMemUsage +
				(2 + SizeString), // "nv",
		},
	} {
		t.Run(fmt.Sprintf(tc.desc), func(t *testing.T) {
			memChecks := []uint64{}
			newMemChecks := []uint64{}
			vm := New()

			vm.Set("checkMem", func(call FunctionCall) Value {
				mem, newMem, err := vm.MemUsage(
					NewMemUsageContext(vm, 100, memUsageLimit, arrLenThreshold, objPropsLenThreshold, TestNativeMemUsageChecker{}),
				)
				if err != nil {
					t.Fatal(err)
				}
				memChecks = append(memChecks, mem)
				newMemChecks = append(newMemChecks, newMem)
				return UndefinedValue()
			})

			nc := vm.CreateNativeClass("MyNativeVal", func(call FunctionCall) interface{} {
				return TestNativeValue{}
			}, nil, nil)
			vm.Set("MyNativeVal", nc.Function)

			_, err := vm.RunString(tc.script)
			if err != nil {
				t.Fatal(err)
			}

			if len(memChecks) < 2 {
				t.Fatalf("expected at least two entries in mem check function, but got %d", len(memChecks))
			}
			if len(newMemChecks) < 2 {
				t.Fatalf("expected at least two entries in new mem check function, but got %d", len(memChecks))
			}

			memDiff := memChecks[len(memChecks)-1] - memChecks[0]
			if memDiff != tc.expectedSizeDiff {
				t.Fatalf("expected memory change to equal %d but got %d instead", tc.expectedSizeDiff, memDiff)
			}
			newMemDiff := newMemChecks[len(newMemChecks)-1] - newMemChecks[0]
			if newMemDiff != tc.expectedNewSizeDiff {
				t.Fatalf("expected new memory change to equal %d but got %d instead", tc.expectedNewSizeDiff, newMemDiff)
			}
		})
	}
}

func TestMemMaxDepth(t *testing.T) {
	for _, tc := range []struct {
		desc          string
		script        string
		expectedDepth int
	}{
		{
			desc:          "nested objects",
			script:        `var x = {"1": {"2": {"3": {"4": {"5": {"6": "abc"}}}}}}`,
			expectedDepth: 6,
		},
		{
			desc: "array",
			script: `var x = []
			x[1] = {"1": {"2": {"3": {"4": {"5": {"6": "abc"}}}}}};`,
			expectedDepth: 7,
		},
		{
			desc: "sparse array (sparseArrayObject)",
			script: `var x = []
			x[5000] = {"1": {"2": {"3": {"4": {"5": {"6": "abc"}}}}}};`,
			expectedDepth: 7,
		},
		{
			desc: "Map",
			script: `var abc = new Map()
			abc.set("obj", {"1": {"2": {"3": {"4": {"5": {"6": "abc"}}}}}});`,
			expectedDepth: 7,
		},
	} {
		t.Run(fmt.Sprintf(tc.desc), func(t *testing.T) {
			vm := New()
			_, err := vm.RunString(tc.script)
			if err != nil {
				t.Fatal(err)
			}

			// All global variables are contained in the Runtime's globalObject field, which causes
			// them to be one level deeper
			_, _, err = vm.MemUsage(
				NewMemUsageContext(vm, tc.expectedDepth, memUsageLimit, arrLenThreshold, objPropsLenThreshold, TestNativeMemUsageChecker{}),
			)
			if err != ErrMaxDepth {
				t.Fatalf("expected mem check to hit depth limit error, but got nil %v", err)
			}

			_, _, err = vm.MemUsage(
				// need to add 2 to the expectedDepth since Object is lazy loaded it adds onto the expected depth
				NewMemUsageContext(vm, tc.expectedDepth+2, memUsageLimit, arrLenThreshold, objPropsLenThreshold, TestNativeMemUsageChecker{}),
			)
			if err != nil {
				t.Fatalf("expected to NOT hit mem check hit depth limit error, but got %v", err)
			}
		})
	}
}

func TestMemArraysWithLenThreshold(t *testing.T) {
	for _, tc := range []struct {
		desc                string
		script              string
		threshold           int
		memLimit            uint64
		expectedSizeDiff    uint64
		expectedNewSizeDiff uint64
	}{
		{
			desc: "array of numbers under length threshold",
			script: `y = []
			y.push([]);
			let i=0;
			checkMem();
			for(i=0;i<20;i++){
				y[0].push(i);
			};
			checkMem()`,
			threshold: 100,
			memLimit:  memUsageLimit,
			expectedSizeDiff: SizeEmptyStruct + // Array overhead
				20*SizeNumber, // size of property values
			expectedNewSizeDiff: SizeEmptyStruct + // Array overhead
				20*SizeNumber, // size of property values,
		},
		{
			desc: "array of numbers over threshold",
			script: `y = []
			y.push([]);
			let i=0;
			checkMem();
			for(i=0;i<200;i++){
				y[0].push(i);
			};
			checkMem()`,
			threshold: 100,
			memLimit:  memUsageLimit,
			expectedSizeDiff: SizeEmptyStruct + // Array overhead
				200*SizeNumber, // size of property values
			expectedNewSizeDiff: SizeEmptyStruct + // Array overhead
				200*SizeNumber, // size of property values,
		},
		{
			desc: "mixed array under threshold",
			script: `y = []
			y.push([]);
			let i = 0;
			checkMem();
			for(i=0;i<100;i++){
				y[0].push(i<50?0:true);
			};
			checkMem()`,
			threshold: 200,
			memLimit:  memUsageLimit,
			expectedSizeDiff: SizeEmptyStruct + // Array overhead
				(50 * SizeNumber) + (50 * SizeBool) + // (450) size of property values
				// stack difference in going from
				//	[..other, []]
				//  to
				//	[..other, true, 50]
				+1,
			expectedNewSizeDiff: SizeEmptyStruct + // Array overhead
				(50 * SizeNumber) + (50 * SizeBool) + // (450) size of property values
				// stack difference in going from
				//	[..other, []]
				//  to
				//	[..other, true, 50]
				+1,
		},
		{
			desc: "array under threshold but over limit",
			script: `y = []
			y.push([]);
			let i = 0;
			checkMem();
			for(i=0;i<10;i++){
				y[0].push(i);
			};
			checkMem()`,
			threshold: 200,
			memLimit:  100,
			// Array overhead, size of property values, only 3 values before we hit the mem limit
			expectedSizeDiff:    SizeEmptyStruct + (3 * SizeNumber),
			expectedNewSizeDiff: SizeEmptyStruct + (3 * SizeNumber),
		},
		{
			desc: "mixed array over threshold",
			script: `y = []
			y.push([]);
			let i = 0;
			checkMem();
			for(i=0;i<100;i++){
				y[0].push(i<50?0:true);
			};
			checkMem()`,
			threshold: 50,
			memLimit:  memUsageLimit,
			expectedSizeDiff: SizeEmptyStruct + // Array overhead
				(50 * SizeNumber) + (50 * SizeBool) + // (450) size of property values
				// stack difference in going from
				//	[..other, []]
				//  to
				//	[..other, true, 50]
				+1,
			expectedNewSizeDiff: SizeEmptyStruct + // Array overhead
				(50 * SizeNumber) + (50 * SizeBool) + // (450) size of property values
				// stack difference in going from
				//	[..other, []]
				//  to
				//	[..other, true, 50]
				+1,
		},
		{
			desc: "mixed scattered array over threshold wcs",
			script: `y = []
			y.push([]);
			let i = 0;
			checkMem();
			for(i=0;i < 100;i++){
				y[0].push(i%10==0?0:true);
			};
			checkMem()`,
			threshold: 50,
			memLimit:  memUsageLimit,
			expectedSizeDiff: SizeEmptyStruct + // Array overhead,
				(100 * SizeNumber) + // size of property values
				// stack difference in going from
				//	[..other, []]
				//  to
				//	[..other, true, 50]
				+1,
			expectedNewSizeDiff: SizeEmptyStruct + // Array overhead,
				(100 * SizeNumber) + // size of property values
				// stack difference in going from
				//	[..other, []]
				//  to
				//	[..other, true, 50]
				+1,
		},
	} {
		t.Run(fmt.Sprintf(tc.desc), func(t *testing.T) {
			memChecks := []uint64{}
			newMemChecks := []uint64{}
			vm := New()

			vm.Set("checkMem", func(call FunctionCall) Value {
				mem, newMem, err := vm.MemUsage(
					NewMemUsageContext(vm, 100, tc.memLimit, tc.threshold, objPropsLenThreshold, TestNativeMemUsageChecker{}),
				)
				if err != nil {
					t.Fatal(err)
				}
				memChecks = append(memChecks, mem)
				newMemChecks = append(newMemChecks, newMem)
				return UndefinedValue()
			})

			nc := vm.CreateNativeClass("MyNativeVal", func(call FunctionCall) interface{} {
				return TestNativeValue{}
			}, nil, nil)
			vm.Set("MyNativeVal", nc.Function)

			_, err := vm.RunString(tc.script)
			if err != nil {
				t.Fatal(err)
			}
			if len(memChecks) < 2 {
				t.Fatalf("expected at least two entries in mem check function, but got %d", len(memChecks))
			}
			if len(newMemChecks) < 2 {
				t.Fatalf("expected at least two entries in new mem check function, but got %d", len(newMemChecks))
			}

			memDiff := memChecks[len(memChecks)-1] - memChecks[0]
			if memDiff != tc.expectedSizeDiff {
				t.Fatalf("expected memory change to equal %d but got %d instead", tc.expectedSizeDiff, memDiff)
			}
			newMemDiff := newMemChecks[len(newMemChecks)-1] - newMemChecks[0]
			if newMemDiff != tc.expectedNewSizeDiff {
				t.Fatalf("expected new memory change to equal %d but got %d instead", tc.expectedNewSizeDiff, newMemDiff)
			}
		})
	}
}

func TestMemObjectsWithPropsLenThreshold(t *testing.T) {
	for _, tc := range []struct {
		desc                string
		script              string
		threshold           int
		memLimit            uint64
		expectedSizeDiff    uint64
		expectedNewSizeDiff uint64
	}{
		{
			desc: "object under threshold",
			script: `y = {}
			let i =0;
			checkMem();
			for (i=0;i<10;i++) {
				y["i"+i] = i
			}
			checkMem()`,
			threshold: 100,
			memLimit:  memUsageLimit,
			// object overhead + len("i0") + value i
			expectedSizeDiff: SizeEmptyStruct + 10*(2+SizeString) + 10*SizeNumber,
			// object overhead + len("i0") + value i
			expectedNewSizeDiff: SizeEmptyStruct + 10*(2+SizeString) + 10*SizeNumber,
		},
		{
			desc: "object under threshold but over limit",
			script: `y = {}
			let i = 0;
			checkMem();
			for (i=0;i<10;i++) {
				y["i"+i] = i
			}
			checkMem()`,
			threshold: 100,
			memLimit:  40,
			// object overhead + len("i0") + value i
			expectedSizeDiff: SizeEmptyStruct + 10*(2+SizeString) + 10*SizeNumber,
			// object overhead + len("i0") + value i
			expectedNewSizeDiff: SizeEmptyStruct + 10*(2+SizeString) + 10*SizeNumber,
		},
		{
			desc: "object over threshold",
			script: `y = {}
			let i = 0;
			checkMem();
			for (i=100;i<200;i++) {
				y["i"+i] = i
			}
			checkMem()`,
			threshold: 75,
			memLimit:  memUsageLimit,
			// object overhead + len("i100") + value i
			expectedSizeDiff: SizeEmptyStruct + 100*(4+SizeString) + 100*SizeNumber,
			// object overhead + len("i100") + value i
			expectedNewSizeDiff: SizeEmptyStruct + 100*(4+SizeString) + 100*SizeNumber,
		},
		{
			desc: "mixed object over threshold",
			script: `y = {}
			let i = 0;
			checkMem();
			for (i=100;i<200;i++) {
				y["i"+i] = i < 150 ? i : "i"+i
			}
			checkMem()`,
			threshold: 200,
			memLimit:  memUsageLimit,
			// object overhead + key len("i100") + value i for the items 100 to 149
			expectedSizeDiff: SizeEmptyStruct + 50*(4+SizeString) + 50*SizeNumber +
				// len("i150") + len("i150") for the items 150 to 199
				50*(4+SizeString) + 50*(4+SizeString) +
				// 150 number + len("i") in "i"+i expression
				3 + (1 + SizeString),
			// object overhead + key len("i100") + value i for the items 100 to 149
			expectedNewSizeDiff: SizeEmptyStruct + 50*(4+SizeString) + 50*SizeNumber +
				// len("i150") + len("i150") for the items 150 to 199
				50*(4+SizeString) + 50*(4+SizeString) +
				// 150 number + len("i") in "i"+i expression
				3 + (1 + SizeString),
		},
	} {
		t.Run(fmt.Sprintf(tc.desc), func(t *testing.T) {
			memChecks := []uint64{}
			newMemChecks := []uint64{}
			vm := New()

			vm.Set("checkMem", func(call FunctionCall) Value {
				mem, newMem, err := vm.MemUsage(
					NewMemUsageContext(vm, 100, tc.memLimit, arrLenThreshold, tc.threshold, TestNativeMemUsageChecker{}),
				)
				if err != nil {
					t.Fatal(err)
				}
				memChecks = append(memChecks, mem)
				newMemChecks = append(newMemChecks, newMem)
				return UndefinedValue()
			})

			nc := vm.CreateNativeClass("MyNativeVal", func(call FunctionCall) interface{} {
				return TestNativeValue{}
			}, nil, nil)
			vm.Set("MyNativeVal", nc.Function)

			_, err := vm.RunString(tc.script)
			if err != nil {
				t.Fatal(err)
			}
			if len(newMemChecks) < 2 {
				t.Fatalf("expected at least two entries in new mem check function, but got %d", len(newMemChecks))
			}

			memDiff := memChecks[len(memChecks)-1] - memChecks[0]
			if memDiff != tc.expectedSizeDiff {
				t.Fatalf("expected memory change to equal %d but got %d instead", tc.expectedSizeDiff, memDiff)
			}
			newMemDiff := newMemChecks[len(newMemChecks)-1] - newMemChecks[0]
			if newMemDiff != tc.expectedNewSizeDiff {
				t.Fatalf("expected new memory change to equal %d but got %d instead", tc.expectedNewSizeDiff, newMemDiff)
			}
		})
	}
}
