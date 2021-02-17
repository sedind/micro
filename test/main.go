package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/sedind/micro"
)

func DoSomething(action string) micro.MiddlewareHandlerFunc {
	return func(next micro.MiddlewareFunc) micro.MiddlewareFunc {
		return func(c *micro.Context) error {
			// do something before calling the next handler
			c.Logger.Infof("Do something %s befor next handler", action)
			err := next(c)
			// do something after call the handler
			c.Logger.Infof("Do something %s after next handler", action)
			return err
		}
	}
}

func main() {
	app := micro.New()

	app.Use(DoSomething("1"))
	app.Use(DoSomething("2"))
	app.Use(DoSomething("3"))

	app.GET("/", func(ctx *micro.Context) micro.ActionResult {
		return micro.JSONResult(
			http.StatusOK,
			micro.VM{
				"HEllo": "world",
			})
	})

	app.GET("/yaml", func(ctx *micro.Context) micro.ActionResult {
		return micro.YAMLResult(
			http.StatusOK,
			micro.VM{
				"HEllo": "world",
			})
	})

	app.GET("/xml", func(ctx *micro.Context) micro.ActionResult {
		type Note struct {
			To      string
			From    string
			Heading string
			Body    string
		}
		return micro.XMLResult(
			http.StatusOK,
			&Note{
				To:      "John",
				From:    "Jane",
				Heading: "Reminder",
				Body:    "Don't forget me this weekend!",
			})
	})

	app.GET("/text", func(ctx *micro.Context) micro.ActionResult {
		return micro.TextResult(http.StatusOK, "Hello World")
	})

	app.GET("/data", func(ctx *micro.Context) micro.ActionResult {
		return micro.DataResult(http.StatusOK, []byte("Hello World"), []string{"text/plain; charset=utf-8"})
	})

	app.GET("/err", func(ctx *micro.Context) micro.ActionResult {
		return micro.ErrorResult(http.StatusBadRequest, errors.New("some bad status error"))
	})

	app.GET("/download", func(ctx *micro.Context) micro.ActionResult {
		r := strings.NewReader("test.csv this is test file content,something,a,b,c,d,e,f,g,h,i,j,k,l")
		return micro.DownloadResult("test.txt", r)
	})

	app.GET("/file", func(ctx *micro.Context) micro.ActionResult {
		return micro.FileResult("go.mod")
	})

	if err := app.Serve(); err != nil {
		app.Logger.Error(err)
	}
}
