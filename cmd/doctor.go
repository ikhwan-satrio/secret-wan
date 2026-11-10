package cmd

import (
	"github.com/ikhwan-satrio/secret-wan/internal/doctor"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose configuration, keys, storage, and secrets",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		doctor.Run()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
