package main

import (
	"fmt"
	"log"
	"os"

	"github.com/christopherklint97/specweaver"
)

func main() {
	// Example 1: Simple usage with Generate function
	fmt.Println("=== Example 1: Simple Generation ===")
	err := simpleGenerate()
	if err != nil {
		log.Fatalf("Simple generation failed: %v", err)
	}

	// Example 2: Advanced usage with separate parser and generator
	fmt.Println("\n=== Example 2: Advanced Usage ===")
	err = advancedGenerate()
	if err != nil {
		log.Fatalf("Advanced generation failed: %v", err)
	}

	fmt.Println("\n✓ All examples completed successfully!")
}

// simpleGenerate demonstrates the simplest way to use SpecWeaver as a library
func simpleGenerate() error {
	// Define the path to your OpenAPI spec
	specPath := "../petstore.yaml"

	// Generate code with a single function call
	err := specweaver.Generate(specPath, specweaver.Options{
		OutputDir:   "./generated-simple",
		PackageName: "petstore",
	})
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Println("✓ Generated code in ./generated-simple/")
	return nil
}

// advancedGenerate demonstrates using the parser and generator separately
// This approach gives you more control and access to the parsed spec
func advancedGenerate() error {
	// Step 1: Create a parser and parse the OpenAPI spec
	parser := specweaver.NewParser()
	err := parser.ParseFile("../petstore.yaml")
	if err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}

	// You can access the parsed specification
	spec := parser.GetSpec()
	fmt.Printf("✓ Parsed OpenAPI %s spec: %s\n", parser.GetVersion(), spec.Info.Title)
	fmt.Printf("  Version: %s\n", spec.Info.Version)
	fmt.Printf("  Description: %s\n", spec.Info.Description)
	fmt.Printf("  Paths: %d\n", len(spec.Paths))
	fmt.Printf("  Schemas: %d\n", len(spec.Components.Schemas))

	// Step 2: Create a generator with custom options
	generator := specweaver.NewGenerator(spec, specweaver.Options{
		OutputDir:   "./generated-advanced",
		PackageName: "api",
	})

	// Step 3: Generate the code
	err = generator.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	fmt.Println("✓ Generated code in ./generated-advanced/")

	// Optional: Clean up the generated directories (for example purposes)
	// In real usage, you would keep the generated code
	defer func() {
		os.RemoveAll("./generated-simple")
		os.RemoveAll("./generated-advanced")
	}()

	return nil
}
