package generate

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type ConversionType int

const (
	ConvertBody ConversionType = iota
	ConvertError

	//optional values must be +1 of the non-optional version
	ConvertStringQueryParam
	ConvertOptionalStringQueryParam

	ConvertInterface
)

type Converter struct {
	ConversionType ConversionType
	Name           string
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
	if queryParamConverter := typeIsQueryParam(typ); queryParamConverter != nil {
		return queryParamConverter
	}

	interfaceType := reflect.TypeOf((*FromRequest)(nil)).Elem()
	if typ.Implements(interfaceType) || reflect.PtrTo(typ).Implements(interfaceType) {
		return &Converter{
			Type:           typ,
			IsPointer:      typ.Kind() == reflect.Ptr,
			ConversionType: ConvertInterface,
		}
	}

	return &Converter{
		Type:           typ,
		ConversionType: ConvertBody,
	}
}

func typeIsQueryParam(typ reflect.Type) *Converter {
	const suffix = "QueryParam"

	paramType := typ
	isOptional := false
	typeName := typ.Name()
	if typ.Kind() == reflect.Ptr {
		isOptional = true
		paramType = typ.Elem()
		typeName = paramType.Name()
	}

	if !strings.HasSuffix(typeName, suffix) {
		return nil
	}

	var conv ConversionType
	switch paramType.Kind() {
	case reflect.String:
		conv = ConvertStringQueryParam
	case reflect.Int:
		conv = ConvertStringQueryParam
	default:
		log.Fatalf("query parameter types must be string or int kind")
	}

	if isOptional {
		conv++
	}

	return &Converter{
		Name:           strings.TrimSuffix(typeName, suffix),
		ConversionType: conv,
		Type:           typ,
	}
}
