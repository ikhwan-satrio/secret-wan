package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/ikhwan-satrio/secret-wan/internal/app"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
)

// Fail prints a styled error to stderr and terminates the process.
func Fail(shortMsg, detailMsg string) {
	fmt.Fprintln(os.Stderr, ui.FormatError(shortMsg, detailMsg))
	os.Exit(1)
}

// FailErr is a convenience wrapper that uses err.Error() as the detail line.
func FailErr(shortMsg string, err error) {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	Fail(shortMsg, detail)
}

// FailApp prefers the service error's own heading and keeps the default heading
// for plain infrastructure errors.
func FailApp(defaultShort string, err error) {
	var userErr *app.UserError
	if errors.As(err, &userErr) {
		Fail(userErr.Short, userErr.Detail)
	}

	FailErr(defaultShort, err)
}

// Warn writes a styled warning to stderr.
func Warn(message string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning, ui.Warning.Render(message))
}

// PrintHeader renders the shared command header.
func PrintHeader(action string) {
	fmt.Println(ui.Title.Render("secret-wan") + ui.Muted.Render("  •  "+action))
	fmt.Println(ui.Divider(44))
}
