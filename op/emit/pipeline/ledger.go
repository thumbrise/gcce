package pipeline

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
)

type PrettyDiffLedger struct {
	out io.Writer
}

func NewPrettyDiffLedger(out io.Writer) *PrettyDiffLedger {
	return &PrettyDiffLedger{
		out: out,
	}
}

func (l *PrettyDiffLedger) Log(path, reason string, prev, next any) {
	var b strings.Builder
	b.WriteString(time.Now().Format("15:04:05.000 "))
	_, _ = fmt.Fprintf(&b, "[LEDGER] %s\n", path)
	_, _ = fmt.Fprintf(&b, "  Reason: %s\n", reason)
	b.WriteString("  Delta:\n")

	diff := compare(prev, next)
	if diff == "" {
		diff = "(no changes)"
	}

	for _, line := range strings.Split(diff, "\n") {
		b.WriteString("    " + line + "\n")
	}

	_, _ = io.WriteString(l.out, b.String())
}

func compare(prev, next any) string {
	diff := cmp.Diff(prev, next)
	if diff == "" {
		return ""
	}

	var out strings.Builder

	for _, line := range strings.Split(diff, "\n") {
		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case '+':
			out.WriteString(color.GreenString(line) + "\n")
		case '-':
			out.WriteString(color.RedString(line) + "\n")
		default:
			out.WriteString(color.WhiteString(line) + "\n")
		}
	}

	return out.String()
}
