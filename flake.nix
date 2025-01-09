{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        goVendorHash = import ./nix/goVendorHash.nix;
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages = rec {
          default = suslik;
          suslik = pkgs.buildGoModule {
            pname = "suslik";
            version = "0.1.0";
            src = ./.;
            vendorHash = goVendorHash;
          };
        };

        devShells = {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              air
              gopls
              go-tools
              jq
            ];
          };
        };
      }
    );
}
