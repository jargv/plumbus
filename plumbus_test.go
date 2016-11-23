package plumbus

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReturnStruct(t *testing.T) {
	type Result struct {
		Message string
	}

	handler := HandlerFunc(func() Result {
		return Result{
			Message: "Victory!",
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("couldn't get: %#v\n", err)
	}

	var result Result
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
	handler := HandlerFunc(func() (string, error) {
		return "", Errorf(http.StatusBadRequest, "result")
	})

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
	type Body struct {
		Message string
	}

	var message string

	handler := HandlerFunc(func(body *Body) {
		message = body.Message
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	bytes := bytes.Buffer{}
	enc := json.NewEncoder(&bytes)
	enc.Encode(&Body{
		Message: "full circle!",
	})

	req, err := http.NewRequest("POST", server.URL, &bytes)
	if err != nil {
		t.Errorf("couldn't make request: %#v\n", err)
	}
	client := http.Client{}
	client.Do(req)

	if message != "full circle!" {
		t.Errorf(`message != "full circle", message == %q`, message)
	}
}

type Param string

func (p *Param) FromRequest(req *http.Request) error {
	*p = Param(req.URL.Query().Get("param"))
	return nil
}

func TestRequestParam(t *testing.T) {
	var param string

	handler := HandlerFunc(func(p Param) {
		param = string(p)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	http.Get(server.URL + `?param=awesome`)

	if param != "awesome" {
		t.Errorf(`param != "awesome", param == %q`, param)
	}
}

func TestRequestMethod(t *testing.T) {
	handler := HandlerFunc(&ByMethod{
		PUT: HandlerFunc(func() string {
			return "nachos"
		}),
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

type userId string

func (ui *userId) FromRequest(req *http.Request) error {
	*ui = userId(req.URL.Query().Get("userId"))
	return nil
}

func TestPathParams(t *testing.T) {
	var result string
	mux := NewServeMux()
	mux.Handle("/user/:userId/name", func(id userId) {
		result = string(id)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	_, err := http.Get(server.URL + "/user/10/name")
	if err != nil {
		t.Fatalf("couldn't make request: %v\n", err)
	}

	if result != "10" {
		t.Fatalf(`result != "10", result == "%v"`, result)
	}
}

type UserId struct {
}

func (ui *UserId) FromRequest(req *http.Request) error {
	return nil
}

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type UserRepo struct {
}

func (ur *UserRepo) FindById(id UserId) (*User, error) {
	return nil, nil
}

func (ur *UserRepo) Edit(id UserId, user *User) error {
	return nil
}

func TestDocumentation(t *testing.T) {
	mux := NewServeMux()
	type user struct {
		Name string
		Age  int
	}

	type result struct {
		Role   string
		Id     int
		User   *user
		Thing1 *int
		Thing2 []int
		Thing3 []*int
		Thing4 []**int
		Thing5 map[string]*user
	}

	users := UserRepo{}

	mux.Handle("/users/:userId/details", func(u user) *result {
		return nil
	})

	mux.Handle("/users/:userId", ByMethod{
		GET: users.FindById,
		PUT: users.Edit,
	})

	mux.Handle("/standerd/handler", func(http.ResponseWriter, *http.Request) {})

	mux.Handle("/any/body", func(interface{}) {})

	docs := mux.Documentation()

	bytes, _ := json.MarshalIndent(docs, "", "  ")
	log.Printf("string(bytes):\n%s", string(bytes))
}
