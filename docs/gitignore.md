# `.gitignore`

> Source: `.gitignore` · Git ignore rules

## Purpose

Tells Git which files and directories should not be tracked.

## Content

```gitignore
/target
```

## Effect

- The `target/` directory at the repository root is ignored.
- This is where local build output lives, including `target/secwan`.
- Because the pattern starts with `/`, it is anchored to the repository root. A
  nested directory named `target` elsewhere would not be ignored by this rule.

## Why it exists

Local builds, release snapshots, and cross-compilation experiments should not be
committed. Keeping build artifacts out of Git avoids large binaries in history.

## Edge cases and notes

- `env.enc` is **not** ignored. For the `dotfiles` provider, the encrypted file
  is intentionally committable so teams can share it. The master key
  (`~/.config/secret-wan/keys.txt`) lives outside the repository and is never
  tracked.
- `secret-wan.toml` is not ignored; it is part of the project definition.
- If you build to a different output path, add that path here.
- Git ignore rules do not affect already tracked files; use `git rm --cached`
  if a previously tracked artifact must be removed.

## Related files

- [`target-secwan.md`](target-secwan.md) — ignored local binary.
- [`internal/provider/provider.md`](internal/provider/provider.md) — `env.enc`
  storage.
