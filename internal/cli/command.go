package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// usageError reports invalid command line usage (exit code 2).
type usageError struct {
	cmd *command
	msg string
}

func (e *usageError) Error() string { return e.msg }

// command is a node of the command tree. A command either has subcommands
// or a run function.
type command struct {
	name     string
	args     string
	summary  string
	flags    *flag.FlagSet
	minArgs  int
	maxArgs  int
	run      func(ctx context.Context, args []string) error
	children []*command
	parent   *command
}

func (c *command) add(children ...*command) *command {
	for _, child := range children {
		child.parent = c
		c.children = append(c.children, child)
	}
	return c
}

func (c *command) path() string {
	if c.parent == nil {
		return c.name
	}
	return c.parent.path() + " " + c.name
}

func (c *command) execute(ctx context.Context, args []string, stderr io.Writer) error {
	fs := c.flags
	if fs == nil {
		fs = flag.NewFlagSet(c.name, flag.ContinueOnError)
	}
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			c.printUsage(stderr)
			return err
		}
		return &usageError{cmd: c, msg: err.Error()}
	}
	positional := fs.Args()

	if len(c.children) > 0 {
		if len(positional) == 0 {
			return &usageError{cmd: c, msg: "missing command"}
		}
		if positional[0] == "help" {
			c.printUsage(stderr)
			return flag.ErrHelp
		}
		for _, child := range c.children {
			if child.name == positional[0] {
				return child.execute(ctx, positional[1:], stderr)
			}
		}
		return &usageError{cmd: c, msg: fmt.Sprintf("unknown command %q", positional[0])}
	}

	if len(positional) < c.minArgs || len(positional) > c.maxArgs {
		return &usageError{cmd: c, msg: "wrong number of arguments"}
	}
	return c.run(ctx, positional)
}

func (c *command) printUsage(w io.Writer) {
	usage := c.path()
	if c.flags != nil {
		usage += " [flags]"
	}
	if len(c.children) > 0 {
		usage += " <command>"
	}
	if c.args != "" {
		usage += " " + c.args
	}
	fmt.Fprintf(w, "Usage: %s\n", usage)
	if c.summary != "" {
		fmt.Fprintf(w, "\n%s\n", c.summary)
	}
	if len(c.children) > 0 {
		fmt.Fprintln(w, "\nCommands:")
		for _, child := range c.children {
			fmt.Fprintf(w, "  %-28s %s\n", strings.TrimSpace(child.name+" "+child.args), child.summary)
		}
	}
	if c.flags != nil {
		fmt.Fprintln(w, "\nFlags:")
		c.flags.SetOutput(w)
		c.flags.PrintDefaults()
		c.flags.SetOutput(io.Discard)
	}
}
