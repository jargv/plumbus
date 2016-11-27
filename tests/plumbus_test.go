package plumbus

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/jargv/plumbus"
	. "github.com/jargv/plumbus/tests/handlers"
)

func TestReturnStruct(t *testing.T) {
	handler := HandlerFunc(ReturnStructHandler)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("couldn't get: %#v\n", err)
	}

	var result ReturnStructResult
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&result)
	if err != nil {
		t.Errorf("couldn't decode: %#v\n", err)
	}

	if result.Message != "Victory!" {
		t.Errorf(`body != "Victory!", body == %q`, result.Message)
	}
}

func TestReturnError(t *testing.T) {
	handler := HandlerFunc(ReturnErrorHandler)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("couldn't get: %#v\n", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Error("statusCode != 'http.StatusBadRequest', statusCode == %d", resp.StatusCode)
	}
}

func TestRequestBody(t *testing.T) {
	handler := HandlerFunc(RequestBodyHandler)

	server := httptest.NewServer(handler)
	defer server.Close()

	bytes := bytes.Buffer{}
	json.NewEncoder(&bytes).Encode(&RequestBodyBody{
		Message: "full circle!",
	})

	_, err := http.Post(server.URL, "", &bytes)
	if err != nil {
		t.Errorf("couldn't make request: %#v\n", err)
	}

	if RequestBodyMessage != "full circle!" {
		t.Errorf(`RequestBodyMessage != "full circle", RequestBodyMessage == %q`, RequestBodyMessage)
	}
}

func TestRequestParam(t *testing.T) {
	handler := HandlerFunc(ParamHandler)

	server := httptest.NewServer(handler)
	defer server.Close()

	http.Get(server.URL + `?param=awesome`)

	if ParamParam1 != "awesome" {
		t.Errorf(`ParamParam1 != "awesome", ParamParam1 == %q`, ParamParam1)
	}

	if ParamParam2 != "awesome" {
		t.Errorf(`ParamParam2 != "awesome", ParamParam2 == %q`, ParamParam2)
	}
}

func TestRequestMethod(t *testing.T) {
	handler := HandlerFunc(&ByMethod{
		PUT: RequestMethodHandler,
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf(": %#v\n", err)
	}

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf(
			`StatusCode != http.StatusMethodNotAllowed, StatusCode = %s`,
			resp.Status,
		)
	}

	req, err := http.NewRequest("PUT", server.URL, nil)
	if err != nil {
		t.Errorf("couldn't make request: %#v\n", err)
	}

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("couldn't make request: %#v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf(
			`StatusCode != http.StatusOK, StatusCode = %s`,
			resp.Status,
		)
	}
}

func TestPathParams(t *testing.T) {
	mux := NewServeMux()
	mux.Handle("/user/:userId/name", PathParamsHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	_, err := http.Get(server.URL + "/user/10/name")
	if err != nil {
		t.Fatalf("couldn't make request: %v\n", err)
	}

	if PathParamsResult != "10" {
		t.Fatalf(`PathParamsResult != "10", PathParamsResult == "%v"`, PathParamsResult)
	}
}

func TestRequiredRequestParam(t *testing.T) {
	server := httptest.NewServer(HandlerFunc(RequiredRequestParamHandler))

	//test that it's required (we should get a StatusBadRequest)
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("making request: %v\n", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf(`resp.StatusCode != htp.StatusBadRequest, resp.StatusCode == "%v"`, resp.StatusCode)
	}

	//test that it's required (we should get a StatusBadRequest)
	resp, err = http.Get(server.URL + "?food=nachos")
	if err != nil {
		t.Fatalf("making request: %v\n", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf(`resp.StatusCode != htp.StatusBadRequest, resp.StatusCode == "%v"`, resp.StatusCode)
	}

	//test that it's converted
	_, err = http.Get(server.URL + "?food=nachos&amount=10")
	if err != nil {
		t.Fatalf("makeing request: %v\n", err)
	}

	if RequiredRequestParamResult != "nachos" {
		t.Fatalf(
			`RequiredRequestParamResult != "nachos", RequiredRequestParamResult == "%v"`,
			RequiredRequestParamResult,
		)
	}

	if RequiredRequestParamAmount != 10 {
		t.Fatalf(
			`RequiredRequestParamAmount != 10, RequiredRequestParamAmount == "%v"`,
			RequiredRequestParamAmount,
		)
	}
}

func TestOptionalRequestParam(t *testing.T) {
	server := httptest.NewServer(HandlerFunc(OptionalRequestParamHandler))

	//test that it's not required
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("makeing request: %v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`resp.StatusCode != http.StatusOK, resp.StatusCode == "%v"`, resp.StatusCode)
	}

	if OptionalRequestParamResult != "not set" {
		t.Fatalf(
			`OptionalRequestParamResult != "not set", OptionalRequestParamResult == "%v"`,
			OptionalRequestParamResult,
		)
	}

	//test that it's not required (on the second, int param)
	resp, err = http.Get(server.URL + "?food=nachos")
	if err != nil {
		t.Fatalf("makeing request: %v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`resp.StatusCode != http.StatusOK, resp.StatusCode == "%v"`, resp.StatusCode)
	}

	if OptionalRequestParamResult != "nachos" {
		t.Fatalf(
			`OptionalRequestParamResult != "not set", OptionalRequestParamResult == "%v"`,
			OptionalRequestParamResult,
		)
	}

	if OptionalRequestParamAmount != 0 {
		t.Fatalf(
			`OptionalRequestParamAmount != 0, OptionalRequestParamAmount == "%v"`,
			OptionalRequestParamAmount,
		)
	}

	//test that it's passed to the handler
	_, err = http.Get(server.URL + "?food=nachos&amount=10")
	if err != nil {
		t.Fatalf("makeing request: %v\n", err)
	}

	if OptionalRequestParamResult != "nachos" {
		t.Fatalf(
			`OptionalRequestParamResult != "nachos", OptionalRequestParamResult == "%v"`,
			OptionalRequestParamResult,
		)
	}

	if OptionalRequestParamAmount != 10 {
		t.Fatalf(
			`OptionalRequestParamAmount != 10, OptionalRequestParamAmount == "%v"`,
			OptionalRequestParamAmount,
		)
	}
}

// // type UserId struct {
// // }

// // func (ui *UserId) FromRequest(req *http.Request) error {
// // 	return nil
// // }

// // type User struct {
// // 	Name string `json:"name"`
// // 	Age  int    `json:"age"`
// // }

// // type UserRepo struct {
// // }

// // func (ur *UserRepo) FindById(id UserId) (*User, error) {
// // 	return nil, nil
// // }

// // func (ur *UserRepo) Edit(id UserId, user *User) error {
// // 	return nil
// // }

// // func TestDocumentation(t *testing.T) {
// // 	mux := NewServeMux()
// // 	type user struct {
// // 		Name string
// // 		Age  int
// // 	}

// // 	type result struct {
// // 		Role   string
// // 		Id     int
// // 		User   *user
// // 		Thing1 *int
// // 		Thing2 []int
// // 		Thing3 []*int
// // 		Thing4 []**int
// // 		Thing5 map[string]*user
// // 	}

// // 	users := UserRepo{}

// // 	mux.Handle("/users/:userId/details", func(u user) *result {
// // 		return nil
// // 	})

// // 	mux.Handle("/users/:userId", ByMethod{
// // 		GET: users.FindById,
// // 		PUT: users.Edit,
// // 	})

// // 	mux.Handle("/standerd/handler", func(http.ResponseWriter, *http.Request) {})

// // 	mux.Handle("/any/body", func(interface{}) {})

// // 	docs := mux.Documentation()

// // 	bytes, _ := json.MarshalIndent(docs, "", "  ")
// // 	log.Printf("string(bytes):\n%s", string(bytes))
// // }
