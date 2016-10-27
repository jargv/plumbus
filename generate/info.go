package generate

import (
	"errors"
	"reflect"
)

type Info struct {
	Inputs            []reflect.Type
	Outputs           []reflect.Type
	IsPointer         []bool
	RequestBodyIndex  int
	ResponseBodyIndex int
	RequestIndex      int
	ResponseIndex     int
	LastIsError       bool
}

func CollectInfo(typ reflect.Type) (*Info, error) {
	info := &Info{}

	nArgs := typ.NumIn()
	info.Inputs = make([]reflect.Type, nArgs)
	info.IsPointer = make([]bool, nArgs)
	info.RequestBodyIndex = -1
	for i := 0; i < nArgs; i++ {
		argType := typ.In(i)
		info.Inputs[i] = argType
		interfaceType := reflect.TypeOf((*FromRequest)(nil)).Elem()

		IsPointer := argType.Kind() == reflect.Ptr
		info.IsPointer[i] = IsPointer
		implementsInterface := argType.Implements(interfaceType)

		if !IsPointer && !implementsInterface {
			//it's possible an automatic reference will do the trick (like go does)
			implementsInterface = reflect.PtrTo(argType).Implements(interfaceType)
		}

		if !implementsInterface {
			if info.RequestBodyIndex != -1 {
				return nil, errors.New("plumbus.Handler: multiple args trying to use request body")
			}
			info.RequestBodyIndex = i
		}
	}

	nResults := typ.NumOut()
	info.Outputs = make([]reflect.Type, nResults)
	info.LastIsError = false
	info.ResponseBodyIndex = -1
	resultTypes := []reflect.Type{}
	for i := 0; i < nResults; i++ {
		out := typ.Out(i)
		info.Outputs[i] = out
		resultTypes = append(resultTypes, out)
		info.LastIsError = out.Implements(reflect.TypeOf((*error)(nil)).Elem())
		if !out.Implements(reflect.TypeOf((*ToResponse)(nil)).Elem()) && !info.LastIsError {
			if info.ResponseBodyIndex != -1 {
				return nil, errors.New("plumbus.Handler: multiple results trying to use response body")
			}
			info.ResponseBodyIndex = i
		}
	}

	return info, nil
}
