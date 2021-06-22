package goja

import (
	"bytes"
	"errors"

	"github.com/dop251/goja/unistring"
)

type Property struct {
	Name  string
	Value Value
}

type NativeClass struct {
	*Object
	runtime    *Runtime
	classProto *Object
	className  string
	classProps []Property
	funcProps  []Property

	Function *Object

	getStacktrace  func(err error) string
	initStacktrace func(err error, stacktrace string)
}

func (r *Runtime) TryToValue(i interface{}) (Value, error) {
	var result Value
	err := r.vm.try(r.vm.ctx, func() {
		result = r.ToValue(i)
	})
	if err.Error() == "" || err.Error() == "<nil>" {
		return result, nil
	}
	return result, err
}

func (r *Runtime) MakeCustomError(name, msg string) *Object {
	e := r.newError(r.global.Error, msg).(*Object)
	e.self.setOwnStr("name", asciiString(name), false)
	return e
}

// CreateNativeErrorClass creates a native class of the given name that inherits from type Error.
// Required is a constructor that will be called upon calling the function (with or without 'new'),
// and getter/setter functions for the stacktrace.
// Optional arguments are the properties that should exist on the class and function
// respectively.
func (r *Runtime) CreateNativeErrorClass(
	className string,
	ctor func(call FunctionCall) error,
	initStacktrace func(err error, stacktrace string),
	getStacktrace func(err error) string,
	classProps []Property,
	funcProps []Property,
) NativeClass {
	classProto := r.builtin_new(r.global.Error, []Value{})
	proto := classProto.self
	for _, prop := range classProps {
		proto._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	}

	classFunc := r.newNativeFuncConstruct(func(args []Value, proto *Object) *Object {
		obj := r.newBaseObject(proto, className)
		call := FunctionCall{
			ctx:       r.vm.ctx,
			This:      obj.val,
			Arguments: args,
		}

		err := ctor(call)
		errMsg := newStringValue(err.Error())
		obj._putProp("message", errMsg, true, false, true)
		obj._putProp("name", asciiString(className), true, false, true)

		g := &_goNativeValue{baseObject: obj, value: err}
		obj.val.self = g
		obj.val.__wrapped = g.value

		ex := &Exception{
			val:        obj.val,
			traceLimit: r.stackTraceLimit,
		}
		stackTrace := bytes.NewBuffer(nil)
		ex.writeShortStack(stackTrace)
		initStacktrace(err, stackTrace.String())

		return obj.val
	}, unistring.String(className), classProto, 1)

	for _, prop := range funcProps {
		classFunc.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	}

	return NativeClass{Object: classFunc, runtime: r, classProto: classProto, className: className, Function: classFunc, getStacktrace: getStacktrace, initStacktrace: initStacktrace}
}

func (r *Runtime) CreateNativeError(name string) (Value, func(err error) Value) {
	proto := r.builtin_new(r.global.Error, []Value{})
	o := proto.self
	o._putProp("name", asciiString(name), true, false, true)

	e := r.newNativeFuncConstructProto(r.builtin_Error, unistring.String(name), proto, r.global.Error, 1)

	return e, func(err error) Value {
		return r.MakeCustomError(name, err.Error())
	}
}

func (r *Runtime) CreateNativeClass(
	className string,
	ctor func(call FunctionCall) interface{},
	classProps []Property,
	funcProps []Property,
) NativeClass {
	classProto := r.builtin_new(r.global.Object, []Value{})
	proto := classProto.self
	for _, prop := range classProps {
		proto._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	}

	classFunc := r.newNativeFuncConstruct(func(args []Value, proto *Object) *Object {
		obj := r.newBaseObject(proto, className)

		call := FunctionCall{
			ctx:       r.vm.ctx,
			This:      obj.val,
			Arguments: args,
		}
		val := ctor(call)
		g := &_goNativeValue{baseObject: obj, value: val}
		obj.val.self = g
		obj.val.__wrapped = g.value
		return obj.val
	}, unistring.String(className), classProto, 1)

	classFunc.self._putProp("name", asciiString(className), true, false, true)
	for _, prop := range funcProps {
		classFunc.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	}

	return NativeClass{
		classProto: classProto,
		className:  className,
		classProps: classProps,
		funcProps:  funcProps,
		Function:   classFunc,
		runtime:    r,
	}
}

type _goNativeValue struct {
	*baseObject
	value interface{}
}

func (n NativeClass) InstanceOf(val interface{}) Value {
	r := n.runtime
	className := n.className
	classProto := n.classProto
	obj, err := r.New(r.newNativeFuncConstruct(func(args []Value, proto *Object) *Object {
		obj := r.newBaseObject(proto, className)
		g := &_goNativeValue{baseObject: obj, value: val}
		obj.val.self = g
		obj.val.__wrapped = g.value
		return obj.val
	}, unistring.String(className), classProto, 1))
	if err != nil {
		panic(err)
	}

	obj.self._putProp("name", asciiString(n.className), true, false, true)
	for _, prop := range n.funcProps {
		obj.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	}

	if err, ok := val.(error); ok {
		obj.self._putProp("message", newStringValue(err.Error()), true, false, true)
		if n.getStacktrace != nil {
			stackTrace := n.getStacktrace(err)
			if len(stackTrace) == 0 {
				ex := &Exception{
					val:        obj,
					traceLimit: r.stackTraceLimit,
				}
				ex.stack = r.vm.captureStack(ex.stack, 0)

				stackTrace = ex.String()
			}
			n.initStacktrace(err, stackTrace)
		}
	}
	return obj
}

// CreateNativeFunction creates a native function that will call the given call function.
// This provides for a way to detail how the function appears to a user within JS
// compared to passing the call in via toValue.
func (r *Runtime) CreateNativeFunction(name, file string, call func(FunctionCall) Value) (Value, error) {
	if call == nil {
		return UndefinedValue(), errors.New("call cannot be nil")
	}

	return r.newNativeFunc(call, nil, unistring.String(name), nil, 1), nil
}

func (r *Runtime) Eval(name, src string, direct, strict bool) (Value, error) {
	this := r.NewObject()

	p, err := r.compile(name, src, strict, true)
	if err != nil {
		panic(err)
	}

	vm := r.vm

	vm.pushCtx()
	vm.prg = p
	vm.pc = 0
	if !direct {
		vm.stash = nil
	}
	vm.sb = vm.sp
	vm.push(this)
	if strict {
		vm.push(valueTrue)
	} else {
		vm.push(valueFalse)
	}
	vm.run()
	vm.popCtx()
	vm.halt = false
	retval := vm.stack[vm.sp-1]
	vm.sp -= 2
	return retval, nil
}
