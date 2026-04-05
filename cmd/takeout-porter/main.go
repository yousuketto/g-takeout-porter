package main

import (
	"fmt"
	"github.com/yousuketto/g-takeout-porter/internal/app"
	"github.com/yousuketto/g-takeout-porter/internal/infra"
	"os"
)

func main() {
	porter := app.NewPorter(infra.NewTakeoutMetadataRepo(), infra.NewLocalStorage())
	err := porter.Run(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
