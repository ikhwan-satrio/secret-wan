# Installing secret-wan on NixOS

`secret-wan` ships a flake with a default package and app for
`x86_64-linux`, `aarch64-linux`, `x86_64-darwin`, and `aarch64-darwin`.
The installed binary is called `secwan`.

The package builds the source from the release tag pinned in
[`flake.nix`](../flake.nix), so Nix builds are reproducible and match the
published GitHub release.

## Run without installing

```bash
nix run github:ikhwan-satrio/secret-wan -- version
```

Everything after `--` is passed to `secwan`, for example:

```bash
nix run github:ikhwan-satrio/secret-wan -- init
```

## Install imperatively

```bash
nix profile install github:ikhwan-satrio/secret-wan
secwan version
```

Update later with:

```bash
nix profile upgrade --all
```

## NixOS with flakes (system-wide)

Add the flake as an input and install its default package:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    secret-wan.url = "github:ikhwan-satrio/secret-wan";
  };

  outputs =
    { self, nixpkgs, secret-wan, ... }:
    {
      nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [
          (
            { pkgs, ... }:
            {
              environment.systemPackages = [
                secret-wan.packages.${pkgs.system}.default
              ];
            }
          )
        ];
      };
    };
}
```

Apply it with:

```bash
sudo nixos-rebuild switch --flake .#myhost
```

If your configuration passes inputs through `specialArgs`, use
`inputs.secret-wan.packages.${pkgs.system}.default` instead.

## Home Manager

```nix
{ inputs, pkgs, ... }:
{
  home.packages = [
    inputs.secret-wan.packages.${pkgs.system}.default
  ];
}
```

For a standalone Home Manager flake, add the same
`secret-wan.url = "github:ikhwan-satrio/secret-wan";` input and pass `inputs`
to your modules.

## Overlay (optional)

```nix
{
  nixpkgs.overlays = [
    (final: prev: {
      secret-wan = inputs.secret-wan.packages.${final.system}.default;
    })
  ];

  environment.systemPackages = [ pkgs.secret-wan ];
}
```

## Updating the package

The flake pins a release tag. Refresh the input to pick up the newest
release published by the project:

```bash
nix flake update secret-wan
```

Then rebuild your system or Home Manager profile.

## Keyring provider on NixOS

The `dotfiles` provider needs no extra runtime services. The `keyring`
provider talks to a D-Bus Secret Service implementation, such as GNOME
Keyring or KeePassXC.

For GNOME Keyring, enable the daemon and PAM integration:

```nix
{
  services.gnome.gnome-keyring.enable = true;
  security.pam.services.login.enableGnomeKeyring = true;
}
```

Log out and back in after the first rebuild so the keyring is unlocked with
your session. On non-GNOME desktops, run KeePassXC with its Secret Service
integration enabled instead.

Verify that a Secret Service is reachable:

```bash
echo "$DBUS_SESSION_BUS_ADDRESS"
secret-tool search --all service secret-wan
```

If `DBUS_SESSION_BUS_ADDRESS` is empty, the application cannot reach the
keyring from that shell.

## Troubleshooting

- **`experimental Nix feature 'nix-command' is disabled`** — enable flakes:

  ```nix
  # /etc/nix/nix.conf
  experimental-features = nix-command flakes
  ```

- **Command not found** — the package installs `secwan`, not `secret-wan`.
  Use `secwan version` to confirm the installation.

- **Keyring errors** — make sure a Secret Service daemon is running in the
  same user session and that `DBUS_SESSION_BUS_ADDRESS` is set. The
  `dotfiles` provider does not require D-Bus.

- **Unsupported system** — the flake currently covers
  `x86_64-linux`, `aarch64-linux`, `x86_64-darwin`, and `aarch64-darwin`.
