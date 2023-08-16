package goja

import (
	"testing"
)

func is(t *testing.T, actual, expected interface{}) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected %v (%T), but got %v (%T)", expected, expected, actual, actual)
	}
}

func TestNativeClass(t *testing.T) {
	vm := New()

	ctor := func(call FunctionCall) interface{} {
		call.This.ToObject(vm.vm.r).Set("woof", 2)
		arg1 := call.Argument(0).Export()
		arg2 := call.Argument(1).Export()
		if str2, ok := arg2.(string); ok {
			str1 := arg1.(string)
			return str1 + str2
		}
		return arg1
	}
	hello := vm.ToValue(func(call FunctionCall) Value {
		return newStringValue("hello world")
	})
	toString := vm.ToValue(func(call FunctionCall) Value {
		exp := call.This.Export()
		return newStringValue(exp.(string))
	})

	var blessedCtor func(value interface{}) Value
	makeThisThing := vm.ToValue(func(call FunctionCall) Value {
		return blessedCtor("never_change")
	})

	cls := vm.CreateNativeClass(
		"Carrot",
		ctor,
		[]Property{
			{
				Name:  "doHello",
				Value: hello,
			},
			{
				Name:  "staticValue",
				Value: valueInt32(42),
			},
			{
				Name:  "toString",
				Value: toString,
			},
		},
		[]Property{
			{
				Name:  "makeThisThing",
				Value: makeThisThing,
			},
		},
	)
	blessedCtor = cls.InstanceOf

	_, err := vm.RunString(`Carrot("yum")`)
	if err == nil {
		t.Fatal("expected an error")
	}

	vm.Set("Carrot", cls.Function)
	ret, err := vm.RunString(`Carrot("yum")`)
	is(t, err, nil)
	is(t, ret.ToObject(vm.vm.r).__wrapped, "yum")

	ret, err = vm.RunString(`new Carrot("yum", "hmm")`)
	is(t, err, nil)
	is(t, ret.ToObject(vm.vm.r).__wrapped, "yumhmm")

	ret, err = vm.RunString(`Carrot("yum").doHello()`)
	is(t, err, nil)
	is(t, ret.String(), "hello world")

	ret, err = vm.RunString(`Carrot("yum").staticValue`)
	is(t, err, nil)
	is(t, ret.ToInt(), 42)

	ret, err = vm.RunString(`(new Carrot("yum")).woof`)
	is(t, err, nil)
	is(t, ret.ToInt(), 2)

	ret, err = vm.RunString(`Carrot.makeThisThing("yummy")`)
	is(t, err, nil)
	is(t, ret.ToObject(vm.vm.r).__wrapped, "never_change")

	ret, err = vm.RunString(`Carrot.makeThisThing("yummy") instanceof Carrot`)
	is(t, err, nil)
	is(t, ret.ToBoolean(), true)

	ret, err = vm.RunString(`Carrot("yum").toString()`)
	is(t, err, nil)
	is(t, ret.String(), "yum")

	ret, err = vm.RunString(`Carrot("yum") instanceof Carrot`)
	is(t, err, nil)
	is(t, ret.ToBoolean(), true)

	ret, err = vm.RunString(`5 instanceof Carrot`)
	is(t, err, nil)
	is(t, ret.ToBoolean(), false)

	ret, err = vm.RunString(`Carrot.name`)
	is(t, err, nil)
	is(t, ret.String(), "Carrot")

	ret, err = vm.RunString(`Carrot("yum").name`)
	is(t, err, nil)
	is(t, ret.String(), "undefined")
}

func TestNewLazyObject(t *testing.T) {
	t.Run("should return an error if no callback is supplied to NewLazyObject", func(t *testing.T) {
		vm := New()
		_, err := vm.NewLazyObject(nil)
		is(t, err.Error(), "create cannot be nil")
	})

	t.Run("creating new lazy object should not allocate size when not accessed", func(t *testing.T) {
		vm := New()
		o, err := vm.NewLazyObject(func(val *Object) *Object {
			o := vm.newBaseObject(nil, "myClass")
			o._putProp("myProp", newStringValue("hello world"), false, false, false)
			return o.val
		})
		is(t, err, nil)

		mem, newMem, err := o.MemUsage(NewMemUsageContext(vm, 100, 100, 100, 1, nil))
		is(t, newMem, SizeEmptyStruct*2) // memUsage of Object + lazyObject
		is(t, mem, SizeEmptyStruct*2)    // memUsage of Object + lazyObject
	})

	t.Run("accessing a lazy object should allocate memory and initialize the object", func(t *testing.T) {
		vm := New()
		var didInitialize bool
		o, err := vm.NewLazyObject(func(val *Object) *Object {
			didInitialize = true
			o := vm.newBaseObject(nil, "myClass")
			o._putProp("myProp", newStringValue("hello world"), false, false, false)
			return o.val
		})
		is(t, err, nil)

		memBeforeAccess, newMemBeforeAccess, err := o.MemUsage(NewMemUsageContext(vm, 100, 100, 100, 1, nil))
		is(t, err, nil)
		is(t, memBeforeAccess, SizeEmptyStruct*2)    // memUsage of Object + lazyObject
		is(t, newMemBeforeAccess, SizeEmptyStruct*2) // memUsage of Object + lazyObject

		obj, ok := o.(*Object)
		is(t, ok, true)

		// access the lazyobject which causes the lazyObject's init callback to run
		hasMyProp := obj.hasOwnProperty(newStringValue("myProp"))
		is(t, hasMyProp, true)
		is(t, didInitialize, true)

		mem, newMem, err := o.MemUsage(NewMemUsageContext(vm, 100, 100, 100, 1, nil))
		is(t, err, nil)
		if mem <= memBeforeAccess || newMem <= newMemBeforeAccess {
			t.Fatal("memory usage of lazy object should have grown after accessing it but it didnt")
		}
	})
}
