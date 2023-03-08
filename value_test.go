package goja

import (
	"testing"
)

func TestIntSameAsInt(t *testing.T) {
	if !valueInt(5).SameAs(valueInt(5)) {
		t.Fatal("values are not equal")
	}
}

func TestIntStrictEqualsInt64(t *testing.T) {
	if !valueInt(5).StrictEquals(valueInt64(5)) {
		t.Fatal("values are not equal")
	}
}

func TestIntStrictEqualsFloat(t *testing.T) {
	if !valueInt(5).StrictEquals(valueFloat(5.0)) {
		t.Fatal("values are not equal")
	}
}

func TestIntZeroStrictEqualsFloatZero(t *testing.T) {
	if !valueInt(0).StrictEquals(valueFloat(0.0)) {
		t.Fatal("values are not equal")
	}
}

func TestFloatArrayIncludes(t *testing.T) {
	vm := New()
	res, err := vm.RunString(`[0.0].includes(0)`)
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}
	if !res.SameAs(valueBool(true)) {
		t.Fatal("value not found in array")
	}

	res, err = vm.RunString(`[0.0].includes(-0)`)
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}
	if !res.SameAs(valueBool(true)) {
		t.Fatal("value not found in array")
	}

	res, err = vm.RunString(`[0].includes(0.0)`)
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}
	if !res.SameAs(valueBool(true)) {
		t.Fatal("value not found in array")
	}
}
