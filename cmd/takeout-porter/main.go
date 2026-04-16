package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/yousuketto/g-takeout-porter/internal/app"
	"github.com/yousuketto/g-takeout-porter/internal/infra"
	"os"
	"strings"
)

func main() {
	var dryRun bool
	pflag.BoolVarP(&dryRun, "dry-run", "d", false, "Dry run mode")
	pflag.Parse()
	args := pflag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Error: accept 2 arguments, but received %d\nUsage: %s [flags] <source_directory> <dest_directory>\n", len(args), os.Args[0])
		os.Exit(1)
	}
	porter := app.NewPorter(infra.NewTakeoutMetadataRepo(), infra.NewLocalStorage())

	if dryRun {
		results, err := porter.DryRun(args[0], args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "dry run error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Dry running is success.")
		fmt.Println(strings.Join(results, "\n"))
		return
	}

	err := porter.Run(args[0], args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
