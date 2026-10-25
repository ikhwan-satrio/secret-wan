package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ikhwan-satrio/secret-wan/config"
	"github.com/ikhwan-satrio/secret-wan/tui" // Pastikan path ini sesuai dengan lokasi package tui Anda
	"github.com/ikhwan-satrio/secret-wan/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inisialisasi master key dan template config",
	Run: func(cmd *cobra.Command, args []string) {
		appConfig := config.UseApp()
		cryptoModule := utils.CryptoModule{AppConfig: appConfig}

		// 1. Tampilkan TUI untuk memilih provider
		p := tea.NewProgram(tui.SelectModel{
			Title:   "Select storage provider:",
			Choices: []string{"dotfiles (Local encrypted file)", "keyring (OS Keychain)"},
		})

		finalModel, err := p.Run()
		if err != nil {
			fmt.Println(FormatError("TUI crashed", err.Error()))
			os.Exit(1)
		}

		selModel := finalModel.(tui.SelectModel)
		if selModel.Cancel {
			fmt.Println("\n" + StyleDim.Render("  Initialization cancelled."))
			return
		}

		// Tentukan nama provider berdasarkan pilihan
		providerName := "dotfiles"
		if strings.HasPrefix(selModel.Result, "keyring") {
			providerName = "keyring"
		}

		// 2. Load or Generate Master Key (Bukan LoadSecrets)
		_, err = cryptoModule.LoadOrGenerateKey()
		if err != nil {
			fmt.Println(FormatError("Failed to initialize master key", err.Error()))
			os.Exit(1)
		}

		// 3. Generate Template TOML yang menyertakan provider
		specTemplate := utils.GetTemplate(providerName)

		if _, err := os.Stat(appConfig.SpecFile); err == nil {
			fmt.Printf("%s %s\n", IconInfo, StyleDim.Render(fmt.Sprintf("%s already exists, skipped creation.", appConfig.SpecFile)))
		} else {
			if err := os.WriteFile(appConfig.SpecFile, []byte(specTemplate), 0644); err != nil {
				fmt.Println(FormatError("Failed to create config file", err.Error()))
				os.Exit(1)
			}
			fmt.Printf("%s %s\n", IconSuccess, StyleSuccess.Render(fmt.Sprintf("Created master key and %s", appConfig.SpecFile)))
		}

		fmt.Println("\n" + StyleInfo.Render("Next steps:"))
		fmt.Println(StyleDim.Render("  1. secman sync          # Verify and set your secrets interactively"))
		fmt.Println(StyleDim.Render("  2. secman check         # Verify all secrets are present and valid"))
		fmt.Println(StyleDim.Render("  3. secman exec -- <cmd> # Run your app with secrets injected"))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
