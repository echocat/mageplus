package http

import (
	"context"
	"errors"
	"fmt"
	sio "github.com/echocat/mageplus/io"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	GlobalEvalResponsePlugins []EvalResponsePlugin

	ErrNoBody = errors.New("no body")
)

type EvalResponsePlugin interface {
	Plugin
	EvalResponse(context.Context, *http.Response, *http.Request) error
}

type EvalResponseFunc func(context.Context, *http.Response, *http.Request) error

func (instance EvalResponseFunc) EvalResponse(ctx context.Context, resp *http.Response, req *http.Request) error {
	return instance(ctx, resp, req)
}

func (instance EvalResponseFunc) Self() Plugin {
	return instance
}

func writeTo(writer io.Writer, reader io.Reader) error {
	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("cannot copy downloaded content: %v", err)
	}
	return nil
}

func WriteTo(writer io.Writer) EvalResponseFunc {
	return EvalBody(func(reader io.Reader) error {
		return writeTo(writer, reader)
	})
}

func WriteToFile(filename string, perm os.FileMode) EvalResponseFunc {
	return EvalBody(func(reader io.Reader) error {
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
		if err != nil {
			return fmt.Errorf("cannot open target file '%s': %v", filename, err)
		}
		defer sio.CloseQuietly(f)
		return writeTo(f, reader)
	})
}

type OnTempFile func(*os.File) error

func WriteToTemporaryFile(dir, pattern string, onTempFile OnTempFile) EvalResponseFunc {
	return EvalBody(func(reader io.Reader) error {
		f, err := ioutil.TempFile(dir, pattern)
		if err != nil {
			return fmt.Errorf("cannot create target file '%s/%s': %v", dir, pattern, err)
		}
		//noinspection GoUnhandledErrorResult
		defer os.Remove(f.Name())
		defer sio.CloseQuietly(f)
		if err := writeTo(f, reader); err != nil {
			return err
		}
		return onTempFile(f)
	})
}

type OnBody func(io.Reader) error

func EvalBody(onBody OnBody) EvalResponseFunc {
	return func(_ context.Context, resp *http.Response, _ *http.Request) error {
		if resp.Body == nil {
			return ErrNoBody
		}
		return onBody(resp.Body)
	}
}
