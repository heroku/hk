package gode

import (
	"os"
	"testing"
)

func TestPackages(t *testing.T) {
	c := setup()
	must(os.RemoveAll(c.ModulesPath))
	must(c.InstallPackage("request"))
	packages, err := c.Packages()
	must(err)
	for _, pkg := range packages {
		if pkg.Name == "request" {
			return
		}
	}
	t.Fatalf("package did not install")
}

func TestPackagesGithubPackage(t *testing.T) {
	c := setup()
	must(os.RemoveAll(c.ModulesPath))
	must(c.InstallPackage("dickeyxxx/heroku-production-check"))
	packages, err := c.Packages()
	must(err)
	for _, pkg := range packages {
		if pkg.Name == "heroku-production-check" {
			return
		}
	}
	t.Fatalf("package did not install")
}
