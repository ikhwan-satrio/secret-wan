package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var recipientMeCmd = &cobra.Command{
	Use:     "me",
	Aliases: []string{"whoami"},
	Short:   "Print your public age recipient",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		session := newSession()
		session.requireExistingIdentity()

		self := session.Identity.Recipient().String()
		fmt.Println(self)

		list, err := session.ListRecipients()
		if err != nil {
			warn(fmt.Sprintf("Could not read recipients: %v", err))
			return
		}

		if len(list.Keys) > 0 && !list.SelfListed() {
			warn("Your key is not in the project's recipient list.")
		}
	},
}
