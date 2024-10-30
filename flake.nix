{
  description = "XOXO: A basic tic-tac-toe player";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05-small";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
    }:
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        myapp = pkgs.callPackage ./. {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };
      in
      {
        apps.default = {
          type = "app";
          program = "${myapp}/bin/xoxo";
        };

        packages.default = myapp;
        packages.dockerImage = pkgs.dockerTools.buildLayeredImage {
          name = "xoxo";
          tag = "latest";
          created = "now";
          config.Entrypoint = [
            "${myapp}/bin/xoxo"
          ];
        };

        devShells.default = pkgs.callPackage ./shell.nix {
          inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
        };
      }
    ));
}
