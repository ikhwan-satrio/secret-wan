package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ikhwan-satrio/secret-wan/internal/app"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
)

// fail prints a styled error to stderr and terminates the process.
func fail(shortMsg, detailMsg string) {
	fmt.Fprintln(os.Stderr, ui.FormatError(shortMsg, detailMsg))
	os.Exit(1)
}

// failErr is a convenience wrapper that uses err.Error() as the detail line.
func failErr(shortMsg string, err error) {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	fail(shortMsg, detail)
}

// failApp prefers the service error's own heading and keeps the default heading
// for plain infrastructure errors.
func failApp(defaultShort string, err error) {
	var userErr *app.UserError
	if errors.As(err, &userErr) {
		fail(userErr.Short, userErr.Detail)
	}

	failErr(defaultShort, err)
}

// warn writes a styled warning to stderr.
func warn(message string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning, ui.Warning.Render(message))
}

func printHeader(action string) {
	fmt.Println(ui.Title.Render("secret-wan") + ui.Muted.Render("  •  "+action))
	fmt.Println(ui.Divider(44))
}
