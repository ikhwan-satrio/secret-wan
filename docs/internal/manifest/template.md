# `internal/manifest/template.go`

> Source: `internal/manifest/template.go` · Package: `manifest`

## Purpose

Generates the starter `secret-wan.toml` content written by `secwan init`.

## Responsibilities

- Derive the default project name from the current directory.
- Insert the selected storage provider.
- Provide a minimal `default` profile with one generated and one required key.
- Return the template as a formatted string.

## Key code

```go
func GetTemplate(provider string) string {
    cwd, _ := os.Getwd()
    parrent := filepath.Base(cwd)

    specTemplate := fmt.Sprintf(`
[project]
name = "%s"
provider = "%s"

[profiles.default]
DB_PASS = { generate = true }
DB_URL = { required = true }
    `, parrent, provider)

    return specTemplate
}
```

The local variable is spelled `parrent`; the value is the base name of the
current working directory.

## Generated TOML

```toml
[project]
name = "<current-directory-name>"
provider = "dotfiles"   # or "keyring"

[profiles.default]
DB_PASS = { generate = true }
DB_URL = { required = true }
```

## Behavior notes

- The template starts with a newline and ends with a newline plus indentation
  from the raw string literal. TOML ignores the leading whitespace, and the
  trailing spaces are harmless.
- `os.Getwd()` errors are ignored. If the working directory cannot be resolved,
  `cwd` is empty and `filepath.Base("")` yields `"."`, which is still valid TOML
  but not a useful project name.
- The project name is not escaped. A directory name containing a double quote
  would produce invalid TOML; this is unlikely but possible.
- Only `dotfiles` and `keyring` are offered by `init`, but the function accepts
  any provider string, so `docker` could be passed programmatically.
- Generated keys use the length configured in `AppConfig.SecretLength`; the
  template itself does not include a length setting.

## Dependencies

- Standard library `fmt`, `os`, `path/filepath`.

## Related files

- [`cmd/init.md`](../../cmd/init.md) — writes the template to disk.
- [`internal/manifest/parser.md`](parser.md) — parses the generated file.
