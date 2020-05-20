package main

import (
	"github.com/sedind/micro"
)

func main() {
	app := micro.New()

	app.GET("/", func(ctx *micro.Context) (micro.ActionResult, error) {
		return micro.RenderJSON(micro.VM{
			"HEllo": "world",
		}), nil
	})

	if err := app.Serve(); err != nil {
		app.Logger.Error(err)
	}
}
