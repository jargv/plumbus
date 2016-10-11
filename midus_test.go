package midus

import (
	"bytes"
	"encoding/json"
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
