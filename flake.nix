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
        suslik = pkgs.buildGo126Module {
          pname = "suslik";
          version = "0.1.0";
          src = ./.;
          vendorHash = goVendorHash;
        };
      in
      {
        packages = {
          default = suslik;
          inherit suslik;
        } // pkgs.lib.optionalAttrs pkgs.stdenv.isLinux {
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
          inherit suslik;
        } // pkgs.lib.optionalAttrs pkgs.stdenv.isLinux {
          suslik-image = self.packages.${system}.suslik-image;
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
