# 🔐 secret-wan

**A modern, secure, and interactive CLI secret manager for developers.**

`secret-wan` allows you to define your secret schemas in a clean TOML file, interactively fill them via a beautiful TUI, and securely store them either in local encrypted files or your OS's native Keychain. Finally, inject them into your application at runtime without leaking them into your shell history.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)
![Release](https://img.shields.io/github/v/release/ikhwan-satrio/secret-wan?style=for-the-badge)

---

## ✨ Features

- 🎨 **Beautiful Interactive TUI**: Powered by [Bubbletea](https://github.com/charmbracelet/bubbletea) for a smooth, vim-like secret input experience.
- 📝 **TOML Configuration**: Say goodbye to YAML indentation errors. Define your schemas in clean, readable TOML.
- 🌍 **Multi-Environment Support**: Manage secrets for `default`, `staging`, `production`, or any custom profile.
- 🎲 **Auto-Generation**: Automatically generate cryptographically secure passwords for specific keys.
- 🗄️ **Flexible Storage Providers**:
  - `dotfiles`: Store secrets in a local, age-encrypted file (`.secrets.age`). Perfect for GitOps.
  - `keyring`: Store secrets in your OS native Keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service/KeePassXC).
- 🚀 **Runtime Injection (`exec`)**: Run your apps with secrets injected directly into the environment, bypassing shell history leaks.
- 🛡️ **Zero-Trust Encryption**: Uses [age](https://github.com/FiloSottile/age), a modern, simple, and highly secure file encryption tool.

---

## 📦 Installation

### Via Go Install
```bash
go install github.com/ikhwan-satrio/secret-wan@latest

### Via Pre-compiled Binaries
Download the latest binary for your OS/Architecture from the [GitHub Releases](https://github.com/ikhwan-satrio/secret-wan/releases) page.

*(Note for Linux users: If you plan to use the `keyring` provider, ensure you have `libsecret` and a Secret Service daemon like `gnome-keyring` or `keepassxc` installed).*

---

## 🚀 Quick Start

### 1. Initialize your project
Run the init command. It will generate your master key and prompt you to choose a storage provider.
```bash
secret-wan init
```

### 2. Define your schema
Edit the generated `secret-wan.toml` file to define your secrets:
```toml
[project]
name = "my-awesome-app"
provider = "dotfiles" # or "keyring"

[profiles.default]
DB_PASSWORD = { generate = true }
DB_URL = { required = true }
API_KEY = { required = true }

[profiles.staging]
AWS_ACCESS_KEY = { required = true }
AWS_SECRET_KEY = { required = true }
```

### 3. Sync your secrets
Run sync to interactively fill in the required secrets. Auto-generated ones will be created instantly.
```bash
secwan sync
# or for a specific profile:
secwan sync -p staging
```

### 4. Run your app with secrets injected
Use `exec` to run your application. Secrets are injected into the child process environment safely.
```bash
secwan exec -- npm start
# or
secwan exec -p staging -- docker-compose up
```

---

## 📖 Command Reference

| Command | Description |
| :--- | :--- |
| `init` | Initialize master key and generate `secret-wan.toml` template. |
| `sync` | Interactively configure secrets based on the TOML schema. |
| `check` | Verify if all required secrets are present and configured. |
| `env` | Print decrypted secrets to stdout in `KEY="VALUE"` format. |
| `exec` | Execute a command with secrets injected into the environment. |

### Global Flags
- `-p, --profile string`: Specify the environment profile to use (default `"default"`).

---

## 🗄️ Storage Providers

When you run `secwan init`, you will be asked to choose a provider:

### 1. Dotfiles (Local Encrypted File)
- **How it works**: Secrets are encrypted using `age` and saved to `.secrets.age` in your project root.
- **Pros**: Easy to commit to Git (the file is encrypted), easy to share with the team (just share the age public key).
- **Cons**: Requires managing the age private key (`~/.config/secret-wan/keys.txt`).

### 2. Keyring (OS Native Keychain)
- **How it works**: Secrets are stored in your operating system's secure credential manager.
- **Pros**: Highest security, leverages OS-level encryption/biometrics, no encrypted files cluttering your repo.
- **Cons**: Harder to share with team members, requires specific OS dependencies (e.g., `libsecret` on Linux).

---

## 🛠️ Development & Building

To build from source:

```bash
# Clone the repository
git clone https://github.com/ikhwan-satrio/secret-wan.git
cd secret-wan

# Build with optimizations (smaller binary size)
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o secret-wan .
```

### Testing Locally with GoReleaser
If you want to test the cross-compilation and release process locally:
```bash
goreleaser release --snapshot --clean
```

---

## 🧰 Tech Stack

Built with love using the best tools in the Go ecosystem:
- **[Cobra](https://github.com/spf13/cobra)**: For the CLI framework.
- **[Bubbletea](https://github.com/charmbracelet/bubbletea)**: For the interactive TUI.
- **[Age](https://github.com/FiloSottile/age)**: For modern, secure file encryption.
- **[BurntSushi/toml](https://github.com/BurntSushi/toml)**: For TOML parsing.
- **[99designs/keyring](https://github.com/99designs/keyring)**: For cross-platform OS keychain access.
- **[GoReleaser](https://goreleaser.com/)**: For automated cross-compilation and releases.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  Made with ❤️ and ☕ by <a href="https://github.com/ikhwan-satrio">Ikhwan Satrio</a>
</p>
