package goja

import (
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
)

func TestObject_(t *testing.T) {
	vm := New()

	for _, tt := range []struct {
		js       string
		expected interface{}
	}{
		{
			`
            var abc = Object.getOwnPropertyDescriptor(Object, "prototype");
            [ [ typeof Object.prototype, abc.writable, abc.enumerable, abc.configurable ],
            ];
		`, "object,false,false,false",
		},
	} {
		actual, err := vm.RunString(tt.js)
		if err != nil {
			t.Fatal(err)
		}
		if failed := gocmp.Diff(actual.String(), tt.expected); failed != "" {
			t.Fatal(failed)
		}
	}
}

func TestObject_new(t *testing.T) {
	vm := New()

	for _, tt := range []struct {
		js       string
		expected interface{}
	}{
		{
			`
            [ new Object("abc"), new Object(2+2) ];
        `, "abc,4",
		},
	} {
		actual, err := vm.RunString(tt.js)
		if err != nil {
			t.Fatal(err)
		}
		if failed := gocmp.Diff(actual.String(), tt.expected); failed != "" {
			t.Fatal(failed)
		}
	}
}

func TestObject_keys(t *testing.T) {
	vm := New()

	for _, tt := range []struct {
		js       string
		expected interface{}
	}{
		{
			`Object.keys({ abc:undefined, def:undefined })`,
			"abc,def",
		},
		{
			`
            function abc() {
                this.abc = undefined;
                this.def = undefined;
            }
            Object.keys(new abc())
		`,
			"abc,def",
		},
		{
			`
	        function def() {
	            this.ghi = undefined;
	        }
	        def.prototype = new abc();
	        Object.keys(new def());
		`,
			"ghi",
		},
		{
			` (function(abc, def, ghi){
	return Object.keys(arguments)
})(undefined, undefined);
`,
			"0,1",
		},
		{
			`
	        (function(abc, def, ghi){
	            return Object.keys(arguments)
	        })(undefined, undefined, undefined, undefined);
		`,
			"0,1,2,3",
		},
	} {
		actual, err := vm.RunString(tt.js)
		if err != nil {
			t.Fatal(err)
		}
		if failed := gocmp.Diff(actual.String(), tt.expected); failed != "" {
			t.Fatal(failed)
		}
	}
}

func TestObject_values(t *testing.T) {
	vm := New()

	for _, tt := range []struct {
		js       string
		expected interface{}
	}{
		{
			`Object.values({ k1: 'abc', k2 :'def' })`, "abc,def",
		},

		{
			`
						function abc() {
							this.k1 = "abc";
							this.k2 = "def";
						}
						Object.values(new abc());
				`, "abc,def",
		},

		{
			`
						function def() {
							this.k3 = "ghi";
						}
						def.prototype = new abc();
						Object.values(new def());
				`, "ghi",
		},

		{
			`
						var ghi = Object.create(
                {
                    k1: "abc",
                    k2: "def"
                },
                {
                    k3: { value: "ghi", enumerable: true },
                    k4: { value: "jkl", enumerable: false }
                }
            );
            Object.values(ghi);
				`, "ghi",
		},

		{
			`
            (function(abc, def, ghi){
                return Object.values(arguments)
            })(0, 1);
        `, "0,1",
		},

		{
			`
            (function(abc, def, ghi){
                return Object.values(arguments)
            })(0, 1, 2, 3);
        `, "0,1,2,3",
		},
	} {
		actual, err := vm.RunString(tt.js)
		if err != nil {
			t.Fatal(err)
		}
		if failed := gocmp.Diff(actual.String(), tt.expected); failed != "" {
			t.Fatal(failed)
		}
	}
}

func TestObject_entries(t *testing.T) {
	vm := New()

	for _, tt := range []struct {
		js       string
		expected interface{}
	}{
		{
			`Object.entries({ k1: 'abc', k2 :'def' })`, "k1,abc,k2,def",
		},

		{
			`
 			      var e = Object.entries({ k1: 'abc', k2 :'def' });
 						[ e[0][0], e[0][1], e[1][0], e[1][1], ];
 				 `, "k1,abc,k2,def",
		},
		{
			`
 						function abc() {
 							this.k1 = "abc";
 							this.k2 = "def";
 						}
 						Object.entries(new abc());
 				`, "k1,abc,k2,def",
		},
		{
			`
 						function def() {
 							this.k3 = "ghi";
 						}
 						def.prototype = new abc();
 						Object.entries(new def());
				 `,
			"k3,ghi",
		},
		{
			`
 						var ghi = Object.create(
 		            {
 		                k1: "abc",
 		                k2: "def"
 		            },
 		            {
 		                k3: { value: "ghi", enumerable: true },
 		                k4: { value: "jkl", enumerable: false }
 		            }
 		        );
 		        Object.entries(ghi);
				 `,
			"k3,ghi",
		},
		{
			`
 		        (function(abc, def, ghi){
 		            return Object.entries(arguments)
 		        })(0, 1);
 		    `, "0,0,1,1",
		},
		{
			`
 		        (function(abc, def, ghi){
 		            return Object.entries(arguments)
 		        })(0, 1, 2, 3);
 		    `, "0,0,1,1,2,2,3,3",
		},
	} {
		actual, err := vm.RunString(tt.js)
		if err != nil {
			t.Fatal(err)
		}
		if failed := gocmp.Diff(actual.String(), tt.expected); failed != "" {
			t.Fatal(failed)
		}
	}
}

func TestObject_fromEntries(t *testing.T) {
	vm := New()

	for _, tt := range []struct {
		js       string
		expected interface{}
	}{
		{
			`
 					 var o = Object.fromEntries([['a', 1], ['b', true], ['c', 'sea']]);
 					 [ o.a, o.b, o.c ]
 				 `, "1,true,sea",
		},
	} {
		actual, err := vm.RunString(tt.js)
		if err != nil {
			t.Fatal(err)
		}
		if failed := gocmp.Diff(actual.String(), tt.expected); failed != "" {
			t.Fatal(failed)
		}
	}
}
