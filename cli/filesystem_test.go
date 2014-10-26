package cli

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHomeDir(t *testing.T) {
	Convey("it isn't blank", t, func() {
		So(homeDir(), ShouldNotBeBlank)
	})
}
