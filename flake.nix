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
      in rec
      {
        packages = rec {
          default = suslik;
          suslik = pkgs.buildGo126Module {
            pname = "suslik";
            version = "0.1.0";
            src = ./.;
            vendorHash = goVendorHash;
          };
          suslik-image = pkgs.dockerTools.buildLayeredImage {
            name = "suslik";
            tag = "latest";
            contents = [
              suslik
              pkgs.dockerTools.caCertificates
              pkgs.busybox
            ];
            config = {
              Cmd = [ "${suslik}/bin/suslik" ];
              Env = [
                "PATH=/bin"
              ];
            };
          };
        };

        checks = {
          x86_64-linux = packages.suslik;
          x86_64-linux-image = packages.suslik-image;
        };

        devShells = {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_26
              air
              gopls
              go-tools
              jq
              act
            ];
          };
        };
      }
    );
}
