package main

import "testing"

// An environment is a named set of values; the active one wins over the
// project's own for the same name.

func TestEnvironmentVarsWinOverTheProjects(t *testing.T) {
	app := newTestApp(t)

	// No environment yet: the project's own values are what a request uses.
	if got := lookupValue(app.Project().Variables(), "baseUrl"); got != "https://api.example.com" {
		t.Fatalf("the project's baseUrl is %q", got)
	}

	staging := app.AddEnvironment("Staging")
	if staging == nil {
		t.Fatal("the environment was not created")
	}
	if got := app.environmentName(); got != "Staging" {
		t.Errorf("the title bar shows %q", got)
	}

	vars := app.EditVarsTarget()
	if vars == nil {
		t.Fatal("there is nothing to edit")
	}
	*vars = append(*vars, Pair{Key: "baseUrl", Value: "https://staging.example.com"})

	if got := lookupValue(app.Project().Variables(), "baseUrl"); got != "https://staging.example.com" {
		t.Errorf("the active environment did not win: %q", got)
	}

	// The project's other values still resolve underneath it.
	if got := lookupValue(app.Project().Variables(), "token"); got != "" {
		t.Errorf("the project's token is %q", got)
	}

	// And it survives a reload.
	reloaded, err := NewApp(app.dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.environmentName(); got != "Staging" {
		t.Errorf("after reloading the active environment is %q", got)
	}
}

func TestNoEnvironmentFallsBackToTheProject(t *testing.T) {
	app := newTestApp(t)

	app.AddEnvironment("Production")
	app.UseEnvironment("")

	if got := app.environmentName(); got != "No environment" {
		t.Errorf("the title bar shows %q", got)
	}
	if got := lookupValue(app.Project().Variables(), "baseUrl"); got != "https://api.example.com" {
		t.Errorf("with no environment the baseUrl is %q", got)
	}
}

func lookupValue(vars []Pair, name string) string {
	value, _ := lookup(vars, name)
	return value
}
