package http

import (
	"context"
	"fmt"
	"net/http"
)

const (
	SkipGlobalValidateResponseCode = "ghem.download.skip.GlobalValidateResponseCode"
)

var (
	GlobalBeforeEvalResponsePlugins = []BeforeEvalResponsePlugin{
		GlobalValidateResponseCode(),
	}
)

type BeforeEvalResponsePlugin interface {
	Plugin
	BeforeEvalResponse(context.Context, *http.Response, *http.Request) (context.Context, *http.Response, error)
}

type BeforeEvalResponseFunc func(context.Context, *http.Response, *http.Request) (context.Context, *http.Response, error)

func (instance BeforeEvalResponseFunc) BeforeEvalResponse(ctx context.Context, resp *http.Response, req *http.Request) (context.Context, *http.Response, error) {
	return instance(ctx, resp, req)
}

func (instance BeforeEvalResponseFunc) Self() Plugin {
	return instance
}

func GlobalValidateResponseCode() BeforeEvalResponsePlugin {
	return BeforeEvalResponseFunc(func(ctx context.Context, resp *http.Response, req *http.Request) (context.Context, *http.Response, error) {
		if v, ok := ctx.Value(SkipGlobalValidateResponseCode).(bool); ok && v {
			return ctx, resp, nil
		}
		actual := resp.StatusCode
		if actual < 100 || actual >= 400 {
			return ctx, resp, fmt.Errorf("status %d - %s", resp.StatusCode, resp.Status)
		}
		return ctx, resp, nil
	})
}
