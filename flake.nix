{
  description = "secret-wan — secure CLI secret manager for developers";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;

      version = "1.1.0";
      rev = "v${version}";
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};

          src = pkgs.fetchFromGitHub {
            owner = "ikhwan-satrio";
            repo = "secret-wan";
            inherit rev;
            hash = "sha256-HqHqufTiGLER8jh91p0N0VGcqI66o2ArfiXRWmj+V3k=";
          };
        in
        {
          default = pkgs.buildGoModule {
            pname = "secret-wan";
            inherit version src;

            # Updated automatically by the "Auto-update Flake on Release"
            # workflow (nix-update) whenever a new GitHub release is published.
            vendorHash = "sha256-vPjN1qzy4vo6qNYDTQtjUpuFriskoULsU9GS6ptZIe0=";

            subPackages = [ "." ];

            ldflags = [
              "-s"
              "-w"
              "-X main.version=${version}"
              "-X main.commit=${rev}"
              "-X main.date=unknown"
            ];

            # CI runs `go test ./...` separately; keep the Nix build focused on
            # producing the release binary.
            doCheck = false;

            postInstall = ''
              if [ -e "$out/bin/secret-wan" ]; then
                mv "$out/bin/secret-wan" "$out/bin/secwan"
              fi
            '';

            meta = {
              description = "Secure secret manager for your projects";
              homepage = "https://github.com/ikhwan-satrio/secret-wan";
              license = pkgs.lib.licenses.mit;
              mainProgram = "secwan";
            };
          };
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              goreleaser
            ];
          };
        }
      );

      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt);
    };
}
