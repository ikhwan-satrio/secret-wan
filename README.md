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
  - `dotfiles`: Store secrets in a local, age-encrypted file (`env.enc` by default, configurable via `EncFile`). Perfect for GitOps.
  - `keyring`: Store secrets in your OS native Keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service/KeePassXC).
  - `docker`: Read a mounted Docker secret at runtime (read-only).
  - `env`: Read a base64 payload from `SECWAN_ENV_ENC` (read-only, ideal for CI).
  - `s3`: Store one encrypted object in AWS S3 or an S3-compatible bucket.
- 🚀 **Runtime Injection (`exec`)**: Run your apps with secrets injected directly into the environment, bypassing shell history leaks.
- 🔑 **Per-Key Management**: Use `list`, `get`, `set`, and `unset` for quick day-to-day changes without opening the full TUI.
- 👥 **Multi-Recipient Sharing**: Add teammates with `recipient add`, list access with `recipient list`, and re-encrypt on demand with `rekey`. Secrets are encrypted to every configured age public key.
- 🩺 **Built-in Diagnostics**: Run `doctor` to validate your manifest, master key, recipients, storage provider, and secret consistency.
- 🤖 **Script-Friendly Output**: `print-env` writes logs to stderr and supports `--json`; `check --strict` catches unmanaged secrets.
- 🛡️ **Zero-Trust Encryption**: Uses [age](https://github.com/FiloSottile/age), a modern, simple, and highly secure file encryption tool.

---

## 📚 Documentation

Every source, workflow, and configuration file is documented in the
[`docs/`](docs/README.md) directory. Start with the
[documentation index](docs/README.md) for an architecture overview and links to
each per-file guide.

---

## 📦 Installation

### Via Go Install
```bash
go install github.com/ikhwan-satrio/secret-wan@latest
```

### Via Pre-compiled Binaries
Download the latest binary for your OS/Architecture from the [GitHub Releases](https://github.com/ikhwan-satrio/secret-wan/releases) page.

Windows users can download either the `.zip` archive or the standalone `secwan-windows-amd64.exe` / `secwan-windows-arm64.exe` assets.

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
provider = "dotfiles" # or "keyring", "docker", "env", "s3"

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
| `check` | Verify if all required secrets are present and configured. Add `--strict` to flag unmanaged secrets and invalid rules. |
| `doctor` | Diagnose the manifest, master key permissions, storage provider, and profile consistency. |
| `list` | List secrets for a profile. Values are masked unless `--reveal` is used. |
| `get` | Print a single secret. Values are masked unless `--reveal` is used. |
| `set` | Create or update one secret without opening the TUI. Use `--generate` for random values. |
| `unset` | Remove one secret from storage. |
| `recipient` | List, identify, add, or remove age recipients. Changes rekey existing secrets automatically. |
| `rekey` | Re-encrypt all stored secrets for the configured recipients. |
| `print-env` | Print decrypted secrets to stdout in `KEY="VALUE"` format. Logs go to stderr. |
| `exec` | Execute a command with secrets injected into the environment. |
| `version` | Print version, commit, build date, Go version, and platform. |

### Common Flags
- `-p, --profile string`: Select the environment profile (available on profile-aware commands, default `"default"`).
- `--version`: Print version information and exit.

### Scripting with `print-env`

`print-env` keeps stdout clean, so it can be evaluated directly:

```bash
eval "$(secwan print-env)"
```

Get structured output for other tools:

```bash
secwan print-env --json
secwan print-env -p staging --json | jq -r '.API_KEY'
```

### Daily secret management

```bash
# List keys (values masked)
secwan list

# Show a value in plain text
secwan get API_KEY --reveal

# Set a value interactively (input is hidden)
secwan set API_KEY

# Generate a secure random value
secwan set DB_PASSWORD --generate

# Remove a key
secwan unset OLD_KEY

# Check for missing and unmanaged secrets
secwan check --strict

# Diagnose the whole setup
secwan doctor
```

---

## 👥 Sharing with your team

`secret-wan` encrypts secrets to every public age recipient listed in
`secret-wan.recipients`. The file contains public keys only and is safe to
commit.

```bash
# Show your public key (send this to a teammate)
secwan recipient me

# See who can decrypt the project
secwan recipient list

# Add a teammate; existing secrets are re-encrypted automatically
secwan recipient add age1...

# Remove access; secrets are re-encrypted for the remaining keys
secwan recipient remove age1...

# Re-encrypt after editing secret-wan.recipients by hand
secwan rekey
```

### How the recipient list works

- New projects created by `init` get a `secret-wan.recipients` file seeded with
  the current key.
- Existing projects without the file keep the original single-recipient
  behavior: the current key is used for new writes.
- `set`, `unset`, and `sync` always encrypt to the full recipient list.
- Normal writes fail if your own key is not in the list, which prevents
  accidental lockout.
- Removing your own key or excluding it with `rekey` requires `--force`.
- A removed key can still decrypt old Git history. Rotate any secret the removed
  teammate should no longer know.

---

## 🗄️ Storage Providers

When you run `secwan init`, you will be asked to choose between `dotfiles` and
`keyring`. The other providers can be selected by editing `project.provider` in
`secret-wan.toml`; `doctor` never requires a specific one.

### 1. Dotfiles (Local Encrypted File)
- **How it works**: Secrets are encrypted using `age` and saved to the configured encrypted file (`env.enc` by default) in your project root.
- **Pros**: Easy to commit to Git (the file is encrypted), easy to share with the team via `secret-wan.recipients` and `recipient add`.
- **Cons**: Requires managing the age private key (`~/.config/secret-wan/keys.txt`).

### 2. Keyring (OS Native Keychain)
- **How it works**: Secrets are stored in your operating system's secure credential manager.
- **Pros**: Highest security, leverages OS-level encryption/biometrics, no encrypted files cluttering your repo.
- **Cons**: Harder to share with team members, requires specific OS dependencies (e.g., `libsecret` on Linux).

### 3. Docker Secret (Read-Only)
- **How it works**: Reads a mounted secret such as `/run/secrets/<project>`.
- **Pros**: No extra configuration in the container; the platform manages distribution.
- **Cons**: Read-only, so `sync`/`set` cannot update it through `secret-wan`.

### 4. Environment Variable (Read-Only)
- **How it works**: Reads base64 ciphertext from `SECWAN_ENV_ENC`.
- **Pros**: Simple for CI pipelines and ephemeral containers.
- **Cons**: Read-only, must be base64-encoded, and visible to the process environment.

### 5. S3-Compatible Object Storage
- **How it works**: Stores one encrypted object per project, by default at
  `<project>/<configured encrypted file>` (default `<project>/env.enc`), using
  AWS SDK for Go v2.
- **Configuration**: `SECWAN_S3_BUCKET` (required), plus optional
  `SECWAN_S3_KEY`, `SECWAN_S3_REGION`, `SECWAN_S3_ENDPOINT`, and
  `SECWAN_S3_PATH_STYLE`. Credentials come from the standard AWS SDK chain.
- **Pros**: Central storage for teams and CI, works with AWS S3, MinIO,
  Cloudflare R2, and Backblaze B2.
- **Cons**: Needs network access, bucket permissions, and credential management.

---

## 🛠️ Development & Building

### Project Structure

```text
.
├── cmd/                  # Cobra command definitions and output helpers
├── internal/
│   ├── app/              # Application service: business logic used by commands
│   ├── config/           # Application configuration
│   ├── crypto/           # age encryption and master key handling
│   ├── manifest/         # TOML manifest parsing, queries, and templates
│   ├── osutil/           # OS-specific helpers
│   ├── provider/         # Storage providers (dotfiles, keyring, Docker, env, S3)
│   ├── recipients/       # Public age recipient list parsing and storage
│   ├── secrets/          # Secret generation and prompt helpers
│   ├── tui/              # Bubble Tea screens
│   └── ui/               # Shared terminal theme and formatting
└── main.go               # CLI entrypoint
```

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

### Automated Releases

Releases are automated with [Release Please](https://github.com/googleapis/release-please). Use [Conventional Commit](https://www.conventionalcommits.org/) messages so the changelog and semver bump are generated correctly:

```text
feat: add Windows support
fix: handle empty manifest
docs: update installation guide
feat!: change config format
```

Flow:

1. Merge a PR into `main` using a conventional commit title.
2. Release Please opens or updates a release PR with the next version and `CHANGELOG.md` entry.
3. Merge the release PR. Release Please creates the tag and GitHub Release.
4. The release workflow builds all binaries and attaches them to the release, including `secwan-windows-amd64.exe` and `secwan-windows-arm64.exe`.

You can also trigger a release manually by pushing a tag:

```bash
git tag v1.0.1
git push origin v1.0.1
```

---

## 🧰 Tech Stack

Built with love using the best tools in the Go ecosystem:
- **[Cobra](https://github.com/spf13/cobra)**: For the CLI framework.
- **[Bubbletea](https://github.com/charmbracelet/bubbletea)**: For the interactive TUI.
- **[Age](https://github.com/FiloSottile/age)**: For modern, secure file encryption.
- **[BurntSushi/toml](https://github.com/BurntSushi/toml)**: For TOML parsing.
- **[99designs/keyring](https://github.com/99designs/keyring)**: For cross-platform OS keychain access.
- **[AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)**: For S3-compatible object storage.
- **[GoReleaser](https://goreleaser.com/)**: For automated cross-compilation and releases.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  Made with ❤️ and ☕ by <a href="https://github.com/ikhwan-satrio">Ikhwan Satrio</a>
</p>
