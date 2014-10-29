package gode

import (
	"os"
	"path/filepath"
	"runtime"
)

// The Node version to install.
// Override this by setting client.Version.
const DefaultNodeVersion = "v0.10.32"

// Client is the interface between Node and Go.
// It also setups up the Node environment if needed.
type Client struct {
	RootPath    string
	NodePath    string
	NpmPath     string
	ModulesPath string
	Version     string
	NodeURL     string
}

// NewClient creates a new Client at the specified rootPath
// The Node installation can then be setup here with client.Setup()
func NewClient(rootPath string) *Client {
	client := &Client{
		RootPath:    rootPath,
		NodePath:    filepath.Join(rootPath, nodeBase(DefaultNodeVersion), "bin", "node"),
		NpmPath:     filepath.Join(rootPath, nodeBase(DefaultNodeVersion), "bin", "npm"),
		ModulesPath: filepath.Join(rootPath, "lib", "node_modules"),
		Version:     DefaultNodeVersion,
		NodeURL:     nodeURL(DefaultNodeVersion),
	}
	os.Setenv("NODE_PATH", client.ModulesPath)
	os.Setenv("NPM_CONFIG_GLOBAL", "true")
	os.Setenv("NPM_CONFIG_PREFIX", client.RootPath)
	os.Setenv("NPM_CONFIG_SPIN", "false")

	return client
}

func nodeBase(version string) string {
	switch {
	case runtime.GOARCH == "386":
		return "node-" + version + "-" + runtime.GOOS + "-x86"
	default:
		return "node-" + version + "-" + runtime.GOOS + "-x64"
	}
}

func nodeURL(version string) string {
	switch {
	case runtime.GOOS == "windows" && runtime.GOARCH == "386":
		return "http://nodejs.org/dist/" + version + "/node.exe"
	case runtime.GOOS == "windows" && runtime.GOARCH == "amd64":
		return "http://nodejs.org/dist/" + version + "/x64/node.exe"
	case runtime.GOARCH == "386":
		return "http://nodejs.org/dist/" + version + "/" + nodeBase(version) + ".tar.gz"
	default:
		return "http://nodejs.org/dist/" + version + "/" + nodeBase(version) + ".tar.gz"
	}
}
