package main

import (
	"fmt"
	"os"

	"waygent/internal/agent"
	"waygent/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "waygent: starting task: %s\n", cfg.Task)
		fmt.Fprintf(os.Stderr, "  model: %s\n", cfg.Model)
		fmt.Fprintf(os.Stderr, "  api: %s\n", cfg.APIURL)
	}

	a := agent.New(cfg)
	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
