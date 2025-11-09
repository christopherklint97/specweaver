module github.com/christopherklint97/specweaver/examples/custom-router

go 1.24.7

require (
	github.com/christopherklint97/specweaver v0.0.0
	github.com/christopherklint97/specweaver/examples/server v0.0.0-20251109080721-36afc05a775a
	github.com/go-chi/chi/v5 v5.2.0
)

replace github.com/christopherklint97/specweaver => ../..
replace github.com/christopherklint97/specweaver/examples/server => ../server
