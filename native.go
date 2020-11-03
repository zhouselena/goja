package goja

import (
	"bytes"
	"errors"
	"fmt"

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
	ctor       func(call FunctionCall) interface{}
	classProps []Property
	funcProps  []Property

	Function *Object //func(call ConstructorCall) *Object
	// runtime *Runtime

	// name string
	// // safe to panic inside these
	// methods     map[string]Value
	// funcProps   map[string]Value
	// newInstance func(call ConstructorCall) interface{}
}

func (r *Runtime) TryToValue(i interface{}) (Value, error) {
	var result Value
	err := r.vm.try(r.ctx, func() {
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
	e.self._putProp("customerror", TrueValue(), false, false, false)
	return e
}

func (r *Runtime) CreateNativeErrorClass(
	className string,
	ctor func(call FunctionCall) error,
	initStacktrace func(err error, stacktrace string),
	getStacktrace func(err error) string,
	classProps []Property,
	funcProps []Property,
) NativeClass {
	//TODO goja handle getStackTrace
	classProto := r.builtin_new(r.global.Error, []Value{})
	o := classProto.self
	o._putProp("name", asciiString(className), true, false, true)

	for _, prop := range classProps {
		o._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	}

	var errMsg valueString
	v := r.newNativeFuncConstruct(func(args []Value, proto *Object) *Object {
		obj := r.newBaseObject(proto, className)
		call := FunctionCall{
			ctx:       r.vm.ctx,
			This:      obj.val,
			Arguments: args,
		}

		err := ctor(call)
		ex := &Exception{
			val: r.newError(r.global.ReferenceError, err.Error()),
		}
		stackTrace := bytes.NewBuffer(nil)
		ex.writeShortStack(stackTrace)
		initStacktrace(err, stackTrace.String())
		errMsg = newStringValue(err.Error())
		obj._putProp("message", errMsg, true, false, true)

		g := &_goNativeValue{baseObject: obj, value: err}
		obj.val.self = g
		obj.val.__wrapped = g.value

		return obj.val
	}, unistring.String(className), classProto, 1)

	// o := r.newNativeFuncObj(val, r.builtin_date, r.builtin_newDate, "Date", r.global.DatePrototype, 7)
	v.self._putProp("message", newStringValue("holy moly"), true, false, true)
	for _, prop := range funcProps {
		// obj.propNames = append(obj.propNames, unistring.String(prop.Name))
		v.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	}
	v.runtime = r
	// o._putProp("parse", r.newNativeFunc(r.date_parse, nil, "parse", nil, 1), true, false, true)
	// o._putProp("UTC", r.newNativeFunc(r.date_UTC, nil, "UTC", nil, 7), true, false, true)
	// o._putProp("now", r.newNativeFunc(r.date_now, nil, "now", nil, 0), true, false, true)

	return NativeClass{Object: v, runtime: r, classProto: classProto, className: className, Function: v}
}

func (r *Runtime) CreateNativeError(name string) (Value, func(err error) Value) {
	proto := r.builtin_new(r.global.Error, []Value{})
	o := proto.self
	o._putProp("name", asciiString(name), true, false, true)

	e := r.newNativeFuncConstructProto(r.builtin_Error, unistring.String(name), proto, r.global.Error, 1)

	return e, func(err error) Value {
		return r.MakeCustomError(name, err.Error())
		// fmt.Println("trying to create new error", err)
		// obj := r.newError(r.global.Error, err.Error()).(*Object)
		// obj.self.setOwnStr("name", asciiString(name), false)
		// f, x := obj.Get("message")
		// fmt.Println("message of error is ", f, x)
		// // obj, err := r.New(e, name, proto, 1))
		// // if err != nil {
		// // 	panic(err)
		// // }
		// return obj
	}
}

func (r *Runtime) CreateNativeClass(
	className string,
	ctor func(call FunctionCall) interface{},
	classProps []Property,
	funcProps []Property,
) NativeClass {
	// needed for instance of
	classProto := r.builtin_new(r.global.Object, []Value{})
	o := classProto.self
	o._putProp("name", asciiString(className), true, false, true)

	for _, prop := range classProps {
		o._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	}

	// v := r.newNativeFuncConstruct(func(args []Value, proto *Object) *Object {
	// 	obj := r.newBaseObject(proto, className)
	// 	call := FunctionCall{
	// 		ctx:       r.vm.ctx,
	// 		This:      obj.val,
	// 		Arguments: args,
	// 	}

	// 	g := &_goNativeValue{baseObject: obj, value: ctor(call)}
	// 	obj.val.self = g
	// 	obj.val.__wrapped = g.value

	// 	return obj.val
	// }, className, classProto, 1)

	// v := r.newNativeFuncConstruct(func(args []Value, proto *Object) *Object {
	// 	obj := r.newBaseObject(proto, className)
	// 	call := FunctionCall{
	// 		ctx:       r.vm.ctx,
	// 		This:      obj.val,
	// 		Arguments: args,
	// 	}

	// 	val := ctor(call)
	// 	// ex := &Exception{
	// 	// 	val: r.newError(r.global.ReferenceError, err.Error()),
	// 	// }
	// 	// stackTrace := bytes.NewBuffer(nil)
	// 	// ex.writeShortStack(stackTrace)
	// 	// initStacktrace(err, stackTrace.String())
	// 	// errMsg = newStringValue(err.Error())
	// 	// obj._putProp("message", errMsg, true, false, true)

	// 	g := &_goNativeValue{baseObject: obj, value: val}
	// 	obj.val.self = g
	// 	obj.val.__wrapped = g.value

	// 	return obj.val
	// }, className, classProto, 1)

	// o := r.newNativeFuncObj(val, r.builtin_date, r.builtin_newDate, "Date", r.global.DatePrototype, 7)
	// v.self._putProp("message", newStringValue("holy moly"), true, false, true)
	// for _, prop := range funcProps {
	// 	// obj.propNames = append(obj.propNames, unistring.String(prop.Name))
	// 	v.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	// }

	// // o := r.newNativeFuncObj(val, r.builtin_date, r.builtin_newDate, "Date", r.global.DatePrototype, 7)
	// // v.self._putProp()
	// for _, prop := range funcProps {
	// 	// obj.propNames = append(obj.propNames, unistring.String(prop.Name))
	// 	v.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	// }
	// o._putProp("parse", r.newNativeFunc(r.date_parse, nil, "parse", nil, 1), true, false, true)
	// o._putProp("UTC", r.newNativeFunc(r.date_UTC, nil, "UTC", nil, 7), true, false, true)
	// o._putProp("now", r.newNativeFunc(r.date_now, nil, "now", nil, 0), true, false, true)
	// toString := asciiString(fmt.Sprintf("[ object %s ]", className))
	ctorImpl := func(call ConstructorCall) *Object {
		fmt.Println("creating new inside ctor", className)
		fCall := FunctionCall{
			ctx:       call.ctx,
			This:      call.This,
			Arguments: call.Arguments,
		}
		val := ctor(fCall)
		fmt.Println("what's the val here", val)
		call.This.__wrapped = val
		// add the toString function first so it can be overridden if user wants to do so
		call.This.self._putProp("name", asciiString(className), true, false, true)
		for _, prop := range funcProps {
			// obj.propNames = append(obj.propNames, unistring.String(prop.Name))
			// responseProto.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
			call.This.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
		}
		for _, prop := range classProps {
			// obj.propNames = append(obj.propNames, unistring.String(prop.Name))
			// responseProto.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
			call.This.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
		}
		return nil
	}
	responseObject := r.ToValue(ctorImpl).(*Object)
	// responseObject.Set("prototype", classProto)
	p := responseObject.Get("prototype")
	pObject := p.(*Object)
	proto := pObject.self

	proto._putProp("name", asciiString(className), true, false, true)
	for _, prop := range classProps {
		proto.setOwnStr(unistring.String(prop.Name), prop.Value, false)
	}
	responseObject.self._putProp("name", asciiString(className), true, false, true)
	for _, prop := range funcProps {
		// obj.propNames = append(obj.propNames, unistring.String(prop.Name))
		// responseProto.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
		responseObject.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
	}

	// r.Set("Response", responseObject)
	for _, prop := range funcProps {
		// obj.propNames = append(obj.propNames, unistring.String(prop.Name))
		proto._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
		// proto.Set(unistring.String(prop.Name), prop.Value)
	}

	v := NativeClass{classProto: pObject, className: className, classProps: classProps, funcProps: funcProps, ctor: ctor, Function: responseObject}
	v.runtime = r

	return v
}

type _goNativeValue struct {
	*baseObject
	value interface{}
}

func (n NativeClass) InstanceOf(val interface{}) Value {
	fmt.Println("creating new ", n.className)
	r := n.runtime
	className := n.className
	classProto := n.classProto
	obj, err := r.New(r.newNativeFuncConstruct(func(args []Value, proto *Object) *Object {
		obj := r.newBaseObject(proto, className)
		// call := FunctionCall{
		// 	ctx:       r.ctx,
		// 	This:      obj.val,
		// 	Arguments: args,
		// }
		obj.class = n.className
		g := &_goNativeValue{baseObject: obj, value: val}
		obj.val.self = g
		obj.val.__wrapped = g.value
		return obj.val
	}, unistring.String(className), classProto, 1))
	if err != nil {
		panic(err)
	}
	if err, ok := val.(error); ok {
		// stackTrace := getStacktrace(blessedValueErr)
		// if len(stackTrace) == 0 {
		// 	stackTrace = newError(self.runtime, className, 0, blessedValueErr, blessedValueErr.Error()).formatWithStack()
		// }

		obj.self._putProp("message", newStringValue(err.Error()), true, false, true)
		// obj.defineProperty("message", toValue_string(blessedValueErr.Error()), 0111, false)
		// initStacktrace(blessedValueErr, stackTrace)
	}
	obj.self._putProp("name", asciiString(n.className), true, false, true)
	for _, prop := range n.funcProps {
		// obj.propNames = append(obj.propNames, unistring.String(prop.Name))
		obj.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
		// proto.Set(unistring.String(prop.Name), prop.Value)
	}
	return obj
}

func (n NativeClass) Constructor(call ConstructorCall) *Object {

	// r := n.runtime
	// obj := r.newBaseObject(n.classProto, n.className)
	// call := FunctionCall{
	// 	ctx:       r.vm.ctx,
	// 	This:      obj.val,
	// 	Arguments: args,
	// }

	// val := ctor(call)
	// ex := &Exception{
	// 	val: r.newError(r.global.ReferenceError, err.Error()),
	// }
	// stackTrace := bytes.NewBuffer(nil)
	// ex.writeShortStack(stackTrace)
	// initStacktrace(err, stackTrace.String())
	// errMsg = newStringValue(err.Error())
	// obj._putProp("message", errMsg, true, false, true)

	//TODO set properties via below
	// responseObject := r.ToValue(responseConstructor).(*Object)
	// responseProto := responseObject.Get("prototype").(*Object)
	// r.Set("Response", responseObject)
	fCall := FunctionCall{
		ctx:       call.ctx,
		This:      call.This,
		Arguments: call.Arguments,
	}
	val := n.ctor(fCall)
	call.This.__wrapped = val

	for _, prop := range n.classProps {
		// obj.propNames = append(obj.propNames, unistring.String(prop.Name))
		call.This.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
		// proto.Set(unistring.String(prop.Name), prop.Value)
	}

	for _, prop := range n.funcProps {
		// obj.propNames = append(obj.propNames, unistring.String(prop.Name))
		call.This.self._putProp(unistring.String(prop.Name), prop.Value, true, false, true)
		// proto.Set(unistring.String(prop.Name), prop.Value)
	}

	return nil
}

// CreateNativeFunction creates a native function that will call the given call function.
// This provides for a way to detail how the function appears to a user within JS
// compared to passing the call in via toValue.
func (r *Runtime) CreateNativeFunction(name, file string, call func(FunctionCall) Value) (Value, error) {
	if call == nil {
		return UndefinedValue(), errors.New("call cannot be nil")
	}

	// r.toObject()
	// r.newNativeFunc()
	return r.newNativeFunc(call, nil, unistring.String(name), nil, 1), nil
	// return toValue_object(r.runtime.newNativeFunction(name, file, line, call)), nil
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
	// ex := r.vm.try(r.ctx)
	// if ex != nil {
	// 	return nil, ex
	// }
	vm.run()
	vm.popCtx()
	vm.halt = false
	retval := vm.stack[vm.sp-1]
	vm.sp -= 2
	return retval, nil
}
