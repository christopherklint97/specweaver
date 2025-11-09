package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/christopherklint97/specweaver/pkg/generator"
	"github.com/christopherklint97/specweaver/pkg/parser"
)

const version = "0.1.0"

func main() {
	// Define flags
	specPath := flag.String("spec", "", "Path to OpenAPI specification file (required)")
	outputDir := flag.String("output", "./generated", "Output directory for generated code")
	packageName := flag.String("package", "api", "Package name for generated code")
	showVersion := flag.Bool("version", false, "Show version information")

	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("SpecWeaver version %s\n", version)
		os.Exit(0)
	}

	// Validate required flags
	if *specPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -spec flag is required\n\n")
		fmt.Fprintf(os.Stderr, "Usage: specweaver -spec <path> [options]\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse the OpenAPI specification
	p := parser.New()
	if err := p.ParseFile(*specPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing OpenAPI spec: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Loaded OpenAPI %s specification: %s\n", p.GetVersion(), p.GetSpec().Info.Title)

	// Generate code
	config := generator.Config{
		OutputDir:   *outputDir,
		PackageName: *packageName,
	}

	gen := generator.NewGenerator(p.GetSpec(), config)
	if err := gen.Generate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
