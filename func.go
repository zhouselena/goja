package goja

import (
	"context"
	"reflect"

	"github.com/dop251/goja/unistring"
)

type baseFuncObject struct {
	ctx context.Context

	baseObject

	lenProp valueProperty
}

type baseJsFuncObject struct {
	baseFuncObject

	stash   *stash
	privEnv *privateEnv

	prg    *Program
	src    string
	strict bool
}

type funcObjectImpl interface {
	source() valueString
}

type funcObject struct {
	baseJsFuncObject
}

type classFuncObject struct {
	baseJsFuncObject
	initFields   *Program
	computedKeys []Value

	privateEnvType *privateEnvType
	privateMethods []Value

	derived bool
}

type methodFuncObject struct {
	baseJsFuncObject
	homeObject *Object
}

type arrowFuncObject struct {
	baseJsFuncObject
	funcObj   *Object
	newTarget Value
}

type nativeFuncObject struct {
	ctx context.Context
	baseFuncObject

	f         func(FunctionCall) Value
	construct func(args []Value, newTarget *Object) *Object
}

type wrappedFuncObject struct {
	nativeFuncObject
	wrapped reflect.Value
}

type boundFuncObject struct {
	nativeFuncObject
	wrapped *Object
}

func (f *nativeFuncObject) export(*objectExportCtx) interface{} {
	return f.f
}

func (f *wrappedFuncObject) exportType() reflect.Type {
	return f.wrapped.Type()
}

func (f *wrappedFuncObject) export(*objectExportCtx) interface{} {
	return f.wrapped.Interface()
}

func (f *funcObject) _addProto(n unistring.String) Value {
	if n == "prototype" {
		if _, exists := f.values[n]; !exists {
			return f.addPrototype()
		}
	}
	return nil
}

func (f *funcObject) getStr(p unistring.String, receiver Value) Value {
	return f.getStrWithOwnProp(f.getOwnPropStr(p), p, receiver)
}

func (f *funcObject) getOwnPropStr(name unistring.String) Value {
	if v := f._addProto(name); v != nil {
		return v
	}

	return f.baseObject.getOwnPropStr(name)
}

func (f *funcObject) setOwnStr(name unistring.String, val Value, throw bool) bool {
	f._addProto(name)
	return f.baseObject.setOwnStr(name, val, throw)
}

func (f *funcObject) setForeignStr(name unistring.String, val, receiver Value, throw bool) (bool, bool) {
	return f._setForeignStr(name, f.getOwnPropStr(name), val, receiver, throw)
}

func (f *funcObject) defineOwnPropertyStr(name unistring.String, descr PropertyDescriptor, throw bool) bool {
	f._addProto(name)
	return f.baseObject.defineOwnPropertyStr(name, descr, throw)
}

func (f *funcObject) deleteStr(name unistring.String, throw bool) bool {
	f._addProto(name)
	return f.baseObject.deleteStr(name, throw)
}

func (f *funcObject) addPrototype() Value {
	proto := f.val.runtime.NewObject()
	proto.self._putProp("constructor", f.val, true, false, true)
	return f._putProp("prototype", proto, true, false, false)
}

func (f *funcObject) hasOwnPropertyStr(name unistring.String) bool {
	if f.baseObject.hasOwnPropertyStr(name) {
		return true
	}

	if name == "prototype" {
		return true
	}
	return false
}

func (f *funcObject) stringKeys(all bool, accum []Value) []Value {
	if all {
		if _, exists := f.values["prototype"]; !exists {
			accum = append(accum, asciiString("prototype"))
		}
	}
	return f.baseFuncObject.stringKeys(all, accum)
}

func (f *funcObject) iterateStringKeys() iterNextFunc {
	if _, exists := f.values["prototype"]; !exists {
		f.addPrototype()
	}
	return f.baseFuncObject.iterateStringKeys()
}

func (f *baseFuncObject) createInstance(newTarget *Object) *Object {
	r := f.val.runtime
	if newTarget == nil {
		newTarget = f.val
	}
	proto := r.getPrototypeFromCtor(newTarget, nil, r.global.ObjectPrototype)

	return f.val.runtime.newBaseObject(proto, classObject).val
}

func (f *baseJsFuncObject) construct(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		newTarget = f.val
	}
	proto := newTarget.self.getStr("prototype", nil)
	var protoObj *Object
	if p, ok := proto.(*Object); ok {
		protoObj = p
	} else {
		protoObj = f.val.runtime.global.ObjectPrototype
	}

	obj := f.val.runtime.newBaseObject(protoObj, classObject).val
	ret := f.call(FunctionCall{
		ctx:       f.val.runtime.vm.ctx,
		This:      obj,
		Arguments: args,
	}, newTarget)

	if ret, ok := ret.(*Object); ok {
		return ret
	}
	return obj
}

func (f *classFuncObject) Call(FunctionCall) Value {
	panic(f.val.runtime.NewTypeError("Class constructor cannot be invoked without 'new'"))
}

func (f *classFuncObject) assertCallable() (func(FunctionCall) Value, bool) {
	return f.Call, true
}

func (f *classFuncObject) export(*objectExportCtx) interface{} {
	return f.Call
}

func (f *classFuncObject) createInstance(args []Value, newTarget *Object) (instance *Object) {
	if f.derived {
		if ctor := f.prototype.self.assertConstructor(); ctor != nil {
			instance = ctor(args, newTarget)
		} else {
			panic(f.val.runtime.NewTypeError("Super constructor is not a constructor"))
		}
	} else {
		instance = f.baseFuncObject.createInstance(newTarget)
	}
	return
}

func (f *classFuncObject) _initFields(instance *Object) {
	if f.privateEnvType != nil {
		penv := instance.self.getPrivateEnv(f.privateEnvType, true)
		penv.methods = f.privateMethods
	}
	if f.initFields != nil {
		vm := f.val.runtime.vm
		vm.pushCtx()
		vm.prg = f.initFields
		vm.stash = f.stash
		vm.privEnv = f.privEnv
		vm.newTarget = nil

		// so that 'super' base could be correctly resolved (including from direct eval())
		vm.push(f.val)

		vm.sb = vm.sp
		vm.push(instance)
		vm.pc = 0
		vm.run()
		vm.popCtx()
		vm.sp -= 2
		vm.halt = false
	}
}

func (f *classFuncObject) construct(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		newTarget = f.val
	}
	if f.prg == nil {
		instance := f.createInstance(args, newTarget)
		f._initFields(instance)
		return instance
	} else {
		var instance *Object
		var thisVal Value
		if !f.derived {
			instance = f.createInstance(args, newTarget)
			f._initFields(instance)
			thisVal = instance
		}
		ret := f._call(f.val.runtime.vm.ctx, args, newTarget, thisVal)

		if ret, ok := ret.(*Object); ok {
			return ret
		}
		if f.derived {
			r := f.val.runtime
			if ret != _undefined {
				panic(r.NewTypeError("Derived constructors may only return object or undefined"))
			}
			if v := r.vm.stack[r.vm.sp+1]; v != nil { // using residual 'this' value (a bit hacky)
				instance = r.toObject(v)
			} else {
				panic(r.newError(r.getReferenceError(), "Must call super constructor in derived class before returning from derived constructor"))
			}
		}
		return instance
	}
}

func (f *classFuncObject) assertConstructor() func(args []Value, newTarget *Object) *Object {
	return f.construct
}

func (f *baseJsFuncObject) Call(call FunctionCall) Value {
	return f.call(call, nil)
}

func (f *arrowFuncObject) Call(call FunctionCall) Value {
	return f._call(call.Context(), call.Arguments, f.newTarget, nil)
}

func (f *baseJsFuncObject) _call(ctx context.Context, args []Value, newTarget, this Value) Value {
	vm := f.val.runtime.vm

	vm.stack.expand(vm.sp + len(args) + 1)
	vm.stack[vm.sp] = f.val
	vm.sp++
	vm.stack[vm.sp] = this
	vm.sp++
	for _, arg := range args {
		if arg != nil {
			vm.stack[vm.sp] = arg
		} else {
			vm.stack[vm.sp] = _undefined
		}
		vm.sp++
	}

	pc := vm.pc
	if pc != -1 {
		vm.pc++ // fake "return address" so that captureStack() records the correct call location
		vm.pushCtx()
		vm.callStack = append(vm.callStack, vmContext{ctx: ctx, pc: -1}) // extra frame so that run() halts after ret
	} else {
		vm.pushCtx()
	}
	vm.args = len(args)
	vm.prg = f.prg
	vm.stash = f.stash
	vm.privEnv = f.privEnv
	vm.newTarget = newTarget
	vm.pc = 0
	vm.ctx = ctx
	vm.run()
	if pc != -1 {
		vm.popCtx()
	}
	vm.pc = pc
	vm.halt = false
	return vm.pop()
}

func (f *baseJsFuncObject) call(call FunctionCall, newTarget Value) Value {
	return f._call(call.Context(), call.Arguments, newTarget, nilSafe(call.This))
}

func (f *baseJsFuncObject) export(*objectExportCtx) interface{} {
	return f.Call
}

func (f *baseFuncObject) exportType() reflect.Type {
	return reflectTypeFunc
}

func (f *baseJsFuncObject) assertCallable() (func(FunctionCall) Value, bool) {
	return f.Call, true
}

func (f *funcObject) assertConstructor() func(args []Value, newTarget *Object) *Object {
	return f.construct
}

func (f *arrowFuncObject) assertCallable() (func(FunctionCall) Value, bool) {
	return f.Call, true
}

func (f *arrowFuncObject) export(*objectExportCtx) interface{} {
	return f.Call
}

func (f *baseFuncObject) init(name unistring.String, length Value) {
	f.baseObject.init()

	f.lenProp.configurable = true
	f.lenProp.value = length
	f._put("length", &f.lenProp)

	f._putProp("name", stringValueFromRaw(name), false, false, true)
}

func hasInstance(val *Object, v Value) bool {
	if v, ok := v.(*Object); ok {
		o := val.self.getStr("prototype", nil)
		if o1, ok := o.(*Object); ok {
			for {
				v = v.self.proto()
				if v == nil {
					return false
				}
				if o1 == v {
					return true
				}
			}
		} else {
			panic(val.runtime.NewTypeError("prototype is not an object"))
		}
	}

	return false
}

func (f *baseFuncObject) hasInstance(v Value) bool {
	return hasInstance(f.val, v)
}

func (f *nativeFuncObject) defaultConstruct(ccall func(ConstructorCall) *Object, args []Value, newTarget *Object) *Object {
	obj := f.createInstance(newTarget)
	ret := ccall(ConstructorCall{
		ctx:       f.ctx,
		This:      obj,
		Arguments: args,
		NewTarget: newTarget,
	})

	if ret != nil {
		return ret
	}
	return obj
}

func (f *nativeFuncObject) assertCallable() (func(FunctionCall) Value, bool) {
	if f.f != nil {
		return f.Call, true
	}
	return nil, false
}

func (f *nativeFuncObject) Call(call FunctionCall) Value {
	vm := f.val.runtime.vm
	prevFuncName := vm.getFuncName()
	// This is done to display the correct function name in the stack trace when executing a
	// native function with Function.prototype.apply/call
	vm.setFuncName(f.getStr("name", nil).string())
	rv := f.f(call)
	vm.setFuncName(prevFuncName)
	return rv
}

func (f *nativeFuncObject) vmCall(vm *vm, n int) {
	if f.f != nil {
		vm.pushCtx()
		vm.prg = nil
		vm.sb = vm.sp - n // so that [sb-1] points to the callee
		ret := f.f(FunctionCall{
			Arguments: vm.stack[vm.sp-n : vm.sp],
			This:      vm.stack[vm.sp-n-2],
		})
		if ret == nil {
			ret = _undefined
		}
		vm.stack[vm.sp-n-2] = ret
		vm.popCtx()
	} else {
		vm.stack[vm.sp-n-2] = _undefined
	}
	vm.sp -= n + 1
	vm.pc++
}

func (f *nativeFuncObject) assertConstructor() func(args []Value, newTarget *Object) *Object {
	return f.construct
}

/*func (f *boundFuncObject) getStr(p unistring.String, receiver Value) Value {
	return f.getStrWithOwnProp(f.getOwnPropStr(p), p, receiver)
}

func (f *boundFuncObject) getOwnPropStr(name unistring.String) Value {
	if name == "caller" || name == "arguments" {
		return f.val.runtime.global.throwerProperty
	}

	return f.nativeFuncObject.getOwnPropStr(name)
}

func (f *boundFuncObject) deleteStr(name unistring.String, throw bool) bool {
	if name == "caller" || name == "arguments" {
		return true
	}
	return f.nativeFuncObject.deleteStr(name, throw)
}

func (f *boundFuncObject) setOwnStr(name unistring.String, val Value, throw bool) bool {
	if name == "caller" || name == "arguments" {
		panic(f.val.runtime.NewTypeError("'caller' and 'arguments' are restricted function properties and cannot be accessed in this context."))
	}
	return f.nativeFuncObject.setOwnStr(name, val, throw)
}

func (f *boundFuncObject) setForeignStr(name unistring.String, val, receiver Value, throw bool) (bool, bool) {
	return f._setForeignStr(name, f.getOwnPropStr(name), val, receiver, throw)
}
*/

func (f *boundFuncObject) hasInstance(v Value) bool {
	return instanceOfOperator(v, f.wrapped)
}

func (f *nativeFuncObject) MemUsage(ctx *MemUsageContext) (memUsage uint64, err error) {
	if f == nil || ctx.IsObjVisited(f) {
		return SizeEmptyStruct, nil
	}
	ctx.VisitObj(f)

	return f.baseFuncObject.MemUsage(ctx)
}

func (f *funcObject) MemUsage(ctx *MemUsageContext) (memUsage uint64, err error) {
	if f == nil || ctx.IsObjVisited(f) {
		return SizeEmptyStruct, nil
	}
	ctx.VisitObj(f)

	memUsage, err = f.baseObject.MemUsage(ctx)
	if err != nil {
		return memUsage, err
	}

	if f.stash != nil {
		inc, err := f.stash.MemUsage(ctx)
		memUsage += inc
		if err != nil {
			return memUsage, err
		}
	}

	return memUsage, nil
}
