package goja

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"unsafe"
)

var (
	valueFalse    Value = valueBool(false)
	valueTrue     Value = valueBool(true)
	_null         Value = valueNull{}
	_NaN          Value = valueFloat(math.NaN())
	_positiveInf  Value = valueFloat(math.Inf(+1))
	_negativeInf  Value = valueFloat(math.Inf(-1))
	_positiveZero Value
	_negativeZero Value = valueFloat(math.Float64frombits(0 | (1 << 63)))
	_epsilon            = valueFloat(2.2204460492503130808472633361816e-16)
	_undefined    Value = valueUndefined{}
)

var (
	reflectTypeInt64  = reflect.TypeOf(int64(0))
	reflectTypeInt32  = reflect.TypeOf(int32(0))
	reflectTypeUInt32 = reflect.TypeOf(uint32(0))
	reflectTypeInt    = reflect.TypeOf(0)

	reflectTypeBool   = reflect.TypeOf(false)
	reflectTypeNil    = reflect.TypeOf(nil)
	reflectTypeFloat  = reflect.TypeOf(float64(0))
	reflectTypeMap    = reflect.TypeOf(map[string]interface{}{})
	reflectTypeArray  = reflect.TypeOf([]interface{}{})
	reflectTypeString = reflect.TypeOf("")
)

var intCache [256]Value

func FalseValue() Value {
	return valueFalse
}
func TrueValue() Value {
	return valueTrue
}

type Value interface {
	ToInt() int
	ToUInt32() uint32
	ToInt32() int32
	ToInt64() int64
	ToString() valueString
	String() string
	ToFloat() float64
	ToNumber() Value
	ToBoolean() bool
	ToObject(*Runtime) *Object
	SameAs(Value) bool
	Equals(Value) bool
	StrictEquals(Value) bool
	Export() (interface{}, error)
	ExportType() reflect.Type

	IsObject() bool
	IsNumber() bool

	assertInt() (int, bool)
	assertUInt32() (uint32, bool)
	assertInt32() (int32, bool)
	assertInt64() (int64, bool)
	assertString() (valueString, bool)
	assertFloat() (float64, bool)

	baseObject(r *Runtime) *Object

	MemUsage(ctx *MemUsageContext) (uint64, error)
}

const (
	BoolSize   = uint64(unsafe.Sizeof(true))
	NumberSize = uint64(unsafe.Sizeof(float64(0)))
	IntSize    = uint64(unsafe.Sizeof(float64(0)))
	EmptySize  = uint64(unsafe.Sizeof((*baseObject)(nil)))
)

type valueNumber struct {
	_type reflect.Type
	val   interface{}
}

type valueInt int
type valueUInt32 uint32
type valueInt32 int32
type valueInt64 int64
type valueFloat float64
type valueBool bool
type valueNull struct{}
type valueUndefined struct {
	valueNull
}

func UndefinedValue() Value {
	return valueUndefined{}
}

type valueUnresolved struct {
	r   *Runtime
	ref string
}

type memberUnresolved struct {
	valueUnresolved
}

type valueProperty struct {
	value        Value
	writable     bool
	configurable bool
	enumerable   bool
	accessor     bool
	getterFunc   *Object
	setterFunc   *Object
}

func propGetter(o Value, v Value, r *Runtime) *Object {
	if v == _undefined {
		return nil
	}
	if obj, ok := v.(*Object); ok {
		if _, ok := obj.self.assertCallable(); ok {
			return obj
		}
	}
	r.typeErrorResult(true, "Getter must be a function: %s", v.ToString())
	return nil
}

func propSetter(o Value, v Value, r *Runtime) *Object {
	if v == _undefined {
		return nil
	}
	if obj, ok := v.(*Object); ok {
		if _, ok := obj.self.assertCallable(); ok {
			return obj
		}
	}
	r.typeErrorResult(true, "Setter must be a function: %s", v.ToString())
	return nil
}

func (i valueNumber) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return NumberSize, nil
}

func (i valueNumber) ToInteger() (int, bool) {
	switch v := i.val.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	}

	return 0, false
}

func (i valueNumber) ToInt() int {
	switch v := i.val.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	}

	return 0
}

func (i valueNumber) ToInt32() int32 {
	switch v := i.val.(type) {
	case int:
		return int32(v)
	case int8:
		return int32(v)
	case int32:
		return int32(v)
	case int64:
		return int32(v)
	case uint32:
		return int32(v)
	case uint64:
		return int32(v)
	}

	return 0
}
func (i valueNumber) ToUInt32() uint32 {
	switch v := i.val.(type) {
	case int:
		return uint32(v)
	case int8:
		return uint32(v)
	case int32:
		return uint32(v)
	case int64:
		return uint32(v)
	case uint32:
		return v
	case uint64:
		return uint32(v)
	}

	return 0
}

func (i valueNumber) ToInt64() int64 {
	switch v := i.val.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	}

	return 0
}

func (i valueNumber) ToString() valueString {
	return asciiString(i.String())
}
func (i valueNumber) IsObject() bool {
	return false
}

func (i valueNumber) IsNumber() bool {
	return true
}

func (i valueNumber) String() string {
	return strconv.FormatInt(int64(i.ToInt()), 10)
}

func (i valueNumber) ToFloat() float64 {
	return float64(i.ToInt64())
}

func (i valueNumber) ToBoolean() bool {
	return i.val != 0
}

func (i valueNumber) ToObject(r *Runtime) *Object {
	return r.newPrimitiveObject(i, r.global.NumberPrototype, classNumber)
}

func (i valueNumber) ToNumber() Value {
	return i
}

func (i valueNumber) SameAs(other Value) bool {
	if otherInt, ok := other.assertInt(); ok {
		return i.ToInt() == otherInt
	}
	if otherFloat, ok := other.assertFloat(); ok {
		return i.ToFloat() == otherFloat
	}
	return false
}

func (i valueNumber) Equals(other Value) bool {
	if o, ok := other.assertInt(); ok {
		return i.ToInt() == o
	}
	if o, ok := other.assertFloat(); ok {
		return float64(i.ToInt()) == o
	}
	if o, ok := other.assertString(); ok {
		return o.ToNumber().Equals(i)
	}
	if o, ok := other.(valueBool); ok {
		return int(i.ToInt()) == o.ToInt()
	}
	if o, ok := other.(*Object); ok {
		return i.Equals(o.self.toPrimitiveNumber())
	}
	return false
}

func (i valueNumber) StrictEquals(other Value) bool {
	if otherInt, ok := other.assertInt(); ok {
		return int(i.ToInt()) == otherInt
	} else if otherFloat, ok := other.assertFloat(); ok {
		return float64(i.ToInt()) == otherFloat
	}
	return false
}

func (i valueNumber) assertInt() (int, bool) {
	return i.ToInt(), true
}
func (i valueNumber) assertInt32() (int32, bool) {
	return i.ToInt32(), true
}
func (i valueNumber) assertUInt32() (uint32, bool) {
	return i.ToUInt32(), true
}
func (i valueNumber) assertInt64() (int64, bool) {
	return i.ToInt64(), true
}

func (i valueNumber) assertFloat() (float64, bool) {
	return i.ToFloat(), false
}

func (i valueNumber) assertString() (valueString, bool) {
	return nil, false
}

func (i valueNumber) baseObject(r *Runtime) *Object {
	return r.global.NumberPrototype
}

func (i valueNumber) Export() (interface{}, error) {
	return i.val, nil
}

func (i valueNumber) ExportType() reflect.Type {
	return i._type
}

func (i valueUInt32) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return NumberSize, nil
}

func (i valueUInt32) ToInt() int {
	return int(i)
}

func (i valueUInt32) ToInt32() int32 {
	return int32(i)
}
func (i valueUInt32) ToUInt32() uint32 {
	return uint32(i)
}

func (i valueUInt32) ToInt64() int64 {
	return int64(i)
}

func (i valueUInt32) ToString() valueString {
	return asciiString(i.String())
}
func (i valueUInt32) IsObject() bool {
	return false
}
func (i valueUInt32) IsNumber() bool {
	return true
}

func (i valueUInt32) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i valueUInt32) ToFloat() float64 {
	return float64(int64(i))
}

func (i valueUInt32) ToBoolean() bool {
	return i != 0
}

func (i valueUInt32) ToObject(r *Runtime) *Object {
	return r.newPrimitiveObject(i, r.global.NumberPrototype, classNumber)
}

func (i valueUInt32) ToNumber() Value {
	return i
}

func (i valueUInt32) SameAs(other Value) bool {
	if otherInt, ok := other.assertUInt32(); ok {
		return uint32(i) == otherInt
	}
	return false
}

func (i valueUInt32) Equals(other Value) bool {
	if o, ok := other.assertInt32(); ok {
		return int32(i) == o
	}
	if o, ok := other.assertFloat(); ok {
		return float64(i) == o
	}
	if o, ok := other.assertString(); ok {
		return o.ToNumber().Equals(i)
	}
	if o, ok := other.(valueBool); ok {
		return int(i) == o.ToInt()
	}
	if o, ok := other.(*Object); ok {
		return i.Equals(o.self.toPrimitiveNumber())
	}
	return false
}

func (i valueUInt32) StrictEquals(other Value) bool {
	if otherInt, ok := other.assertInt32(); ok {
		return int32(i) == otherInt
	} else if otherFloat, ok := other.assertFloat(); ok {
		return float64(i) == otherFloat
	}
	return false
}

func (i valueUInt32) assertInt() (int, bool) {
	return 0, false
}
func (i valueUInt32) assertUInt32() (uint32, bool) {
	return uint32(i), true
}
func (i valueUInt32) assertInt32() (int32, bool) {
	return 0, false
}
func (i valueUInt32) assertInt64() (int64, bool) {
	return 0, false
}

func (i valueUInt32) assertFloat() (float64, bool) {
	return 0, false
}

func (i valueUInt32) assertString() (valueString, bool) {
	return nil, false
}

func (i valueUInt32) baseObject(r *Runtime) *Object {
	return r.global.NumberPrototype
}

func (i valueUInt32) Export() (interface{}, error) {
	return uint32(i), nil
}

func (i valueUInt32) ExportType() reflect.Type {
	return reflectTypeUInt32
}

func (i valueInt32) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return NumberSize, nil
}

func (i valueInt32) ToInt() int {
	return int(i)
}

func (i valueInt32) ToInt32() int32 {
	return int32(i)
}
func (i valueInt32) ToUInt32() uint32 {
	return uint32(i)
}

func (i valueInt32) ToInt64() int64 {
	return int64(i)
}

func (i valueInt32) ToString() valueString {
	return asciiString(i.String())
}
func (i valueInt32) IsObject() bool {
	return false
}
func (i valueInt32) IsNumber() bool {
	return true
}

func (i valueInt32) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i valueInt32) ToFloat() float64 {
	return float64(int64(i))
}

func (i valueInt32) ToBoolean() bool {
	return i != 0
}

func (i valueInt32) ToObject(r *Runtime) *Object {
	return r.newPrimitiveObject(i, r.global.NumberPrototype, classNumber)
}

func (i valueInt32) ToNumber() Value {
	return i
}

func (i valueInt32) SameAs(other Value) bool {
	if otherInt, ok := other.assertInt32(); ok {
		return int32(i) == otherInt
	}
	return false
}

func (i valueInt32) Equals(other Value) bool {
	if o, ok := other.assertInt32(); ok {
		return int32(i) == o
	}
	if o, ok := other.assertFloat(); ok {
		return float64(i) == o
	}
	if o, ok := other.assertString(); ok {
		return o.ToNumber().Equals(i)
	}
	if o, ok := other.(valueBool); ok {
		return int(i) == o.ToInt()
	}
	if o, ok := other.(*Object); ok {
		return i.Equals(o.self.toPrimitiveNumber())
	}
	return false
}

func (i valueInt32) StrictEquals(other Value) bool {
	if otherInt, ok := other.assertInt32(); ok {
		return int32(i) == otherInt
	} else if otherFloat, ok := other.assertFloat(); ok {
		return float64(i) == otherFloat
	}
	return false
}

func (i valueInt32) assertInt() (int, bool) {
	return 0, false
}
func (i valueInt32) assertUInt32() (uint32, bool) {
	return 0, false
}
func (i valueInt32) assertInt32() (int32, bool) {
	return int32(i), true
}
func (i valueInt32) assertInt64() (int64, bool) {
	return 0, false
}

func (i valueInt32) assertFloat() (float64, bool) {
	return 0, false
}

func (i valueInt32) assertString() (valueString, bool) {
	return nil, false
}

func (i valueInt32) baseObject(r *Runtime) *Object {
	return r.global.NumberPrototype
}

func (i valueInt32) Export() (interface{}, error) {
	return int32(i), nil
}

func (i valueInt32) ExportType() reflect.Type {
	return reflectTypeInt32
}

func (i valueInt64) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return NumberSize, nil
}

func (i valueInt64) ToInt() int {
	return int(i)
}

func (i valueInt64) ToInt32() int32 {
	return int32(i)
}
func (i valueInt64) ToUInt32() uint32 {
	return uint32(i)
}

func (i valueInt64) ToInt64() int64 {
	return int64(i)
}

func (i valueInt64) ToString() valueString {
	return asciiString(i.String())
}
func (i valueInt64) IsObject() bool {
	return false
}

func (i valueInt64) IsNumber() bool {
	return true
}
func (i valueInt64) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i valueInt64) ToFloat() float64 {
	return float64(int64(i))
}

func (i valueInt64) ToBoolean() bool {
	return i != 0
}

func (i valueInt64) ToObject(r *Runtime) *Object {
	return r.newPrimitiveObject(i, r.global.NumberPrototype, classNumber)
}

func (i valueInt64) ToNumber() Value {
	return i
}

func (i valueInt64) SameAs(other Value) bool {
	if otherInt, ok := other.assertInt64(); ok {
		return int64(i) == otherInt
	}
	return false
}

func (i valueInt64) Equals(other Value) bool {
	if o, ok := other.assertInt64(); ok {
		return int64(i) == o
	}
	if o, ok := other.assertFloat(); ok {
		return float64(i) == o
	}
	if o, ok := other.assertString(); ok {
		return o.ToNumber().Equals(i)
	}
	if o, ok := other.(valueBool); ok {
		return int(i) == o.ToInt()
	}
	if o, ok := other.(*Object); ok {
		return i.Equals(o.self.toPrimitiveNumber())
	}
	return false
}

func (i valueInt64) StrictEquals(other Value) bool {
	if otherInt, ok := other.assertInt64(); ok {
		return int64(i) == otherInt
	} else if otherFloat, ok := other.assertFloat(); ok {
		return float64(i) == otherFloat
	}
	return false
}

func (i valueInt64) assertInt() (int, bool) {
	return 0, true
}
func (i valueInt64) assertUInt32() (uint32, bool) {
	return 0, false
}
func (i valueInt64) assertInt32() (int32, bool) {
	return 0, false
}
func (i valueInt64) assertInt64() (int64, bool) {
	return int64(i), true
}

func (i valueInt64) assertFloat() (float64, bool) {
	return 0, false
}

func (i valueInt64) assertString() (valueString, bool) {
	return nil, false
}

func (i valueInt64) baseObject(r *Runtime) *Object {
	return r.global.NumberPrototype
}

func (i valueInt64) Export() (interface{}, error) {
	return int64(i), nil
}

func (i valueInt64) ExportType() reflect.Type {
	return reflectTypeInt64
}

func (i valueInt) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return NumberSize, nil
}

func (i valueInt) ToInt() int {
	return int(i)
}

func (i valueInt) ToInt32() int32 {
	return int32(i)
}
func (i valueInt) ToUInt32() uint32 {
	return uint32(i)
}

func (i valueInt) ToInt64() int64 {
	return int64(i)
}

func (i valueInt) ToString() valueString {
	return asciiString(i.String())
}
func (i valueInt) IsObject() bool {
	return false
}
func (i valueInt) IsNumber() bool {
	return true
}

func (i valueInt) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i valueInt) ToFloat() float64 {
	return float64(int64(i))
}

func (i valueInt) ToBoolean() bool {
	return i != 0
}

func (i valueInt) ToObject(r *Runtime) *Object {
	return r.newPrimitiveObject(i, r.global.NumberPrototype, classNumber)
}

func (i valueInt) ToNumber() Value {
	return i
}

func (i valueInt) SameAs(other Value) bool {
	if otherInt, ok := other.assertInt(); ok {
		return int(i) == otherInt
	}
	return false
}

func (i valueInt) Equals(other Value) bool {
	if o, ok := other.assertInt(); ok {
		return int(i) == o
	}
	if o, ok := other.assertFloat(); ok {
		return float64(i) == o
	}
	if o, ok := other.assertString(); ok {
		return o.ToNumber().Equals(i)
	}
	if o, ok := other.(valueBool); ok {
		return int(i) == o.ToInt()
	}
	if o, ok := other.(*Object); ok {
		return i.Equals(o.self.toPrimitiveNumber())
	}
	return false
}

func (i valueInt) StrictEquals(other Value) bool {
	if otherInt, ok := other.assertInt(); ok {
		return int(i) == otherInt
	} else if otherFloat, ok := other.assertFloat(); ok {
		return float64(i) == otherFloat
	}
	return false
}

func (i valueInt) assertInt() (int, bool) {
	return int(i), true
}
func (i valueInt) assertUInt32() (uint32, bool) {
	return 0, false
}
func (i valueInt) assertInt32() (int32, bool) {
	return 0, false
}
func (i valueInt) assertInt64() (int64, bool) {
	return 0, false
}

func (i valueInt) assertFloat() (float64, bool) {
	return 0, false
}

func (i valueInt) assertString() (valueString, bool) {
	return nil, false
}

func (i valueInt) baseObject(r *Runtime) *Object {
	return r.global.NumberPrototype
}

func (i valueInt) Export() (interface{}, error) {
	return int64(i), nil
}

func (i valueInt) ExportType() reflect.Type {
	return reflectTypeInt
}

func (o valueBool) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return BoolSize, nil
}

func (o valueBool) ToInt64() int64 {
	if o {
		return 1
	}
	return 0
}
func (o valueBool) ToInt32() int32 {
	if o {
		return 1
	}
	return 0
}
func (o valueBool) ToUInt32() uint32 {
	if o {
		return 1
	}
	return 0
}
func (o valueBool) ToInt() int {
	if o {
		return 1
	}
	return 0
}

func (o valueBool) ToString() valueString {
	if o {
		return stringTrue
	}
	return stringFalse
}

func (o valueBool) String() string {
	if o {
		return "true"
	}
	return "false"
}
func (o valueBool) IsObject() bool {
	return false
}
func (o valueBool) IsNumber() bool {
	return false
}

func (o valueBool) ToFloat() float64 {
	if o {
		return 1.0
	}
	return 0
}

func (o valueBool) ToBoolean() bool {
	return bool(o)
}

func (o valueBool) ToObject(r *Runtime) *Object {
	return r.newPrimitiveObject(o, r.global.BooleanPrototype, "Boolean")
}

func (o valueBool) ToNumber() Value {
	if o {
		return valueInt(1)
	}
	return valueInt(0)
}

func (o valueBool) SameAs(other Value) bool {
	if other, ok := other.(valueBool); ok {
		return o == other
	}
	return false
}

func (b valueBool) Equals(other Value) bool {
	if o, ok := other.(valueBool); ok {
		return b == o
	}

	if b {
		return other.Equals(intToValue(1))
	} else {
		return other.Equals(intToValue(0))
	}

}

func (o valueBool) StrictEquals(other Value) bool {
	if other, ok := other.(valueBool); ok {
		return o == other
	}
	return false
}

func (o valueBool) assertInt() (int, bool) {
	return 0, false
}

func (o valueBool) assertInt32() (int32, bool) {
	return 0, false
}
func (o valueBool) assertUInt32() (uint32, bool) {
	return 0, false
}

func (o valueBool) assertInt64() (int64, bool) {
	return 0, false
}

func (o valueBool) assertFloat() (float64, bool) {
	return 0, false
}

func (o valueBool) assertString() (valueString, bool) {
	return nil, false
}

func (o valueBool) baseObject(r *Runtime) *Object {
	return r.global.BooleanPrototype
}

func (o valueBool) Export() (interface{}, error) {
	return bool(o), nil
}

func (o valueBool) ExportType() reflect.Type {
	return reflectTypeBool
}

func (n valueNull) ToInt() int {
	return 0
}
func (n valueNull) ToInt32() int32 {
	return 0
}
func (n valueNull) ToUInt32() uint32 {
	return 0
}
func (n valueNull) ToInt64() int64 {
	return 0
}

func (n valueNull) ToString() valueString {
	return stringNull
}

func (n valueNull) String() string {
	return "null"
}

func (u valueUndefined) ToString() valueString {
	return stringUndefined
}

func (u valueUndefined) String() string {
	return "undefined"
}
func (u valueUndefined) IsObject() bool {
	return false
}
func (u valueUndefined) IsNumber() bool {
	return false
}

func (u valueUndefined) ToNumber() Value {
	return _NaN
}

func (u valueUndefined) SameAs(other Value) bool {
	_, same := other.(valueUndefined)
	return same
}

func (u valueUndefined) StrictEquals(other Value) bool {
	_, same := other.(valueUndefined)
	return same
}

func (u valueUndefined) ToFloat() float64 {
	return math.NaN()
}

func (n valueNull) ToFloat() float64 {
	return 0
}

func (n valueNull) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return EmptySize, nil
}

func (n valueNull) ToBoolean() bool {
	return false
}

func (n valueNull) ToObject(r *Runtime) *Object {
	r.typeErrorResult(true, "Cannot convert undefined or null to object")
	return nil
	//return r.newObject()
}

func (n valueNull) ToNumber() Value {
	return intToValue(0)
}

func (n valueNull) SameAs(other Value) bool {
	_, same := other.(valueNull)
	return same
}

func (n valueNull) Equals(other Value) bool {
	switch other.(type) {
	case valueUndefined, valueNull:
		return true
	}
	return false
}

func (n valueNull) StrictEquals(other Value) bool {
	_, same := other.(valueNull)
	return same
}

func (n valueNull) assertInt() (int, bool) {
	return 0, false
}
func (n valueNull) assertUInt32() (uint32, bool) {
	return 0, false
}
func (n valueNull) assertInt32() (int32, bool) {
	return 0, false
}
func (n valueNull) assertInt64() (int64, bool) {
	return 0, false
}

func (n valueNull) assertFloat() (float64, bool) {
	return 0, false
}

func (n valueNull) assertString() (valueString, bool) {
	return nil, false
}

func (n valueNull) baseObject(r *Runtime) *Object {
	return nil
}

func (n valueNull) Export() (interface{}, error) {
	return nil, nil
}
func (n valueNull) IsObject() bool {
	return false
}
func (n valueNull) IsNumber() bool {
	return false
}

func (n valueNull) ExportType() reflect.Type {
	return reflectTypeNil
}

func (p *valueProperty) MemUsage(ctx *MemUsageContext) (uint64, error) {
	total := EmptySize
	total += uint64(len(p.String())) // count size of property name towards total object size.
	if p.value != nil {
		inc, err := p.value.MemUsage(ctx)
		total += inc
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func (p *valueProperty) ToInt() int {
	return 0
}

func (p *valueProperty) ToInt32() int32 {
	return 0
}
func (p *valueProperty) ToUInt32() uint32 {
	return 0
}

func (p *valueProperty) ToInt64() int64 {
	return 0
}

func (p *valueProperty) ToString() valueString {
	return stringEmpty
}

func (p *valueProperty) String() string {
	return ""
}
func (p *valueProperty) IsObject() bool {
	return false
}
func (p *valueProperty) IsNumber() bool {
	return false
}

func (p *valueProperty) ToFloat() float64 {
	return math.NaN()
}

func (p *valueProperty) ToBoolean() bool {
	return false
}

func (p *valueProperty) ToObject(r *Runtime) *Object {
	return nil
}

func (p *valueProperty) ToNumber() Value {
	return nil
}

func (p *valueProperty) assertInt() (int, bool) {
	return 0, false
}
func (p *valueProperty) assertUInt32() (uint32, bool) {
	return 0, false
}
func (p *valueProperty) assertInt32() (int32, bool) {
	return 0, false
}
func (p *valueProperty) assertInt64() (int64, bool) {
	return 0, false
}

func (p *valueProperty) assertFloat() (float64, bool) {
	return 0, false
}

func (p *valueProperty) assertString() (valueString, bool) {
	return nil, false
}

func (p *valueProperty) isWritable() bool {
	return p.writable || p.setterFunc != nil
}

func (p *valueProperty) get(this Value) Value {
	if p.getterFunc == nil {
		if p.value != nil {
			return p.value
		}
		return _undefined
	}
	call, _ := p.getterFunc.self.assertCallable()
	return call(FunctionCall{
		This: this,
	})
}

func (p *valueProperty) set(this, v Value) {
	if p.setterFunc == nil {
		p.value = v
		return
	}
	call, _ := p.setterFunc.self.assertCallable()
	call(FunctionCall{
		This:      this,
		Arguments: []Value{v},
	})
}

func (p *valueProperty) SameAs(other Value) bool {
	if otherProp, ok := other.(*valueProperty); ok {
		return p == otherProp
	}
	return false
}

func (p *valueProperty) Equals(other Value) bool {
	return false
}

func (p *valueProperty) StrictEquals(other Value) bool {
	return false
}

func (n *valueProperty) baseObject(r *Runtime) *Object {
	r.typeErrorResult(true, "BUG: baseObject() is called on valueProperty") // TODO error message
	return nil
}

func (n *valueProperty) Export() (interface{}, error) {
	panic("Cannot export valueProperty")
}

func (n *valueProperty) ExportType() reflect.Type {
	panic("Cannot export valueProperty")
}

func (f valueFloat) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return NumberSize, nil
}

func (f valueFloat) ToInt() int {
	switch {
	case math.IsNaN(float64(f)):
		return 0
	case math.IsInf(float64(f), 1):
		return math.MaxInt64
	case math.IsInf(float64(f), -1):
		return math.MinInt64
	}
	return int(f)
}
func (f valueFloat) ToInt32() int32 {
	switch {
	case math.IsNaN(float64(f)):
		return 0
	case math.IsInf(float64(f), 1):
		return int32(math.MaxInt32)
	case math.IsInf(float64(f), -1):
		return int32(math.MinInt32)
	}
	return int32(f)
}
func (f valueFloat) ToUInt32() uint32 {
	switch {
	case math.IsNaN(float64(f)):
		return 0
	case math.IsInf(float64(f), 1):
		return uint32(math.MaxInt32)
	case math.IsInf(float64(f), -1):
		return uint32(f)
	}
	return uint32(f)
}
func (f valueFloat) ToInt64() int64 {
	switch {
	case math.IsNaN(float64(f)):
		return 0
	case math.IsInf(float64(f), 1):
		return int64(math.MaxInt64)
	case math.IsInf(float64(f), -1):
		return int64(math.MinInt64)
	}
	return int64(f)
}

func (f valueFloat) ToString() valueString {
	return asciiString(f.String())
}
func (f valueFloat) IsObject() bool {
	return false
}
func (f valueFloat) IsNumber() bool {
	return true
}

var matchLeading0Exponent = regexp.MustCompile(`([eE][\+\-])0+([1-9])`) // 1e-07 => 1e-7

func (f valueFloat) String() string {
	value := float64(f)
	if math.IsNaN(value) {
		return "NaN"
	} else if math.IsInf(value, 0) {
		if math.Signbit(value) {
			return "-Infinity"
		}
		return "Infinity"
	} else if f == _negativeZero {
		return "0"
	}
	exponent := math.Log10(math.Abs(value))
	if exponent >= 21 || exponent < -6 {
		return matchLeading0Exponent.ReplaceAllString(strconv.FormatFloat(value, 'g', -1, 64), "$1$2")
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (f valueFloat) ToFloat() float64 {
	return float64(f)
}

func (f valueFloat) ToBoolean() bool {
	return float64(f) != 0.0 && !math.IsNaN(float64(f))
}

func (f valueFloat) ToObject(r *Runtime) *Object {
	return r.newPrimitiveObject(f, r.global.NumberPrototype, "Number")
}

func (f valueFloat) ToNumber() Value {
	return f
}

func (f valueFloat) SameAs(other Value) bool {
	if o, ok := other.assertFloat(); ok {
		this := float64(f)
		if math.IsNaN(this) && math.IsNaN(o) {
			return true
		} else {
			ret := this == o
			if ret && this == 0 {
				ret = math.Signbit(this) == math.Signbit(o)
			}
			return ret
		}
	} else if o, ok := other.assertInt(); ok {
		this := float64(f)
		ret := this == float64(o)
		if ret && this == 0 {
			ret = !math.Signbit(this)
		}
		return ret
	}
	return false
}

func (f valueFloat) Equals(other Value) bool {
	if o, ok := other.assertFloat(); ok {
		return float64(f) == o
	}

	if o, ok := other.assertInt(); ok {
		return float64(f) == float64(o)
	}

	if _, ok := other.assertString(); ok {
		return float64(f) == other.ToFloat()
	}

	if o, ok := other.(valueBool); ok {
		return float64(f) == o.ToFloat()
	}

	if o, ok := other.(*Object); ok {
		return f.Equals(o.self.toPrimitiveNumber())
	}

	return false
}

func (f valueFloat) StrictEquals(other Value) bool {
	if o, ok := other.assertFloat(); ok {
		return float64(f) == o
	} else if o, ok := other.assertInt(); ok {
		return float64(f) == float64(o)
	}
	return false
}

func (f valueFloat) assertInt() (int, bool) {
	return 0, false
}
func (f valueFloat) assertUInt32() (uint32, bool) {
	return 0, false
}
func (f valueFloat) assertInt32() (int32, bool) {
	return 0, false
}
func (f valueFloat) assertInt64() (int64, bool) {
	return 0, false
}

func (f valueFloat) assertFloat() (float64, bool) {
	return float64(f), true
}

func (f valueFloat) assertString() (valueString, bool) {
	return nil, false
}

func (f valueFloat) baseObject(r *Runtime) *Object {
	return r.global.NumberPrototype
}

func (f valueFloat) Export() (interface{}, error) {
	return float64(f), nil
}

func (f valueFloat) ExportType() reflect.Type {
	return reflectTypeFloat
}

func (o *Object) MemUsage(ctx *MemUsageContext) (uint64, error) {
	if o.__wrapped != nil {
		return EmptySize, nil
	}

	switch x := o.self.(type) {
	case *objectGoReflect:
		return EmptySize, nil
	case *objectGoMapReflect:
		return EmptySize, nil
	case *objectGoMapSimple:
		return EmptySize, nil
	case *objectArrayBuffer:
		return EmptySize, nil
	case *objectGoSlice:
		return EmptySize, nil
	case *objectGoSliceReflect:
		return EmptySize, nil
	case *arrayObject:
		return x.MemUsage(ctx)
		// fmt.Println("it is an array")
		// return EmptySize, nil
	case *nativeFuncObject:
		// total := uint64(0)
		return x.MemUsage(ctx)
		// for k, v := range x.values {

		// }
		// x.baseFuncObject
		// return x.baseFuncObject.MemUsage(ctx)
		// for k, v := range  {
		// 	v.(*valueProperty
		// 	spew.Dump("what is v", v)

		// 	// inc, err := v.MemUsage(ctx)
		// 	// total += inc
		// 	total += uint64(len(k)) // count size of property name towards total object size.
		// 	// if err != nil {
		// 	// 	return total, err
		// 	// }
		// }
		// return total, nil
	case *baseObject:
		if o.self == nil {
			return 0, nil
		}
		// if ctx.IsObjVisited(o.self) {
		// 	return 0, nil
		// }
		// ctx.VisitObj(o.self)
		err := ctx.Descend()
		if err != nil {
			return 0, err
		}
		total, err := o.self.MemUsage(ctx)
		ctx.Ascend()
		return total, err
	case *lazyObject:
		if o.self == nil {
			return 0, nil
		}
		// err := ctx.Descend()
		// if err != nil {
		// 	return 0, err
		// }
		total, err := o.self.MemUsage(ctx)
		// ctx.Ascend()
		return total, err
	case *primitiveValueObject:
		if o.self == nil {
			return 0, nil
		}
		// err := ctx.Descend()
		// if err != nil {
		// 	return 0, err
		// }
		total, err := o.self.MemUsage(ctx)
		// ctx.Ascend()
		return total, err
	case *stringObject:
		if x.val == nil {
			return 0, nil
		}
		return x.MemUsage(ctx)
		// fmt.Println("descending lazy!!")
		// err := ctx.Descend()
		// if err != nil {
		// 	return 0, err
		// }
		// total, err := o.self.MemUsage(ctx)
		// ctx.Ascend()
		// return total, err
	default:
		// spew.Dump("what is default", o.self)

		// if err != nil {
		// 	return 0, err
		// }

		// spew.Dump("what is default", inner)
		// if _, ok := inner.(func(FunctionCall) Value); ok {
		// 	i++
		// 	fmt.Println("i is ", i)
		// }
		// // err = ctx.Descend()
		// // if err != nil {
		// // 	return 0, err
		// // }
		// switch val := inner.(type) {
		// // case Value:
		// // 	fmt.Println("descending!!")
		// // 	err := ctx.Descend()
		// // 	if err != nil {
		// // 		return 0, err
		// // 	}
		// // 	total, err := val.MemUsage(ctx)
		// // 	return total, err
		// case *baseObject:
		// 	fmt.Println("descending!!")
		// 	err := ctx.Descend()
		// 	if err != nil {
		// 		return 0, err
		// 	}
		// 	total, err := val.MemUsage(ctx)
		// 	ctx.Ascend()
		// 	return total, err
		// case *nativeFuncObject:
		// 	fmt.Println("descending!!")
		// 	err := ctx.Descend()
		// 	if err != nil {
		// 		return 0, err
		// 	}
		// 	total, err := val.MemUsage(ctx)
		// 	ctx.Ascend()
		// 	return total, err
		// default:

		// 	spew.Dump("inner type type?", o.self)
		// }
	}
	return 0, nil
}
func (o *Object) ToInt() int {
	return o.self.toPrimitiveNumber().ToNumber().ToInt()
}
func (o *Object) ToInt32() int32 {
	return o.self.toPrimitiveNumber().ToNumber().ToInt32()
}
func (o *Object) ToUInt32() uint32 {
	return o.self.toPrimitiveNumber().ToNumber().ToUInt32()
}
func (o *Object) ToInt64() int64 {
	return o.self.toPrimitiveNumber().ToNumber().ToInt64()
}

func (o *Object) ToString() valueString {
	return o.self.toPrimitiveString().ToString()
}

func (o *Object) String() string {
	return o.self.toPrimitiveString().String()
}

func (o *Object) ToFloat() float64 {
	return o.self.toPrimitiveNumber().ToFloat()
}

func (o *Object) ToBoolean() bool {
	return true
}

func (o *Object) ToObject(r *Runtime) *Object {
	return o
}
func (o *Object) IsObject() bool {
	return true
}
func (o *Object) IsNumber() bool {
	return false
}

func (o *Object) ToNumber() Value {
	return o.self.toPrimitiveNumber().ToNumber()
}

func (o *Object) SameAs(other Value) bool {
	if other, ok := other.(*Object); ok {
		return o == other
	}
	return false
}

func (o *Object) Equals(other Value) bool {
	if other, ok := other.(*Object); ok {
		return o == other || o.self.equal(other.self)
	}

	if _, ok := other.assertInt(); ok {
		return o.self.toPrimitive().Equals(other)
	}

	if _, ok := other.assertFloat(); ok {
		return o.self.toPrimitive().Equals(other)
	}

	if other, ok := other.(valueBool); ok {
		return o.Equals(other.ToNumber())
	}

	if _, ok := other.assertString(); ok {
		return o.self.toPrimitive().Equals(other)
	}
	return false
}

func (o *Object) StrictEquals(other Value) bool {
	if other, ok := other.(*Object); ok {
		return o == other || o.self.equal(other.self)
	}
	return false
}

func (o *Object) assertInt() (int, bool) {
	return 0, false
}
func (o *Object) assertUInt32() (uint32, bool) {
	return 0, false
}
func (o *Object) assertInt32() (int32, bool) {
	return 0, false
}
func (o *Object) assertInt64() (int64, bool) {
	return 0, false
}

func (o *Object) assertFloat() (float64, bool) {
	return 0, false
}

func (o *Object) assertString() (valueString, bool) {
	return nil, false
}

func (o *Object) baseObject(r *Runtime) *Object {
	return o
}

func (o *Object) Export() (interface{}, error) {
	if o.__wrapped != nil {
		return o.__wrapped, nil
	}

	if bo, ok := o.baseObject(o.runtime).self.(*objectGoReflect); ok {
		return bo.export()
	}

	return o.self.export()
}

func (o *Object) ExportType() reflect.Type {
	return o.self.exportType()
}

func (o *Object) Get(name string) (Value, error) {
	return o.self.getStr(name), nil
}

func (o *Object) Keys() (keys []string) {
	for item, f := o.self.enumerate(false, false)(); f != nil; item, f = f() {
		keys = append(keys, item.name)
	}

	return
}

// DefineDataProperty is a Go equivalent of Object.defineProperty(o, name, {value: value, writable: writable,
// configurable: configurable, enumerable: enumerable})
func (o *Object) DefineDataProperty(name string, value Value, writable, configurable, enumerable Flag) error {
	return tryFunc(func() {
		o.self.defineOwnProperty(newStringValue(name), propertyDescr{
			Value:        value,
			Writable:     writable,
			Configurable: configurable,
			Enumerable:   enumerable,
		}, true)
	})
}

// DefineAccessorProperty is a Go equivalent of Object.defineProperty(o, name, {get: getter, set: setter,
// configurable: configurable, enumerable: enumerable})
func (o *Object) DefineAccessorProperty(name string, getter, setter Value, configurable, enumerable Flag) error {
	return tryFunc(func() {
		o.self.defineOwnProperty(newStringValue(name), propertyDescr{
			Getter:       getter,
			Setter:       setter,
			Configurable: configurable,
			Enumerable:   enumerable,
		}, true)
	})
}

func (o *Object) Set(name string, value interface{}) error {
	return tryFunc(func() {
		o.self.putStr(name, o.runtime.ToValue(value), true)
	})
}

// MarshalJSON returns JSON representation of the Object. It is equivalent to JSON.stringify(o).
// Note, this implements json.Marshaler so that json.Marshal() can be used without the need to Export().
func (o *Object) MarshalJSON() ([]byte, error) {
	ctx := _builtinJSON_stringifyContext{
		r: o.runtime,
	}
	ex := o.runtime.vm.try(context.Background(), func() {
		if !ctx.do(o) {
			ctx.buf.WriteString("null")
		}
	})
	if ex != nil {
		return nil, ex
	}
	return ctx.buf.Bytes(), nil
}

// ClassName returns the class name
func (o *Object) Class() string {
	return o.self.className()
}

func (o valueUnresolved) throw() {
	fmt.Println("here we are great")
	o.r.throwReferenceError(o.ref)
}

func (o valueUnresolved) MemUsage(ctx *MemUsageContext) (uint64, error) {
	o.throw()
	return 0, nil
}
func (o valueUnresolved) ToInt() int {
	o.throw()
	return 0
}
func (o valueUnresolved) ToInt32() int32 {
	o.throw()
	return 0
}
func (o valueUnresolved) ToUInt32() uint32 {
	o.throw()
	return 0
}
func (o valueUnresolved) ToInt64() int64 {
	o.throw()
	return 0
}

func (o valueUnresolved) ToString() valueString {
	o.throw()
	return nil
}

func (o valueUnresolved) String() string {
	o.throw()
	return ""
}

func (o valueUnresolved) ToFloat() float64 {
	o.throw()
	return 0
}

func (o valueUnresolved) ToBoolean() bool {
	o.throw()
	return false
}

func (o valueUnresolved) ToObject(r *Runtime) *Object {
	o.throw()
	return nil
}

func (o valueUnresolved) ToNumber() Value {
	o.throw()
	return nil
}

func (o valueUnresolved) SameAs(other Value) bool {
	o.throw()
	return false
}

func (o valueUnresolved) Equals(other Value) bool {
	o.throw()
	return false
}

func (o valueUnresolved) StrictEquals(other Value) bool {
	o.throw()
	return false
}

func (o valueUnresolved) assertInt() (int, bool) {
	o.throw()
	return 0, false
}
func (o valueUnresolved) assertUInt32() (uint32, bool) {
	o.throw()
	return 0, false
}
func (o valueUnresolved) assertInt32() (int32, bool) {
	o.throw()
	return 0, false
}
func (o valueUnresolved) assertInt64() (int64, bool) {
	o.throw()
	return 0, false
}

func (o valueUnresolved) assertFloat() (float64, bool) {
	o.throw()
	return 0, false
}

func (o valueUnresolved) assertString() (valueString, bool) {
	o.throw()
	return nil, false
}

func (o valueUnresolved) baseObject(r *Runtime) *Object {
	o.throw()
	return nil
}

func (o valueUnresolved) Export() (interface{}, error) {
	o.throw()
	return nil, nil
}

func (o valueUnresolved) ExportType() reflect.Type {
	o.throw()
	return nil
}
func (o valueUnresolved) IsObject() bool {
	return false
}
func (o valueUnresolved) IsNumber() bool {
	return false
}

func init() {
	for i := 0; i < 256; i++ {
		intCache[i] = valueInt(i - 128)
	}
	_positiveZero = intToValue(0)
}
