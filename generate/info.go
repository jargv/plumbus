package generate

import (
	"fmt"
	"reflect"
)

type ConversionType int

const (
	ConvertBody ConversionType = iota
	ConvertError
	ConvertQueryParam
	ConvertInterface
)

type Converter struct {
	ConversionType ConversionType
	Type           reflect.Type
	IsPointer      bool
}

type Info struct {
	Inputs            []*Converter
	Outputs           []*Converter
	ResponseBodyIndex int
	LastIsError       bool
}

func CollectInfo(typ reflect.Type) (*Info, error) {
	if typ.Kind() != reflect.Func {
		return nil, fmt.Errorf(
			"internal plumbus error: expected value of kind 'function', got %s",
			typ.Kind(),
		)
	}

	info := &Info{
		ResponseBodyIndex: -1,
	}

	for i := 0; i < typ.NumIn(); i++ {
		info.Inputs = append(info.Inputs, inputConverter(typ.In(i)))
	}

	for i := 0; i < typ.NumOut(); i++ {
		output := outputConverter(typ.Out(i))
		info.Outputs = append(info.Outputs, output)
		if i == typ.NumOut()-1 {
			info.LastIsError = output.ConversionType == ConvertError
		}
		if output.ConversionType == ConvertBody {
			//todo: check for multiples here
			info.ResponseBodyIndex = i
		}
	}

	return info, nil
}

func outputConverter(typ reflect.Type) *Converter {
	conv := &Converter{
		Type: typ,
	}

	interfaceType := reflect.TypeOf((*ToResponse)(nil)).Elem()
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	switch true {
	case typ.Implements(interfaceType) || reflect.PtrTo(typ).Implements(interfaceType):
		conv.ConversionType = ConvertInterface
	case typ.Implements(errorType):
		conv.ConversionType = ConvertError
	default:
		conv.ConversionType = ConvertBody
	}

	return conv
}

func inputConverter(typ reflect.Type) *Converter {
	conv := &Converter{
		Type:      typ,
		IsPointer: typ.Kind() == reflect.Ptr,
	}

	interfaceType := reflect.TypeOf((*FromRequest)(nil)).Elem()

	switch true {
	case typ.Implements(interfaceType) || reflect.PtrTo(typ).Implements(interfaceType):
		conv.ConversionType = ConvertInterface
	default:
		conv.ConversionType = ConvertBody
	}

	return conv
}
