package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/benjamonnguyen/opendoorchat/frontend/components"
)

func main() {
	start := time.Now()
	ctx := context.Background()

	// app.html
	f, err := os.Create("frontend/public/app.html")
	handle(err)
	defer f.Close()
	err = components.App().Render(ctx, f)
	handle(err)

	// new-chat.html
	f, err = os.Create("frontend/public/new-chat.html")
	handle(err)
	defer f.Close()
	err = components.NewChat().Render(ctx, f)
	handle(err)

	//
	fmt.Printf("generated HTML files in %s\n", time.Since(start).Truncate(time.Millisecond))
}

func handle(e error) {
	if e != nil {
		panic(e)
	}
}
