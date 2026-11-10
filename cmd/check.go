package cmd

import (
	"github.com/ikhwan-satrio/secret-wan/internal/check"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check secret completeness based on config",
	Run: func(cmd *cobra.Command, args []string) {
		profileName, _ := cmd.Flags().GetString("profile")
		strict, _ := cmd.Flags().GetBool("strict")

		check.Run(profileName, strict)
	},
}

func init() {
	checkCmd.Flags().StringP("profile", "p", "default", "Profile to check")
	checkCmd.Flags().Bool("strict", false, "Also report unmanaged secrets and invalid manifest rules")
	rootCmd.AddCommand(checkCmd)
}
