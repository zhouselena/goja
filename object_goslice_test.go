package goja

import "testing"

func TestGoSliceBasic(t *testing.T) {
	const SCRIPT = `
	var sum = 0;
	for (var i = 0; i < a.length; i++) {
		sum += a[i];
	}
	sum;
	`
	r := New()
	r.Set("a", []interface{}{1, 2, 3, 4})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}
	if i := v.ToInt(); i != 10 {
		t.Fatalf("Expected 10, got: %d", i)
	}
}

func TestGoSliceIn(t *testing.T) {
	const SCRIPT = `
	var idx = "";
	for (var i in a) {
		idx += i;
	}
	idx;
	`
	r := New()
	r.Set("a", []interface{}{1, 2, 3, 4})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}
	if i := v.String(); i != "0123" {
		t.Fatalf("Expected '0123', got: '%s'", i)
	}
}

func TestGoSliceExpand(t *testing.T) {
	const SCRIPT = `
	var l = a.length;
	for (var i = 0; i < l; i++) {
		a[l + i] = a[i] * 2;
	}

	var sum = 0;
	for (var i = 0; i < a.length; i++) {
		sum += a[i];
	}
	sum;
	`
	r := New()
	a := []interface{}{1, 2, 3, 4}
	r.Set("a", &a)
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}
	sum := 0
	for _, v := range a {
		sum += v.(int)
	}
	if i := v.ToInt(); i != sum {
		t.Fatalf("Expected %d, got: %d", sum, i)
	}
}

func TestGoSliceProto(t *testing.T) {
	const SCRIPT = `
	a.join(",")
	`

	r := New()
	a := []interface{}{1, 2, 3, 4}
	r.Set("a", a)
	ret, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}
	if s := ret.String(); s != "1,2,3,4" {
		t.Fatalf("Unexpected result: '%s'", s)
	}
}
