package config

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

const (
	indent = "    " // 4 spaces
)

// Display is a temporary struct to display a config object to the console
type Display struct {
	topLevel string
	writer   io.Writer
	entries  []Entry
}

func newDisplay(name string, w io.Writer) *Display {
	return &Display{
		topLevel: name,
		writer:   w,
		entries:  make([]Entry, 0, 60),
	}
}

func (d *Display) addEntry(stack []string, key string, values []string) {
	entry := Entry{
		section: strings.Join(stack, "."), // in theory we should only have zero or one level, but don't fail if we get more
		key:     key,
		values:  values,
	}
	d.entries = append(d.entries, entry)
}

func (d *Display) addKeyOnlyEntry(stack []string, key string) {
	d.addEntry(stack, key, nil)
	d.entries[len(d.entries)-1].keyOnly = true
}

func (d *Display) Flush() {
	tabWriter := tabwriter.NewWriter(d.writer, 0, 2, 2, ' ', 0)
	// title
	fmt.Fprintf(tabWriter, "%s:\n", d.topLevel)
	section := ""
	prefix := indent
	for _, entry := range d.entries {
		if section != entry.section {
			// new section
			section = entry.section
			// reset indentation
			prefix = indent
			if section != "" {
				// sub-section
				fmt.Fprintf(tabWriter, "\n%s%s:", prefix, section)
				prefix += indent
			}
			fmt.Fprintln(tabWriter, "")
		}
		if entry.key == "" {
			continue
		}
		if entry.keyOnly {
			fmt.Fprintf(tabWriter, "%s%s\n", prefix, entry.key)
			continue
		}
		if len(entry.values) > 0 {
			fmt.Fprintf(tabWriter, "%s%s:\t%s\n", prefix, entry.key, entry.values[0])
		}
		if len(entry.values) > 1 {
			for i := 1; i < len(entry.values); i++ {
				fmt.Fprintf(tabWriter, "%s\t%s\n", prefix, entry.values[i])
			}
		}
	}
	tabWriter.Flush()
}

// Entry of configuration to display to the console
type Entry struct {
	section string
	key     string
	keyOnly bool
	values  []string
}
