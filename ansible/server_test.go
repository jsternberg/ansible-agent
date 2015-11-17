package ansible

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerPing(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()

	server := NewServer()
	server.ServeHTTP(res, req)

	if assert.Equal(200, res.Code) {
		var out map[string]string
		decoder := json.NewDecoder(res.Body)
		if err := decoder.Decode(&out); err != nil {
			t.Fatal(err)
		}

		assert.Equal(0, len(out))
	}
}

func TestServerExec(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{}
	form.Add("command", "echo hello world")

	req, err := http.NewRequest("POST", "/exec", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	server := NewServer()
	server.ServeHTTP(res, req)

	if assert.Equal(200, res.Code) {
		var out map[string]interface{}
		decoder := json.NewDecoder(res.Body)
		if err := decoder.Decode(&out); err != nil {
			t.Fatal(err)
		}

		status, ok := out["status"]
		if assert.True(ok, "missing 'status' from json response") {
			assert.Equal(0, int(status.(float64)))
		}
		stdin, ok := out["stdin"]
		if assert.True(ok, "missing 'stdin' from json response") {
			assert.Equal("", stdin.(string))
		}
		stdout, ok := out["stdout"]
		if assert.True(ok, "missing 'stdout' from json response") {
			assert.Equal("hello world\n", stdout.(string))
		}
		stderr, ok := out["stderr"]
		if assert.True(ok, "missing 'stderr' from json response") {
			assert.Equal("", stderr.(string))
		}
	}
}
