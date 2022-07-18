package fans

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	applicationJson = "application/json"
	applicationForm = "application/x-www-form-urlencoded"
)

// E 一个键值对
type E struct {
	name  string
	value interface{}
}

// ES 一组有顺序的键值对
type ES []E

func (e ES) Map() M {
	m := make(M)
	for i := range e {
		m[e[i].name] = e[i].value
	}
	return m
}

// M 一组无顺序的键值对
type M map[string]interface{}

func (m M) Param() string {
	v := make(url.Values)
	for k := range m {
		v.Add(k, fmt.Sprint(m[k]))
	}
	return v.Encode()
}

type Request interface {
	Get(u string, param M, body io.Reader, header ...E) (*gjson.Result, error)
	Post(u string, param M, body io.Reader, header ...E) (*gjson.Result, error)
}

type R struct {
	c         *http.Client
	reqHeader map[string]string
}

func NewR(timeout int) *R {
	c := http.Client{Timeout: time.Duration(timeout) * time.Second}
	return &R{
		c: &c,
		reqHeader: map[string]string{
			"User-Agent":      "Mozilla/5.0 BiliDroid/6.73.1 (bbcallen@gmail.com) os/android model/Mi 10 Pro mobi_app/android build/6731100 channel/xiaomi innerVer/6731110 osVer/12 network/2",
			"Accept":          applicationJson,
			"Accept-Language": "zh-CN,zh;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
		},
	}
}

func (r *R) request(method string, u string, param M,
	body io.Reader, header ...E) (*gjson.Result, error) {
	var req *http.Request
	var err error
	req, err = http.NewRequest(method, u+"?"+param.Param(), body)
	if err != nil {
		return nil, err
	}
	h := req.Header
	rh := r.reqHeader
	for k := range rh {
		h.Add(k, rh[k])
	}
	for i := range header {
		h.Add(header[i].name, fmt.Sprint(header[i].value))
	}
	return handResp(r.c.Do(req))
}

func (r *R) Get(u string, param M, body io.Reader, header ...E) (*gjson.Result, error) {
	return r.request(http.MethodGet, u, param, body, header...)
}

func (r *R) Post(u string, param M, body io.Reader, header ...E) (*gjson.Result, error) {
	return r.request(http.MethodPost, u, param, body, header...)
}
