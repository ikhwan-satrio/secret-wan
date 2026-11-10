package recipient

import (
	"fmt"

	"github.com/ikhwan-satrio/secret-wan/internal/cli"
	"github.com/spf13/cobra"
)

func newMeCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "me",
		Aliases: []string{"whoami"},
		Short:   "Print your public age recipient",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			session := cli.NewSession()
			session.RequireExistingIdentity()

			self := session.Identity.Recipient().String()
			fmt.Println(self)

			list, err := session.ListRecipients()
			if err != nil {
				cli.Warn(fmt.Sprintf("Could not read recipients: %v", err))
				return
			}

			if len(list.Keys) > 0 && !list.SelfListed() {
				cli.Warn("Your key is not in the project's recipient list.")
			}
		},
	}
}
