package goja

import (
	"fmt"
	"testing"
)

const testNativeValueMemUsage = 100

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

	for _, tc := range []struct {
		description      string
		script           string
		expectedSizeDiff uint64
	}{
		{
			"number",
			`x = []
			x.push(0)
			checkMem()
			x.push(0)
			checkMem()`,
			SizeNumber,
		},
		{
			"boolean",
			`x = []
			x.push(true)
			checkMem()
			x.push(true)
			checkMem()`,
			SizeBool,
		},
		{
			"null",
			`x = []
			x.push(null)
			checkMem()
			x.push(null)
			checkMem()`,
			SizeEmpty,
		},
		{
			"undefined",
			`x = []
			x.push(undefined)
			checkMem()
			x.push(undefined)
			checkMem()`,
			SizeEmpty,
		},
		{
			"string",
			`x = []
			x.push("12345")
			checkMem()
			x.push("12345")
			checkMem()`,
			5,
		},
		{
			"string with multi-byte characters",
			`x = []
			x.push("\u2318")
			checkMem()
			x.push("\u2318")
			checkMem()`,
			3, // single char with 3-byte width
		},
		{
			"nested objects",
			`y = []
			y.push(null)
			checkMem()
			y.push({"a":10, "b":"1234", "c":{}})
			checkMem()`,
			SizeEmpty + SizeEmpty + // outer object + reference to its prototype
				(1 + SizeNumber) + // "a" and number
				(1 + 4) + // "b" and string
				(1 + SizeEmpty + SizeEmpty), //  "c" (object + prototype reference)
		},
		{
			"array of numbers",
			`y = []
			var i = 0;
			y.push([]);
			checkMem();
			for(i=0;i<20;i++){
				y[0].push(i);
			};
			checkMem()`,
			// Array overhead,
			// size of property values,
			SizeEmpty + 20*SizeNumber,
		},
		{
			"overhead of a single new scope",
			`checkMem();
			(function(){
				checkMem();
			})();`, // over
			emptyFunctionScopeOverhead,
		},
		{
			"previous function scopes should not affect the current memory",
			`checkMem();
			(function(){
			})();
			checkMem();`,
			0,
		},
		{
			"overhead of each scope is equivalent regardless of depth",
			`checkMem();
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
			emptyFunctionScopeOverhead * 6,
		},
		{"values attached to lexical scope in a function",
			`checkMem();
			(function(){
				var zzzx = 10;
				checkMem();
			})();`,
			// function overhead plus the number value of the "zzzx" property and its string name
			// functionOverhead + SizeNumber + 4,
			emptyFunctionScopeOverhead + SizeNumber,
		},
		{
			"cyclical data structure",
			// cyclical data structure does not recurse infinitely
			// and does not artificially inflate mem count. The only change in mem
			// between the two checks is for the new property names for "y" and "x".
			`var zzza = {}
			 var zzzb = {}
			 checkMem();
			 zzza.y = zzzb
			 zzzb.x = zzza
			 checkMem()`,
			2 + SizeEmpty + SizeEmpty, // "x" and "y" property names + references to each object
		},
		{
			"sparse array (arrayObject)",
			`x = []
			x[1] = "abcd";
			checkMem()
			x[10] = "abc";
			checkMem()`,
			3,
		},
		{
			"sparse array (sparseArrayObject)",
			`x = []
			x[5000] = "abcd";
			checkMem()
			x[5001] = "abc";
			checkMem()`,
			SizeInt32 + 3,
		},
		{
			"array with non-numeric keys",
			`x = []
			x["a"] = 3;
			checkMem()
			x[2] = "abc";
			x["c"] = 3;
			checkMem()
			`,
			// len("abc") + len("a") + SizeNumber
			3 + 1 + SizeNumber,
		},
		{
			"reference to array",
			`x = []
			x[1] = "abcd";
			x[10] = "abc";
			checkMem()
			y = x;
			checkMem()`,
			// len("y") + reference to array
			1 + SizeEmpty,
		},
		{
			"reference to sparse array",
			`x = []
			x[5000] = "abcb";
			x[5001] = "abc";
			checkMem()
			y = x;
			// len("y") + reference to array
			checkMem()`,
			1 + SizeEmpty,
		},
		{
			"Date object",
			`
			d1 = new Date();
			checkMem();
			d2 = new Date();
			checkMem()
			`,
			// len("d2") + size of msec + reference to visited base object + base object prototype reference
			2 + SizeNumber + SizeEmpty + SizeEmpty,
		},
		{
			"Empty object",
			`
			checkMem();
			o = {}
			checkMem()
			`,
			// len("o") + object's starting size + reference to prototype
			1 + SizeEmpty + SizeNumber,
		},
		{
			"Map",
			`var m = new Map();
			m.set("a", 1);
			checkMem();
			m.set("abc", {"a":10, "b":"1234"});
			checkMem();`,
			3 + // "abc"
				SizeEmpty + SizeEmpty + // outer object + reference to its prototype
				(1 + SizeNumber) + // "a" and number
				(1 + 4), // "b" and string
		},
		{
			"Proxy",
			`var target = {
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
			6 + // "proxy2"
				SizeEmpty + // proxy overhead
				SizeEmpty + SizeEmpty + // base object + prototype
				SizeEmpty + // target object reference
				SizeEmpty, // handler object reference
		},
		{
			"String",
			`str1 = new String("hi")

			checkMem();
			str2 = new String("hello")
			checkMem();
			`,
			4 + // "str2"
				5 + // "hello"
				SizeEmpty + SizeEmpty + // base object + prototype
				6 + SizeNumber, // "length" + number
		},
		{
			"Typed array",
			`var ta = new Uint8Array(1);
			checkMem();
			ta2 = new Uint8Array([1, 2, 3, 4]),
			checkMem();
			`,
			3 + // "ta2"
				SizeEmpty + // typed array overhead
				SizeEmpty + SizeEmpty + // base object + prototype
				4 + SizeEmpty + SizeEmpty + // array buffer data +  base object + prototype
				SizeEmpty, // default constructor
		},
		{
			"ArrayBuffer",
			`var buffer = new ArrayBuffer(16);

			checkMem();
			buffer2 = new ArrayBuffer(16);
			checkMem();`,
			7 + // "buffer2"
				16 + // data size
				SizeEmpty + SizeEmpty, // base object + prototype
		},
		{
			"DataView",
			`var buffer = new ArrayBuffer(16);
			var view = new DataView(buffer, 0);
			var buffer2 = new ArrayBuffer(16);

			checkMem();
			view2 = new DataView(buffer2, 0);
			checkMem();`,
			5 + // "view2"
				SizeEmpty + // DataView overhead
				SizeEmpty + SizeEmpty + // base object + prototype
				SizeEmpty, // array buffer reference
		},
		{
			"Number",
			`num1 = new Number("1")

			checkMem();
			num2 = new Number("2")
			checkMem();`,
			4 + // "num2"
				SizeNumber +
				SizeEmpty + SizeEmpty, // base object + prototype
		},
		{
			"stash",
			`checkMem();
			try {
				throw new Error("abc");
			} catch(e) {
				checkMem();
			}
			`,
			7 + 3 + // Error "message" field + len("abc")
				4 + 5 + // Error "name" field + len("Error")
				SizeEmpty + SizeEmpty, // base object + prototype
		},
		{
			"Native value",
			`checkMem();
			nv = new MyNativeVal()
			checkMem();`,
			testNativeValueMemUsage + 2,
		},
	} {
		t.Run(fmt.Sprintf(tc.description), func(t *testing.T) {
			memChecks := []uint64{}
			vm := New()
			vm.Set("checkMem", func(call FunctionCall) Value {
				mem, err := vm.MemUsage(NewMemUsageContext(vm, 100, TestNativeMemUsageChecker{}))
				if err != nil {
					t.Fatal(err)
				}
				memChecks = append(memChecks, mem)
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

			memDiff := memChecks[len(memChecks)-1] - memChecks[0]
			if memDiff != tc.expectedSizeDiff {
				t.Fatalf("expected memory change to equal %d but got %d instead", tc.expectedSizeDiff, memDiff)
			}
		})
	}
}

func TestMemMaxDepth(t *testing.T) {
	for _, tc := range []struct {
		description   string
		script        string
		expectedDepth int
	}{
		{
			"nested objects",
			`var x = {"1": {"2": {"3": {"4": {"5": {"6": "abc"}}}}}}`,
			6,
		},
		{
			"array",
			`var x = []
			x[1] = {"1": {"2": {"3": {"4": {"5": {"6": "abc"}}}}}};`,
			7,
		},
		{
			"sparse array (sparseArrayObject)",
			`var x = []
			x[5000] = {"1": {"2": {"3": {"4": {"5": {"6": "abc"}}}}}};`,
			7,
		},
		{
			"Map",
			`var abc = new Map()
			abc.set("obj", {"1": {"2": {"3": {"4": {"5": {"6": "abc"}}}}}});`,
			7,
		},
	} {
		t.Run(fmt.Sprintf(tc.description), func(t *testing.T) {
			vm := New()
			_, err := vm.RunString(tc.script)
			if err != nil {
				t.Fatal(err)
			}

			// All global variables are contained in the Runtime's globalObject field, which causes
			// them to be one level deeper
			_, err = vm.MemUsage(NewMemUsageContext(vm, tc.expectedDepth, TestNativeMemUsageChecker{}))
			if err != ErrMaxDepth {
				t.Fatalf("expected mem check to hit depth limit error, but got nil %v", err)
			}

			_, err = vm.MemUsage(NewMemUsageContext(vm, tc.expectedDepth+1, TestNativeMemUsageChecker{}))
			if err != nil {
				t.Fatalf("expected to NOT hit mem check hit depth limit error, but got %v", err)
			}
		})
	}
}
