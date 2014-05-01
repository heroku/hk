// WARNING: This code is auto-generated from the Heroku Platform API JSON Schema
// by a Ruby script (gen/gen.rb). Changes should be made to the generation
// script rather than the generated files.

package heroku

// A build result contains the output from a build.
type BuildResult struct {
	// identity of build
	Build struct {
		Id     string `json:"id"`
		Status string `json:"status"`
	} `json:"build"`

	// status from the build
	ExitCode int `json:"exit_code"`

	// a single line of output to STDOUT or STDERR from the build.
	Lines []struct {
		Stream string `json:"stream"`
		Line   string `json:"line"`
	} `json:"lines"`
}

// Info for existing result.
//
// appIdentity is the unique identifier of the BuildResult's App. buildIdentity
// is the unique identifier of the BuildResult's Build.
func (c *Client) BuildResultInfo(appIdentity string, buildIdentity string) (*BuildResult, error) {
	var buildResult BuildResult
	return &buildResult, c.Get(&buildResult, "/api/apps/"+appIdentity+"/builds/"+buildIdentity+"/result")
}
