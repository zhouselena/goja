package goja

import (
	"hash/maphash"
	"reflect"
	"strconv"

	"github.com/dop251/goja/unistring"
)

func (i valueNumber) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return SizeNumber, nil
}

func (i valueNumber) ToInteger() int64 {
	return i.ToInt64()
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
		return v
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

func (i valueNumber) hash(*maphash.Hash) uint64 {
	return uint64(i.ToInt64())
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

func (i valueNumber) IsObject() bool {
	return false
}

func (i valueNumber) IsNumber() bool {
	return true
}

func (i valueNumber) ToString() Value {
	return asciiString(i.String())
}
func (i valueNumber) toString() valueString {
	return asciiString(i.String())
}
func (i valueNumber) String() string {
	return strconv.FormatInt(i.ToInt64(), 10)
}

func (i valueNumber) string() unistring.String {
	return unistring.String(i.String())
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
		return i.ToFloat() == o
	}
	if o, ok := other.assertString(); ok {
		return o.ToNumber().Equals(i)
	}
	if o, ok := other.(valueBool); ok {
		return i.ToInt() == o.ToInt()
	}
	if o, ok := other.(*Object); ok {
		return i.Equals(o.self.toPrimitiveNumber())
	}
	return false
}

func (i valueNumber) StrictEquals(other Value) bool {
	if otherInt, ok := other.assertInt(); ok {
		return i.ToInt() == otherInt
	} else if otherFloat, ok := other.assertFloat(); ok {
		return i.ToFloat() == otherFloat
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

func (i valueNumber) Export() interface{} {
	return i.val
}

func (i valueNumber) ExportType() reflect.Type {
	return i._type
}

func (i valueUInt32) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return SizeInt32, nil
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
func (i valueUInt32) ToInteger() int64 {
	return int64(i)
}

func (i valueUInt32) ToString() Value {
	return asciiString(i.String())
}
func (i valueUInt32) toString() valueString {
	return asciiString(i.String())
}
func (i valueUInt32) String() string {
	return strconv.FormatInt(i.ToInteger(), 10)
}

func (i valueUInt32) string() unistring.String {
	return unistring.String(i.String())
}
func (i valueUInt32) IsObject() bool {
	return false
}
func (i valueUInt32) IsNumber() bool {
	return true
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
	if o, ok := other.assertUInt32(); ok {
		return uint32(i) == o
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
	if otherInt, ok := other.assertUInt32(); ok {
		return uint32(i) == otherInt
	} else if otherFloat, ok := other.assertFloat(); ok {
		return float64(i) == otherFloat
	}
	return false
}

func (i valueUInt32) hash(*maphash.Hash) uint64 {
	return uint64(i)
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

func (i valueUInt32) Export() interface{} {
	return uint32(i)
}

func (i valueUInt32) ExportType() reflect.Type {
	return reflectTypeUInt32
}

func (i valueInt32) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return SizeInt32, nil
}

func (i valueInt32) hash(*maphash.Hash) uint64 {
	return uint64(i)
}

func (i valueInt32) ToInt() int {
	return int(i)
}
func (i valueInt32) ToInteger() int64 {
	return int64(i)
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

func (i valueInt32) ToString() Value {
	return asciiString(i.String())
}
func (i valueInt32) toString() valueString {
	return asciiString(i.String())
}
func (i valueInt32) String() string {
	return strconv.FormatInt(i.ToInteger(), 10)
}

func (i valueInt32) string() unistring.String {
	return unistring.String(i.String())
}
func (i valueInt32) IsObject() bool {
	return false
}
func (i valueInt32) IsNumber() bool {
	return true
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

func (i valueInt32) Export() interface{} {
	return int32(i)
}

func (i valueInt32) ExportType() reflect.Type {
	return reflectTypeInt32
}

func (i valueInt64) MemUsage(ctx *MemUsageContext) (uint64, error) {
	return SizeNumber, nil
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
func (i valueInt64) ToInteger() int64 {
	return int64(i)
}

func (i valueInt64) ToString() Value {
	return asciiString(i.String())
}
func (i valueInt64) toString() valueString {
	return asciiString(i.String())
}
func (i valueInt64) String() string {
	return strconv.FormatInt(i.ToInteger(), 10)
}

func (i valueInt64) string() unistring.String {
	return unistring.String(i.String())
}
func (i valueInt64) IsObject() bool {
	return false
}

func (i valueInt64) IsNumber() bool {
	return true
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
	return int(i), true
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

func (i valueInt64) hash(*maphash.Hash) uint64 {
	return uint64(i)
}

func (i valueInt64) baseObject(r *Runtime) *Object {
	return r.global.NumberPrototype
}

func (i valueInt64) Export() interface{} {
	return int64(i)
}

func (i valueInt64) ExportType() reflect.Type {
	return reflectTypeInt64
}
