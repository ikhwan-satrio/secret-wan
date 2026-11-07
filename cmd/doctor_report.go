package cmd

import (
	"fmt"
	"os"

	"github.com/ikhwan-satrio/secret-wan/internal/ui"
)

type checkLevel int

const (
	checkOK checkLevel = iota
	checkWarn
	checkFail
)

type doctorCheck struct {
	level  checkLevel
	label  string
	detail string
}

type doctorSection struct {
	title  string
	checks []doctorCheck
}

type doctorReport struct {
	sections []*doctorSection
	fails    int
	warns    int
}

func (r *doctorReport) section(title string) *doctorSection {
	section := &doctorSection{title: title}
	r.sections = append(r.sections, section)
	return section
}

func (r *doctorReport) add(section *doctorSection, level checkLevel, label, detail string) {
	section.checks = append(section.checks, doctorCheck{level: level, label: label, detail: detail})

	switch level {
	case checkWarn:
		r.warns++
	case checkFail:
		r.fails++
	}
}

func (r *doctorReport) ok(section *doctorSection, label, detail string) {
	r.add(section, checkOK, label, detail)
}

func (r *doctorReport) warn(section *doctorSection, label, detail string) {
	r.add(section, checkWarn, label, detail)
}

func (r *doctorReport) fail(section *doctorSection, label, detail string) {
	r.add(section, checkFail, label, detail)
}

func (r *doctorReport) render() {
	for _, section := range r.sections {
		if len(section.checks) == 0 {
			continue
		}

		fmt.Println(ui.Bold.Render(section.title))
		for _, check := range section.checks {
			icon := ui.IconSuccess
			detailStyle := ui.Muted

			switch check.level {
			case checkWarn:
				icon = ui.IconWarning
				detailStyle = ui.Warning
			case checkFail:
				icon = ui.IconError
				detailStyle = ui.Error
			}

			fmt.Printf("  %s %s %s\n",
				icon,
				fmt.Sprintf("%-26s", check.label),
				detailStyle.Render(check.detail),
			)
		}
		fmt.Println()
	}
}

func (r *doctorReport) finish() {
	r.render()
	printDoctorSummary(r)

	if r.fails > 0 {
		os.Exit(1)
	}
}

func printDoctorSummary(report *doctorReport) {
	fmt.Println(ui.Divider(44))

	switch {
	case report.fails == 0 && report.warns == 0:
		fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render("Everything looks good."))
	case report.fails == 0:
		fmt.Printf("%s %s\n", ui.IconWarning,
			ui.Warning.Render(fmt.Sprintf("%d warning(s), no errors.", report.warns)))
	default:
		fmt.Printf("%s %s\n", ui.IconError,
			ui.Error.Render(fmt.Sprintf("%d error(s), %d warning(s).", report.fails, report.warns)))
	}

	if report.fails > 0 {
		fmt.Println(ui.Muted.Render("Fix the errors above, then run 'secwan doctor' again."))
	}
}
