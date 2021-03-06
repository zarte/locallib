// Copyright 2013 Martini Authors
// Copyright 2015 The Macaron Authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gzip

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zarte/macaron.v1"
)

func Test_Gzip(t *testing.T) {
	Convey("Gzip response content", t, func() {
		before := false

		m := macaron.New()
		m.Use(Gziper(Options{-10}))
		m.Use(func(r http.ResponseWriter) {
			r.(macaron.ResponseWriter).Before(func(rw macaron.ResponseWriter) {
				before = true
			})
		})
		m.Get("/", func() string { return "hello wolrd!" })

		// Not yet gzip.
		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)
		m.ServeHTTP(resp, req)

		_, ok := resp.HeaderMap[_HEADER_CONTENT_ENCODING]
		So(ok, ShouldBeFalse)

		ce := resp.Header().Get(_HEADER_CONTENT_ENCODING)
		So(strings.EqualFold(ce, "gzip"), ShouldBeFalse)

		// Gzip now.
		resp = httptest.NewRecorder()
		req.Header.Set(_HEADER_ACCEPT_ENCODING, "gzip")
		m.ServeHTTP(resp, req)

		_, ok = resp.HeaderMap[_HEADER_CONTENT_ENCODING]
		So(ok, ShouldBeTrue)

		ce = resp.Header().Get(_HEADER_CONTENT_ENCODING)
		So(strings.EqualFold(ce, "gzip"), ShouldBeTrue)

		So(before, ShouldBeTrue)
	})
}

type hijackableResponse struct {
	Hijacked bool
	header   http.Header
}

func newHijackableResponse() *hijackableResponse {
	return &hijackableResponse{header: make(http.Header)}
}

func (h *hijackableResponse) Header() http.Header           { return h.header }
func (h *hijackableResponse) Write(buf []byte) (int, error) { return 0, nil }
func (h *hijackableResponse) WriteHeader(code int)          {}
func (h *hijackableResponse) Flush()                        {}
func (h *hijackableResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.Hijacked = true
	return nil, nil, nil
}

func Test_ResponseWriter_Hijack(t *testing.T) {
	Convey("Hijack response", t, func() {
		hijackable := newHijackableResponse()

		m := macaron.New()
		m.Use(Gziper())
		m.Use(func(rw http.ResponseWriter) {
			hj, ok := rw.(http.Hijacker)
			So(ok, ShouldBeTrue)

			hj.Hijack()
		})

		r, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		r.Header.Set(_HEADER_ACCEPT_ENCODING, "gzip")
		m.ServeHTTP(hijackable, r)

		So(hijackable.Hijacked, ShouldBeTrue)
	})
}
