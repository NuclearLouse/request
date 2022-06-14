package request

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Address struct {
	Proto string
	Host  string
	// Path      string
	UserPass *url.Userinfo
	// ValuesURL url.Values
}

// http://www.example.com/api/v1/getUser?fname=foo&sname=bar
// Proto = http, Host = www.example.com, Path = api/v1/getUser, Values: fname=foo sname=bar
// "postgres://user:password@127.0.0.1/dbname?sslmode=disable"
// Proto = postgres, User = user, Pass = password, Host = 127.0.0.1, Path = dbname, Values = sslmode=disable
func NewAddress(proto, host string, userPass ...string) *Address {
	var userinfo *url.Userinfo
	switch len(userPass) {
	case 1:
		userinfo = url.UserPassword(userPass[0], "")
	case 2:
		userinfo = url.UserPassword(userPass[0], userPass[1])
	}
	return &Address{
		Proto:    proto,
		Host:     host,
		UserPass: userinfo,
	}
}

func (a *Address) SetEndpoint(path string, keyval ...interface{}) string {
	args := url.Values{}
	if len(keyval) >= 2 {
		switch len(keyval) {
		case 2:
			args.Set(fmt.Sprintf("%v", keyval[0]), fmt.Sprintf("%v", keyval[1]))
		default:
			for i := 0; i < len(keyval); i += 2 {
				if (i + 1) == len(keyval) {
					break
				}
				args.Set(fmt.Sprintf("%v", keyval[i]), fmt.Sprintf("%v", keyval[i+1]))
			}
		}
	}
	u := url.URL{
		Scheme:   a.Proto,
		Host:     a.Host,
		Path:     path,
		User:     a.UserPass,
		RawQuery: args.Encode(),
	}
	return u.String()
}

type Params struct {
	Method string
	URL    string
	Body   io.Reader
	Header map[string]string
	Client *http.Client
}

var defaultClient = &http.Client{
	Timeout: time.Duration(5) * time.Second,
}

//Default: Method = GET, Client.Timeout = 5s
func Do(p *Params) (*http.Response, error) {
	if p.Method == "" {
		p.Method = http.MethodGet
	}
	if p.Client == nil {
		p.Client = defaultClient
	}
	req, err := http.NewRequest(p.Method, p.URL, p.Body)
	if err != nil {
		return nil, err
	}

	if p.Header != nil {
		for key, val := range p.Header {
			req.Header.Set(key, val)
		}
	}
	return p.Client.Do(req)
}
