// Copyright 2012 by sdm. All rights reserved.
// license that can be found in the LICENSE file.

package wk

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type RouteData map[string]string

func (r RouteData) Int(name string) (int, bool) {
	if s, ok := r[name]; ok {
		if i, err := strconv.Atoi(s); err == nil {
			return i, true
		}
	}
	return 0, false
}

func (r RouteData) IntOr(name string, v int) int {
	if s, ok := r[name]; ok {
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}
	}
	return v
}

func (r RouteData) Bool(name string) (bool, bool) {
	if s, ok := r[name]; ok {
		if b, err := strconv.ParseBool(s); err != nil {
			return b, true
		}
	}
	return false, false
}

func (r RouteData) BoolOr(name string, v bool) bool {
	if s, ok := r[name]; ok {
		if b, err := strconv.ParseBool(s); err != nil {
			return b
		}
	}
	return v
}

func (r RouteData) Float(name string) (float64, bool) {
	if s, ok := r[name]; ok {
		if f, err := strconv.ParseFloat(s, 64); err != nil {
			return f, true
		}
	}
	return 0, false
}

func (r RouteData) FloatOr(name string, v float64) float64 {
	if s, ok := r[name]; ok {
		if f, err := strconv.ParseFloat(s, 64); err != nil {
			return f
		}
	}
	return v
}

func (r RouteData) Str(name string) (string, bool) {
	if s, ok := r[name]; ok {
		return s, true
	}
	return "", false
}

func (r RouteData) StrOr(name string, v string) string {
	if s, ok := r[name]; ok {
		return s
	}
	return v
}

// http context
type HttpContext struct {

	// http request 
	Request *http.Request

	// http response writer
	Resonse http.ResponseWriter

	// http method
	Method string //http method

	// request path
	RequestPath string

	// physical Path
	PhysicalPath string

	// route data
	RouteData RouteData

	// view Data
	ViewData map[string]interface{}

	// http result    
	Result HttpResult

	// last error
	Error error

	// Flash Variables
	Flash map[string]interface{}
}

func (ctx *HttpContext) String() string {
	return fmt.Sprintf("%s %s %v %v \n;", ctx.Method, ctx.RequestPath, ctx.Result, ctx.Error)
}

// RouteValue return route data value
func (ctx *HttpContext) RouteValue(name string) (string, bool) {
	v, ok := ctx.RouteData[name]
	return v, ok
}

// FormValue alias of Request FormValue
func (ctx *HttpContext) FormValue(name string) string {
	return ctx.Request.FormValue(name)
}

func (ctx *HttpContext) FormInt(name string) (int, bool) {
	if s := ctx.FormValue(name); s != "" {
		if i, err := strconv.Atoi(s); err == nil {
			return i, true
		}
	}
	return 0, false
}

func (ctx *HttpContext) FormIntOr(name string, v int) int {
	if s := ctx.FormValue(name); s != "" {
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}
	}
	return v
}

func (ctx *HttpContext) FormBool(name string) (bool, bool) {
	if s := ctx.FormValue(name); s != "" {
		if b, err := strconv.ParseBool(s); err != nil {
			return b, true
		}
	}
	return false, false
}

func (ctx *HttpContext) FormBoolOr(name string, v bool) bool {
	if s := ctx.FormValue(name); s != "" {
		if b, err := strconv.ParseBool(s); err != nil {
			return b
		}
	}
	return v
}

func (ctx *HttpContext) FormFloat(name string) (float64, bool) {
	if s := ctx.FormValue(name); s != "" {
		if f, err := strconv.ParseFloat(s, 64); err != nil {
			return f, true
		}
	}
	return 0, false
}

func (ctx *HttpContext) FormFloatOr(name string, v float64) float64 {
	if s := ctx.FormValue(name); s != "" {
		if f, err := strconv.ParseFloat(s, 64); err != nil {
			return f
		}
	}
	return v
}

// // QueryValue return value from request URL query 
// func (ctx *HttpContext) QueryValue(name string) []string {
// 	return ctx.Request.URL.Query()[name]
// }

// UserAgent return request User-Agent header
func (ctx *HttpContext) UserAgent() string {
	return ctx.Request.Header.Get(HeaderUserAgent)
}

// Header return request header by name
func (ctx *HttpContext) Header(name string) string {
	return ctx.Request.Header.Get(name)
}

// SetHeader set resonse http header
func (ctx *HttpContext) SetHeader(key string, value string) {
	ctx.Resonse.Header().Set(key, value)
}

// AddHeader add response http header
func (ctx *HttpContext) AddHeader(key string, value string) {
	ctx.Resonse.Header().Add(key, value)
}

// ContentType set response Content-Type header
func (ctx *HttpContext) ContentType(ctype string) {
	ctx.Resonse.Header().Set(HeaderContentType, ctype)
}

// Status write status code to http header
func (ctx *HttpContext) Status(code int) {
	ctx.Resonse.WriteHeader(code)
}

// Accept return request Accept header
func (ctx *HttpContext) Accept() string {
	return ctx.Request.Header.Get(HeaderAccept)
}

// Write writes data to resposne
func (ctx *HttpContext) Write(b []byte) (int, error) {
	return ctx.Resonse.Write(b)
}

// Expires set reponse Expires header
func (ctx *HttpContext) Expires(t string) {
	ctx.SetHeader(HeaderExpires, t)
}

// SetCookie set cookie to response
func (ctx *HttpContext) SetCookie(cookie *http.Cookie) {
	http.SetCookie(ctx.Resonse, cookie)
}

// Cookie return cookie from request
func (ctx *HttpContext) Cookie(name string) (*http.Cookie, error) {
	return ctx.Request.Cookie(name)
}

// Flush flush response immediately
func (ctx *HttpContext) Flush() {
	// _, buf, _ := ctx.Resonse.(http.Hijacker).Hijack()
	// if buf != nil {
	// 	buf.Flush()
	// }
	f, ok := ctx.Resonse.(http.Flusher)
	if ok {
		f.Flush()
	}
}

// SetFlash
func (ctx *HttpContext) SetFlash(key string, v interface{}) {
	if ctx.Flash == nil {
		ctx.Flash = make(map[string]interface{})
	}
	ctx.Flash[key] = v
}

func (ctx *HttpContext) ReadBody() ([]byte, error) {
	return ioutil.ReadAll(ctx.Request.Body)
}
