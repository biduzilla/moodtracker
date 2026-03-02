package main

import "gateway/internal/api"

func main() {
	app := api.NewApp()
	err := app.Server()
	if err != nil {
		app.Logger.PrintError(err, nil)
	}
}
