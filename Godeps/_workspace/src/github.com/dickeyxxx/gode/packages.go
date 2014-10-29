package gode

import (
	"encoding/json"
	"os"
	"os/exec"
)

// Package represents an npm package.
type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Packages returns a list of npm packages installed.
func (c *Client) Packages() ([]Package, error) {
	cmd := c.execNpm("list", "--json", "--depth=0")
	var response map[string]map[string]Package
	output, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(output).Decode(&response)
	if err != nil {
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}
	var packages []Package
	for name, p := range response["dependencies"] {
		p.Name = name
		packages = append(packages, p)
	}
	return packages, nil
}

// InstallPackage installs an npm package.
func (c *Client) InstallPackage(name string) error {
	cmd := c.execNpm("install", name)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (c *Client) execNpm(args ...string) *exec.Cmd {
	cmd := exec.Command(c.NpmPath, args...)
	cmd.Stderr = os.Stderr
	return cmd
}
