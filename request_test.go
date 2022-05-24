package request

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeEndpointAddress(t *testing.T) {
	endpoint := "dbname"

	testCases := []struct {
		name     string
		address  *Address
		keyval   []interface{}
		expected string
	}{
		{
			name:     "Эндпоинт без параметров запроса",
			address:  NewAddress("postgres", "localhost"),
			keyval:   nil,
			expected: "postgres://localhost/dbname",
		},
		{
			name:     "Эндпоинт с параметрами запроса",
			address:  NewAddress("postgres", "localhost"),
			keyval:   []interface{}{"sslmode", "disable"},
			expected: "postgres://localhost/dbname?sslmode=disable",
		},
		{
			name:     "Эндпоинт с юзер-инфо и параметрами",
			address:  NewAddress("postgres", "localhost", "user", "password"),
			keyval:   []interface{}{"sslmode", "disable"},
			expected: "postgres://user:password@localhost/dbname?sslmode=disable",
		},
		{
			name:     "Эндпоинт с юзер-инфо и нечетным числом параметров",
			address:  NewAddress("postgres", "localhost", "user", "password"),
			keyval:   []interface{}{"sslmode", "disable", "pool_max_conns"},
			expected: "postgres://user:password@localhost/dbname?sslmode=disable",
		},
		{
			name:     "Зндпоинт с разнотипными параметрами",
			address:  NewAddress("postgres", "localhost", "user", "password"),
			keyval:   []interface{}{"sslmode", "disable", "pool_max_conns", 25, "bool_key", true},
			expected: "postgres://user:password@localhost/dbname?bool_key=true&pool_max_conns=25&sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.address.SetEndpoint(endpoint, tc.keyval...).String())
		})
	}
}

func TestRequest(t *testing.T) {
	addr := NewAddress("http", "pie.dev")

	res, err := DoRequestDefault(addr.SetEndpoint("get").String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode, "Дефолтный запрос")

	type bodyRequest struct {
		Key string `json:"key"`
	}

	body := bodyRequest{
		Key: "value",
	}

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(body); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name           string
		params         *Params
		expectedCode   int
		expectedBody   string
		expectedHeader string
	}{
		{
			name: "Не дефолтный метод с телом запроса",
			params: &Params{
				Method: http.MethodPost,
				URL:    addr.SetEndpoint("post").URL(),
				Body:   b,
			},
			expectedCode:   http.StatusOK,
			expectedBody:   "value",
			expectedHeader: "Go-http-client/1.1",
		},
		{
			name: "С установкой хедера",
			params: &Params{
				Method: http.MethodGet,
				URL:    addr.SetEndpoint("headers").URL(),
				Header: map[string]string{
					"User-Agent": "Bacon/1.0",
				},
				Body: nil,
			},
			expectedCode:   http.StatusOK,
			expectedHeader: "Bacon/1.0",
		},
	}

	type bodyResponse struct {
		Headers struct {
			Cookie    string `json:"Cookie"`
			UserAgent string `json:"User-Agent"`
			XFoo      string `json:"X-Foo"`
		} `json:"headers"`
		Json struct {
			Key string `json:"key"`
		} `json:"json"`
		Origin string   `json:"origin"`
		Url    string   `json:"url"`
		Args   struct{} `json:"args,omitempty"`
		Data   string   `json:"data,omitempty"`
		Files  struct{} `json:"files,omitempty"`
		Form   struct{} `json:"form,omitempty"`
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := DoRequestWithParams(tc.params)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expectedCode, res.StatusCode)

			var body bodyResponse
			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			assert.Equal(t, tc.expectedHeader, body.Headers.UserAgent)
			assert.Equal(t, tc.expectedBody, body.Json.Key)
		})
	}
}

/*
	prms := &Params{
		Method: http.MethodPost,
		URL:    addr.SetEndpoint("post").URL(),
	}

	res, err = prms.DoRequestWithParams()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, res.StatusCode)

	prms = &Params{
		Method: http.MethodGet,
		URL:    addr.SetEndpoint("headers").URL(),
		Header: map[string]string{
			"Cookie":     "valued-visitor=yes;foo=bar",
			"User-Agent": "Bacon/1.0",
			"X-Foo":      "Bar",
		},
	}

	res, err = prms.DoRequestWithParams()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, res.StatusCode)
	var bodyJSON struct {
		Headers struct {
			Cookie    string `json:"Cookie"`
			UserAgent string `json:"User-Agent"`
			XFoo      string `json:"X-Foo"`
		} `json:"headers"`
		JSON struct {
			Key string `json:"key"`
		} `json:"json"`
	}
	if err := json.NewDecoder(res.Body).Decode(&bodyJSON); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "valued-visitor=yes;foo=bar", bodyJSON.Headers.Cookie)
*/
