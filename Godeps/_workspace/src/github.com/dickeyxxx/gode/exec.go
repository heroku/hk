package gode

import "os/exec"

// RunScript runs a given script in node
// Returns an *os/exec.Cmd instance
func (c *Client) RunScript(script string) *exec.Cmd {
	return exec.Command(c.NodePath, "-e", script)
}
