package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/yousuketto/g-takeout-porter/internal/app"
	"github.com/yousuketto/g-takeout-porter/internal/infra"
	"os"
	"runtime"
	"strings"
)

func main() {
	var dryRun bool
	pflag.BoolVarP(&dryRun, "dry-run", "d", false, "Dry run mode")
	var concurrency int
	numberOfCPUs := runtime.NumCPU()
	pflag.IntVarP(&concurrency, "concurrency", "c", numberOfCPUs, "concurrency number")
	pflag.Parse()
	args := pflag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Error: accept 2 arguments, but received %d\nUsage: %s [flags] <source_directory> <dest_directory>\n", len(args), os.Args[0])
		os.Exit(1)
	}
	if concurrency <= 0 {
		fmt.Fprintf(os.Stderr, "Warning: concurrency option must be at least 1. Falling back to default value (number of CPUs: %d)\n", numberOfCPUs)
		concurrency = numberOfCPUs
	}
	porter := app.NewPorter(infra.NewTakeoutMetadataRepo(), infra.NewLocalStorage(concurrency))

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
