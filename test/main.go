package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/sedind/micro"
)

func main() {
	app := micro.New()

	app.GET("/", func(ctx *micro.Context) micro.ActionResult {
		return micro.RenderJSON(
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
		return micro.RenderXML(
			http.StatusOK,
			&Note{
				To:      "John",
				From:    "Jane",
				Heading: "Reminder",
				Body:    "Don't forget me this weekend!",
			})
	})

	app.GET("/text", func(ctx *micro.Context) micro.ActionResult {
		return micro.RenderText(http.StatusOK, "Hello World")
	})

	app.GET("/data", func(ctx *micro.Context) micro.ActionResult {
		return micro.RenderData(http.StatusOK, []byte("Hello World"), []string{"text/plain; charset=utf-8"})
	})

	app.GET("/err", func(ctx *micro.Context) micro.ActionResult {
		return micro.ErrorResult(http.StatusBadRequest, errors.New("some bad status error"))
	})

	app.GET("/file", func(ctx *micro.Context) micro.ActionResult {
		r := strings.NewReader("test.csv this is test file content,something,a,b,c,d,e,f,g,h,i,j,k,l")
		return micro.FileResult("test.txt", r)
	})

	if err := app.Serve(); err != nil {
		app.Logger.Error(err)
	}
}
