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

	"github.com/heynemann/level/channel"
	. "github.com/onsi/gomega"

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
	handler := channel.WebApp.NoListen().Handler

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL: "http://example.com",
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: &GinkgoReporter{},
		Printers: []httpexpect.Printer{
		//httpexpect.NewDebugPrinter(&GinkgoPrinter{}, true),
		},
	})

	return e.Request(method, url)
}
