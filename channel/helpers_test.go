// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package channel_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/heynemann/level/channel"
	. "github.com/onsi/gomega"
	"github.com/valyala/fasthttp"

	"github.com/gavv/httpexpect"
	"github.com/uber-go/zap"
)

// GetDefaultTestApp returns a new Santiago API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp(logger zap.Logger) (*channel.Channel, error) {
	options := channel.DefaultOptions()
	return channel.New(options, logger)
}

// Get returns a test request against specified URL
func Get(channel *channel.Channel, url string) *httpexpect.Response {
	req := sendRequest(channel, "GET", url)
	return req.Expect()
}

// PostBody returns a test request against specified URL
func PostBody(channel *channel.Channel, url string, payload string) *httpexpect.Response {
	return sendBody(channel, "POST", url, payload)
}

// PutBody returns a test request against specified URL
func PutBody(channel *channel.Channel, url string, payload string) *httpexpect.Response {
	return sendBody(channel, "PUT", url, payload)
}

func sendBody(channel *channel.Channel, method string, url string, payload string) *httpexpect.Response {
	req := sendRequest(channel, method, url)
	return req.WithBytes([]byte(payload)).Expect()
}

// PostJSON returns a test request against specified URL
func PostJSON(channel *channel.Channel, url string, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(channel, "POST", url, payload)
}

// PutJSON returns a test request against specified URL
func PutJSON(channel *channel.Channel, url string, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(channel, "PUT", url, payload)
}

func sendJSON(channel *channel.Channel, method, url string, payload map[string]interface{}) *httpexpect.Response {
	req := sendRequest(channel, method, url)
	return req.WithJSON(payload).Expect()
}

// Delete returns a test request against specified URL
func Delete(channel *channel.Channel, url string) *httpexpect.Response {
	req := sendRequest(channel, "DELETE", url)
	return req.Expect()
}

//GinkgoReporter implements tests for httpexpect
type GinkgoReporter struct {
}

// Errorf implements Reporter.Errorf.
func (g *GinkgoReporter) Errorf(message string, args ...interface{}) {
	Expect(false).To(BeTrue(), fmt.Sprintf(message, args...))
}

//GinkgoPrinter reports errors to stdout
type GinkgoPrinter struct{}

//Logf reports to stdout
func (g *GinkgoPrinter) Logf(source string, args ...interface{}) {
	fmt.Printf(source, args...)
}

func sendRequest(channel *channel.Channel, method, url string) *httpexpect.Request {
	api := channel.WebApp
	srv := api.Servers.Main()

	if srv == nil { // maybe the user called this after .Listen/ListenTLS/ListenUNIX, the tester can be used as standalone (with no running iris instance) or inside a running instance/app
		srv = api.ListenVirtual(api.Config.Tester.ListeningAddr)
	}

	opened := api.Servers.GetAllOpened()
	h := srv.Handler
	baseURL := srv.FullHost()
	if len(opened) > 1 {
		baseURL = ""
		//we have more than one server, so we will create a handler here and redirect by registered listening addresses
		h = func(reqCtx *fasthttp.RequestCtx) {
			for _, s := range opened {
				if strings.HasPrefix(reqCtx.URI().String(), s.FullHost()) { // yes on :80 should be passed :80 also, this is inneed for multiserver testing
					s.Handler(reqCtx)
					break
				}
			}
		}
	}

	if api.Config.Tester.ExplicitURL {
		baseURL = ""
	}

	testConfiguration := httpexpect.Config{
		BaseURL: baseURL,
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(h),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: &GinkgoReporter{},
	}
	if api.Config.Tester.Debug {
		testConfiguration.Printers = []httpexpect.Printer{
			httpexpect.NewDebugPrinter(&GinkgoPrinter{}, true),
		}
	}

	return httpexpect.WithConfig(testConfiguration).Request(method, url)
}
