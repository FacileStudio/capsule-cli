package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

var (
	cyan   = color.New(color.FgCyan)
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
	faint  = color.New(color.Faint)
)

// DisableColor disables colored output.
func DisableColor() {
	color.NoColor = true
}

// Step prints an informational step message.
func Step(format string, a ...any) {
	fmt.Fprintf(os.Stdout, "%s %s\n", cyan.Sprint("▸"), fmt.Sprintf(format, a...))
}

// Success prints a success message.
func Success(format string, a ...any) {
	fmt.Fprintf(os.Stdout, "%s %s\n", green.Sprint("✓"), fmt.Sprintf(format, a...))
}

// Warn prints a warning message to stderr.
func Warn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "%s %s\n", yellow.Sprint("!"), fmt.Sprintf(format, a...))
}

// Error prints an error message to stderr.
func Error(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "%s %s\n", red.Sprint("✗"), fmt.Sprintf(format, a...))
}

// Hint prints a hint message to stdout.
func Hint(format string, a ...any) {
	fmt.Fprintf(os.Stdout, "  %s\n", faint.Sprint(fmt.Sprintf(format, a...)))
}

// Dim returns a string formatted with faint coloring.
func Dim(s string) string {
	return faint.Sprint(s)
}

// Table writes aligned columns to stdout.
func Table(header []string, rows [][]string) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, faint.Sprint(strings.Join(header, "\t")))
	for _, row := range rows {
		fmt.Fprintln(writer, strings.Join(row, "\t"))
	}
	writer.Flush()
}

// JSON encodes and prints a value as indented JSON.
func JSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

// Plain prints a formatted line to stdout.
func Plain(format string, a ...any) {
	fmt.Fprintf(os.Stdout, format+"\n", a...)
}
