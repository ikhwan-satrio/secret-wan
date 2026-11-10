package recipient

import (
	"fmt"

	"github.com/ikhwan-satrio/secret-wan/internal/cli"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configured age recipients",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			session := cli.NewSession()
			list, err := session.ListRecipients()
			if err != nil {
				cli.FailErr("Failed to read recipients", err)
			}

			cli.PrintHeader("recipient list")
			fmt.Println(ui.Muted.Render("File: ") + ui.Accent.Render(list.Path))
			fmt.Println()

			if len(list.Keys) == 0 {
				fmt.Println(ui.Muted.Render("No recipient list yet; the project uses single-recipient mode."))
				if list.SelfKnown {
					fmt.Printf("  %s %s\n", ui.Key.Render(list.Self), ui.Muted.Render("(you, implicit)"))
				} else {
					fmt.Println(ui.Muted.Render("No master key found. Run 'secwan init' first."))
				}
				return
			}

			for _, key := range list.Keys {
				if key == list.Self {
					fmt.Printf("  %s %s\n", ui.Key.Render(key), ui.Success.Render("(you)"))
					continue
				}
				fmt.Printf("  %s\n", ui.Key.Render(key))
			}

			fmt.Println()
			fmt.Println(ui.Muted.Render(fmt.Sprintf("%d recipient(s) configured.", len(list.Keys))))

			if list.SelfKnown && !list.SelfListed() {
				cli.Warn("Your key is not in this list. Future writes would not be decryptable by you.")
			}
		},
	}
}
