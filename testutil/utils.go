package testutil

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"

	crud "github.com/iov-one/cosmos-sdk-crud"
)

// CheckPrimaryKeyImmutability checks if the primary key of a crud.Object is dependent over an exported field
// It panics if it is the case, as crud.Object should have immutable primary keys
func CheckPrimaryKeyImmutability(object crud.Object) {

	pk := append([]byte{}, object.PrimaryKey()...)

	reflectedObj := reflect.ValueOf(object).Elem()
	mutateAllExportedFields(&reflectedObj)

	pkAfter := reflectedObj.MethodByName("PrimaryKey").Call(nil)[0].Bytes()

	if !bytes.Equal(pk, pkAfter) {
		panic(fmt.Errorf("primary key of type %T implementing crud.Object is not immutable", object))
	}
}

func mutateAllExportedFields(obj *reflect.Value) {

	// This object is not exported, skip it
	// Check if the field is mutable
	if !obj.CanSet() {
		// If not, try to use a copy, to make mutation work in maps
		cpy := reflect.Indirect(reflect.New(obj.Type()))
		*obj = cpy
		// If even with a copy it is not mutable, then it is not exported
		if !obj.CanSet() {
			return
		}
	}

	switch obj.Type().Kind() {
	case reflect.String:
		fallthrough
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		mutateSlice(obj)

	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint8:
		obj.SetUint(randomUint())

	case reflect.Int:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int8:
		obj.SetInt(int64(randomUint()))

	case reflect.Float64:
		fallthrough
	case reflect.Float32:
		obj.SetFloat(randomFloat())

	case reflect.Bool:
		obj.SetBool(randomBool())

	case reflect.Complex128:
		fallthrough
	case reflect.Complex64:
		obj.SetComplex(randomComplex())

	case reflect.Map:
		mutateMap(obj)

	case reflect.Func:
		//TODO: use functions
	case reflect.Ptr:
		fallthrough
	case reflect.UnsafePointer:
		//TODO: mutate pointers

	case reflect.Interface:
		fallthrough
	case reflect.Struct:
		for i := 0; i < obj.NumField(); i++ {
			field := obj.Field(i)
			mutateAllExportedFields(&field)
		}
	}
}

func randomFloat() float64 {
	return rand.Float64()
}

func randomUint() uint64 {
	return rand.Uint64()
}

func randomBool() bool {
	return rand.Intn(2) == 1
}

func randomComplex() complex128 {
	return complex(randomFloat(), randomFloat())
}

func mutateMap(obj *reflect.Value) {
	iter := obj.MapRange()
	for iter.Next() {
		val := iter.Value()
		mutateAllExportedFields(&val)
		obj.SetMapIndex(iter.Key(), val)
	}
}

func mutateSlice(obj *reflect.Value) {
	for i := 0; i < obj.Len(); i++ {
		val := obj.Index(i)
		mutateAllExportedFields(&val)
	}
}
