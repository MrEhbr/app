{
  description = "A Nix-flake-based Go development environment";

  inputs.nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1.*.tar.gz";

  outputs =
    { self
    , nixpkgs
    ,
    }:
    let
      goVersion = 21;
      overlays = [
        (final: prev: {
          go = prev."go_1_${toString goVersion}";
          gofumpt = prev.gofumpt.overrideAttrs { version = "0.5.0"; };
        })
      ];
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f:
        nixpkgs.lib.genAttrs supportedSystems (system:
          f {
            pkgs = import nixpkgs { inherit overlays system; };
          });
    in
    {
      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go

            # goimports, godoc, formaters etc.
            gotools
            gofumpt
            golines
            goimports-reviser

            # https://github.com/golangci/golangci-lint
            golangci-lint

            # Go lsp
            gopls
            # Go debuger
            delve

            # Generators
            go-minimock
          ];
        };
      });
    };
}
