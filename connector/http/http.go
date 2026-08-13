package http

import (
	"net/http"
	"time"

	dge "github.com/wiryax/dage/core"
)

type HTTPConnector struct {
	client  *http.Client
	timeout time.Duration
}

func (f *HTTPConnector) Validate(_ *dge.GraphContext) error {
	return nil
}
func (f *HTTPConnector) Acquire(_ any) any {
	client := http.Client{
		Transport: nil,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			panic("TODO")
		},
		Jar:     nil,
		Timeout: f.timeout,
	}

	return client
}
func (f *HTTPConnector) Release() error {
	return nil
}
