package http

import (
	"context"
	"fmt"
	"github.com/echocat/mageplus/io"
	"net/http"
)

var (
	Client = http.DefaultClient
)

func Execute(url string, plugins ...Plugin) error {
	ctx := context.Background()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("cannot create request for '%s': %v", url, err)
	}

	if plugins != nil {
		for _, plugin := range plugins {
			if i, ok := plugin.(BeforeRequestPlugin); ok {
				var err error
				if ctx, req, err = i.BeforeRequest(ctx, req); err != nil {
					return fmt.Errorf("cannot prepare request for '%s': %v", url, err)
				}
			}
		}
	}

	for _, plugin := range GlobalBeforeRequestPlugins {
		var err error
		if ctx, req, err = plugin.BeforeRequest(ctx, req); err != nil {
			return fmt.Errorf("cannot prepare request for '%s': %v", url, err)
		}
	}

	resp, err := Client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute reqest of '%s': %v", url, err)
	}
	if resp.Body != nil {
		defer io.CloseQuietly(resp.Body)
	}

	if plugins != nil {
		for _, plugin := range plugins {
			if i, ok := plugin.(BeforeEvalResponsePlugin); ok {
				var err error
				if ctx, resp, err = i.BeforeEvalResponse(ctx, resp, req); err != nil {
					return fmt.Errorf("cannot prepare response of '%s': %v", url, err)
				}
			}
		}
	}
	for _, plugin := range GlobalBeforeEvalResponsePlugins {
		var err error
		if ctx, resp, err = plugin.BeforeEvalResponse(ctx, resp, req); err != nil {
			return fmt.Errorf("cannot prepare response of '%s': %v", url, err)
		}
	}

	if plugins != nil {
		for _, plugin := range plugins {
			if i, ok := plugin.(EvalResponsePlugin); ok {
				if err := i.EvalResponse(ctx, resp, req); err != nil {
					return fmt.Errorf("cannot evaluate response of '%s': %v", url, err)
				}
			}
		}
	}
	for _, plugin := range GlobalEvalResponsePlugins {
		if err := plugin.EvalResponse(ctx, resp, req); err != nil {
			return fmt.Errorf("cannot evaluate response of '%s': %v", url, err)
		}
	}

	return nil
}
