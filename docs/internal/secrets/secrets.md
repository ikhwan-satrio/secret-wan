# `internal/secrets/secrets.go`

> Source: `internal/secrets/secrets.go` · Package: `secrets`

## Purpose

Provides secret value generation and terminal/piped input prompts for commands
that need a single value, primarily `secwan set`.

## Responsibilities

- Generate a random URL-safe secret (`GenerateSecureSecret`).
- Read a secret from a terminal without echo.
- Read a secret from piped stdin when stdin is not a terminal.
- Read a visible (non-secret) value from the terminal.
- Keep prompts on stderr so stdout remains script-friendly.

## Key type

```go
type SecretModule struct {
    AppConfig config.AppConfig
}
```

## Key functions

### `GenerateSecureSecret() (string, error)`

1. Allocates `AppConfig.SecretLength` random bytes.
2. Fills them with `crypto/rand.Read`.
3. Encodes with `base64.URLEncoding`.

The output is longer than `SecretLength` because of base64 expansion
(approximately 4/3 of the input length). This function is distinct from
`CryptoModule.GenerateSecurePassword`, which produces exactly the requested
number of characters from a custom charset.

### `PromptUserInput(key string, isSecret bool) (string, error)`

#### Secret + non-terminal stdin

```go
if isSecret && !term.IsTerminal(int(os.Stdin.Fd())) {
    data, err := io.ReadAll(os.Stdin)
    ...
    return strings.TrimRight(string(data), "\r\n"), nil
}
```

When input is piped, the function reads the entire stream and removes trailing
carriage returns and newlines. This supports:

```bash
printf 'super-secret' | secwan set API_KEY
```

#### Secret + terminal

1. Writes `Enter value for <KEY> (hidden input): ` to stderr.
2. Calls `term.ReadPassword` so characters are not echoed.
3. Writes a newline to stderr.
4. Returns the raw password string.

#### Non-secret + terminal

1. Writes `Enter value for <KEY>: ` to stderr.
2. Reads a line with `bufio.Reader`.
3. Trims surrounding whitespace with `strings.TrimSpace`.

## Dependencies

- `internal/config` for the secret length.
- `golang.org/x/term` for hidden input and terminal detection.
- `crypto/rand`, `encoding/base64`, `io`, `os`, `bufio`, `strings`.

## Edge cases and notes

- Secret piping trims only trailing CR/LF, preserving leading/trailing spaces
  that may be meaningful in a secret.
- Interactive hidden input does not trim spaces.
- The visible prompt path trims all surrounding whitespace.
- Prompts go to stderr; values are returned to the caller and never printed
  here.
- `PromptUserInput` does not validate emptiness. `cmd/set.go` rejects an empty
  result afterward.
- `GenerateSecureSecret` is currently not used by the main command paths, which
  use the character-based generator in the crypto module.
- Reading from a pipe blocks until EOF.

## Related files

- [`internal/crypto/crypto.md`](../crypto/crypto.md) — `GenerateSecurePassword`.
- [`cmd/set.md`](../../cmd/set.md) — main consumer of `PromptUserInput`.
- [`internal/config/config.md`](../config/config.md) — `SecretLength`.
