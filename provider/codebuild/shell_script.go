package codebuild

import (
	"bytes"
	"fmt"
	"io"

	"github.com/piotrkubisa/localcb/ci"
)

// ShellScript transforms CodeBuild definition file to localcb.sh shell script
type ShellScript struct {
	Buffer *bytes.Buffer
}

// NewShellScript creates new ShellScript with a brand new buffer
func NewShellScript() *ShellScript {
	return &ShellScript{
		Buffer: new(bytes.Buffer),
	}
}

func (sr *ShellScript) ExtractStage(stage *ci.Stage) error {
	sr.enterStage(sr.Buffer, stage)

	for _, cmd := range stage.Commands {
		io.WriteString(sr.Buffer, cmd.Exec)
		io.WriteString(sr.Buffer, "\n")
	}

	sr.exitStage(sr.Buffer, stage)
	return nil
}

func (sr *ShellScript) enterStage(w io.Writer, s *ci.Stage) {
	io.WriteString(w, "# ********************************\n")
	fmt.Fprintf(w, "# > Entering '%s' stage\n", s.Name)
	fmt.Fprintf(w, "# * Found %d command(s)\n", len(s.Commands))
	if len(s.Commands) > 0 {
		io.WriteString(w, "# ---\n")
	}
}

func (sr *ShellScript) exitStage(w io.Writer, s *ci.Stage) {
	if len(s.Commands) > 0 {
		io.WriteString(w, "# ---\n")
	}
	fmt.Fprintf(w, "# < Exiting '%s' stage\n", s.Name)
	io.WriteString(w, "# ********************************\n")
	io.WriteString(w, "\n")
}
