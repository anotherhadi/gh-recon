{
  description =
    "Retrieves and aggregates public OSINT data about a GitHub user using Go and the GitHub API. Finds hidden emails in commit history, previous usernames, friends, other GitHub accounts, and more.";

  inputs = { nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable"; };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" ];

      forAllSystems = f:
        nixpkgs.lib.genAttrs supportedSystems
        (system: f system (import nixpkgs { inherit system; }));

      pname = "gh-recon";
      version = "1.3.0";

      ldflags = [ "-s" "-w" ];

    in {
      packages = forAllSystems (system: pkgs: {
        "${pname}" = pkgs.buildGoModule {
          inherit pname version ldflags;

          src = ./.;

          vendorHash = "sha256-S8IzmdiVvBtnQQl0AewGZ1yuitvrdnVQ/Jf2230g3Mg=";

          meta = with pkgs.lib; {
            description =
              "Retrieves and aggregates public OSINT data about a GitHub user using Go and the GitHub API. Finds hidden emails in commit history, previous usernames, friends, other GitHub accounts, and more.";
            homepage = "https://github.com/anotherhadi/gh-recon";
            platforms = platforms.unix;
          };
        };
      });

      defaultPackage =
        forAllSystems (system: pkgs: self.packages.${system}.${pname});
    };
}
