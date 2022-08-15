package goja

import (
	"testing"
)

func TestNumberEquality(t *testing.T) {
	vm := New()

	res, err := vm.RunString(`var a = new Number('5') 
	a`)
	if err != nil {
		t.Fatal(err)
	}
	if !valueInt(5).Equals(res) {
		t.Fatal("values are not equal")
	}
}

func TestInt64SameAsFloat(t *testing.T) {
	if !valueInt64(5).SameAs(valueFloat(5.0)) {
		t.Fatal("values are not equal")
	}

	if !valueInt64(0).SameAs(valueFloat(0.0)) {
		t.Fatal("values are not equal")
	}

	if !valueInt64(0).SameAs(valueFloat(-0.0)) {
		t.Fatal("values are not equal")
	}
}

func TestIntStringEquality(t *testing.T) {
	vm := New()

	res, err := vm.RunString(`"0"==0`)
	if err != nil {
		t.Fatal(err)
	}
	if !valueBool(true).Equals(res) {
		t.Fatal("values are not equal")
	}

	res, err = vm.RunString(`"0.0"===0`)
	if err != nil {
		t.Fatal(err)
	}
	if !valueBool(false).Equals(res) {
		t.Fatal("values should not be equal")
	}
}
