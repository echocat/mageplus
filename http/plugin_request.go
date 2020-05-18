package http

import (
	"context"
	"net/http"
)

var (
	GlobalBeforeRequestPlugins []BeforeRequestPlugin
)

type BeforeRequestPlugin interface {
	Plugin
	BeforeRequest(context.Context, *http.Request) (context.Context, *http.Request, error)
}

func Method(method string) BeforeRequestPlugin {
	return BeforeRequestFunc(func(ctx context.Context, req *http.Request) (context.Context, *http.Request, error) {
		req.Method = method
		return ctx, req, nil
	})
}

func BearerAuth(token string) BeforeRequestPlugin {
	return Header("Authorization", "Bearer: "+token)
}

func BasicAuth(user, password string) BeforeRequestPlugin {
	return BeforeRequestFunc(func(ctx context.Context, req *http.Request) (context.Context, *http.Request, error) {
		req.SetBasicAuth(user, password)
		return ctx, req, nil
	})
}

func Header(name, value string) BeforeRequestPlugin {
	return BeforeRequestFunc(func(ctx context.Context, req *http.Request) (context.Context, *http.Request, error) {
		req.Header.Set(name, value)
		return ctx, req, nil
	})
}

type BeforeRequestFunc func(context.Context, *http.Request) (context.Context, *http.Request, error)

func (instance BeforeRequestFunc) BeforeRequest(ctx context.Context, req *http.Request) (context.Context, *http.Request, error) {
	return instance(ctx, req)
}

func (instance BeforeRequestFunc) Self() Plugin {
	return instance
}
