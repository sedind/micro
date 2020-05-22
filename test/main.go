package main

import (
	"errors"
	"net/http"

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

	if err := app.Serve(); err != nil {
		app.Logger.Error(err)
	}
}
