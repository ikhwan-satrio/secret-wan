package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/ikhwan-satrio/secret-wan/config"
	"github.com/ikhwan-satrio/secret-wan/utils"
	"github.com/spf13/cobra"
)

var printEnvCmd = &cobra.Command{
	Use:   "print-env",
	Short: "Menampilkan nilai secret yang terdekripsi dalam format KEY=VALUE",
	Run: func(cmd *cobra.Command, args []string) {
		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			fmt.Println(FormatError("Failed to read profile flag", err.Error()))
			os.Exit(1)
		}

		appConfig := config.UseApp()
		cryptoModule := utils.CryptoModule{AppConfig: appConfig}
		parserModule := utils.ParserModule{AppConfig: appConfig}

		// 1. Load Master Key
		identity, err := cryptoModule.LoadOrGenerateKey()
		if err != nil {
			fmt.Println(FormatError("Failed to load master key", err.Error()))
			os.Exit(1)
		}

		// 2. Parse Manifest
		manifest, err := parserModule.ParseManifest()
		if err != nil {
			fmt.Println(FormatError("Failed to parse manifest", err.Error()))
			os.Exit(1)
		}

		// 3. Validasi Profile
		targetProfile, exists := manifest.Profiles[profileName]
		if !exists {
			validProfiles := getProfileNames(manifest.Profiles)
			errMsg := fmt.Sprintf("Profile '%s' tidak ditemukan. Profile yang tersedia: %v", profileName, validProfiles)
			fmt.Println(FormatError("Invalid profile", errMsg))
			os.Exit(1)
		}

		// 4. Load Secrets menggunakan Provider yang tepat (Dotfiles atau Keyring)
		provider := utils.GetStorageProvider(appConfig, manifest.Project.Provider, manifest.Project.Name)
		secrets, err := cryptoModule.LoadSecrets(provider, identity)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(FormatError("Secrets storage not found",
					fmt.Sprintf("Jalankan 'secman sync -p %s' terlebih dahulu.", profileName)))
			} else {
				fmt.Println(FormatError("Failed to decrypt secrets", err.Error()))
			}
			os.Exit(1)
		}

		// 5. Filter hanya secret yang didefinisikan di profile ini
		allowedKeys := make(map[string]bool)
		for _, k := range targetProfile.Required {
			allowedKeys[k] = true
		}
		for _, k := range targetProfile.AutoGen {
			allowedKeys[k] = true
		}

		// Kumpulkan key yang benar-benar ada nilainya, lalu urutkan agar output rapi
		var activeKeys []string
		for k := range allowedKeys {
			if _, exists := secrets[k]; exists {
				activeKeys = append(activeKeys, k)
			}
		}
		sort.Strings(activeKeys)

		// 6. Tampilkan Output
		if len(activeKeys) == 0 {
			fmt.Println(StyleDim.Render(fmt.Sprintf("# Tidak ada secret yang dikonfigurasi untuk profile '%s'", targetProfile.Name)))
			return
		}

		fmt.Printf("%s %s\n", StyleDim.Render("# Secrets for profile:"), StyleAccent.Render(targetProfile.Name))
		fmt.Printf("%s\n", StyleDim.Render("# WARNING: Output ini berisi data sensitive.不要 share screenshot-nya."))
		
		for _, k := range activeKeys {
			// Format KEY="VALUE" agar aman jika di-copy paste (menangani spasi atau karakter khusus)
			fmt.Printf("%s=\"%s\"\n", k, secrets[k])
		}
	},
}

func init() {
	printEnvCmd.Flags().StringP("profile", "p", "default", "Profile environment yang ingin ditampilkan")
	rootCmd.AddCommand(printEnvCmd)
}
