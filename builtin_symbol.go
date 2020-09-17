package goja

func (r *Runtime) symbolproto_toString(call FunctionCall) Value {
	var b bool
	switch o := call.This.(type) {
	case valueBool:
		b = bool(o)
		goto success
	case *Object:
		if p, ok := o.self.(*primitiveValueObject); ok {
			if b1, ok := p.pValue.(valueBool); ok {
				b = bool(b1)
				goto success
			}
		}
	}
	r.typeErrorResult(true, "Method Symbol.prototype.toString is called on incompatible receiver")

success:
	if b {
		return stringTrue
	}
	return stringFalse
}

func (r *Runtime) symbolproto_valueOf(call FunctionCall) Value {
	switch o := call.This.(type) {
	case valueBool:
		return o
	case *Object:
		if p, ok := o.self.(*primitiveValueObject); ok {
			if b, ok := p.pValue.(valueBool); ok {
				return b
			}
		}
	}

	r.typeErrorResult(true, "Method Symbol.prototype.valueOf is called on incompatible receiver")
	return nil
}

func (r *Runtime) builtin_newSymbol(args []Value) *Object {
	var v Value
	if len(args) > 0 {
		if args[0].ToBoolean() {
			v = valueTrue
		} else {
			v = valueFalse
		}
	} else {
		v = valueFalse
	}
	return r.newPrimitiveObject(v, r.global.SymbolPrototype, classBoolean)
}

func (r *Runtime) builtin_Symbol(call FunctionCall) Value {
	if len(call.Arguments) > 0 {
		if call.Arguments[0].ToBoolean() {
			return valueTrue
		} else {
			return valueFalse
		}
	} else {
		return valueFalse
	}
}

func (r *Runtime) initSymbol() {
	r.global.SymbolPrototype = r.newPrimitiveObject(valueFalse, r.global.ObjectPrototype, classSymbol)
	o := r.global.SymbolPrototype.self
	o._putProp("toString", r.newNativeFunc(r.symbolproto_toString, nil, "toString", nil, 0), true, false, true)
	o._putProp("valueOf", r.newNativeFunc(r.symbolproto_valueOf, nil, "valueOf", nil, 0), true, false, true)

	r.global.Symbol = r.newNativeFunc(r.builtin_Symbol, r.builtin_newSymbol, "Symbol", r.global.SymbolPrototype, 1)
	r.addToGlobal("Symbol", r.global.Symbol)
}
