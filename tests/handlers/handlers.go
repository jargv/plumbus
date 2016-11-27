package handlers

import (
	"net/http"

	. "github.com/jargv/plumbus"
)

type ReturnStructResult struct {
	Message string
}

//go:generate plumbus ReturnStructHandler
func ReturnStructHandler() ReturnStructResult {
	return ReturnStructResult{
		Message: "Victory!",
	}
}

//go:generate plumbus ReturnErrorHandler
func ReturnErrorHandler() (string, error) {
	return "", Errorf(http.StatusBadRequest, "result")
}

var RequestBodyMessage string

type RequestBodyBody struct {
	Message string
}

//go:generate plumbus RequestBodyHandler
func RequestBodyHandler(body *RequestBodyBody) {
	RequestBodyMessage = body.Message
}

type ParamType string

func (p *ParamType) FromRequest(req *http.Request) error {
	*p = ParamType(req.URL.Query().Get("param"))
	return nil
}

var ParamParam1 string
var ParamParam2 string

//go:generate plumbus ParamHandler
func ParamHandler(p1 ParamType, p2 *ParamType) {
	ParamParam1 = string(p1)
	ParamParam2 = string(*p2)
}

//go:generate plumbus RequestMethodHandler
func RequestMethodHandler() string {
	return "nachos"
}

type userId string

func (ui *userId) FromRequest(req *http.Request) error {
	*ui = userId(req.URL.Query().Get("userId"))
	return nil
}

var PathParamsResult string

//go:generate plumbus PathParamsHandler
func PathParamsHandler(id userId) {
	PathParamsResult = string(id)
}

type foodQueryParam string
type amountQueryParam int

var RequiredRequestParamResult string
var RequiredRequestParamAmount int

//go:generate plumbus RequiredRequestParamHandler
func RequiredRequestParamHandler(food foodQueryParam, a amountQueryParam) {
	RequiredRequestParamAmount = int(a)
	RequiredRequestParamResult = string(food)
}

var (
	OptionalRequestParamResult string = "not set"
	OptionalRequestParamAmount int    = 0
)

//go:generate plumbus OptionalRequestParamHandler
func OptionalRequestParamHandler(amount *amountQueryParam, food *foodQueryParam) {
	if food != nil {
		OptionalRequestParamResult = string(*food)
	}
	if amount != nil {
		OptionalRequestParamAmount = int(*amount)
	}
}
