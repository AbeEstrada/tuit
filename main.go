package main

import (
	"log"

	"github.com/AbeEstrada/tuit/tui"
)

func main() {
	app, err := tui.CreateApp()
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}
	defer app.Close() // Ensure the terminal state is restored

	if err := app.Run(); err != nil {
		log.Fatalf("app exited with error: %v", err)
	}
}
