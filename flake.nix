{
  description = "Wayland desktop automation agent powered by visual LLMs";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      pkgsFor = system: import nixpkgs { inherit system; };
    in
    {
      packages = forAllSystems (system: rec {
        waygent =
          let
            pkgs = pkgsFor system;
          in
          pkgs.buildGoModule {
            pname = "waygent";
            version = "0.1.0";
            src = self;

            vendorHash = null;

            subPackages = [ "cmd/waygent" ];

            ldflags = [
              "-s"
              "-w"
            ];

            nativeBuildInputs = [ pkgs.makeWrapper ];

            postInstall = ''
              wrapProgram $out/bin/waygent \
                --prefix PATH : ${
                  pkgs.lib.makeBinPath [
                    pkgs.ydotool
                    pkgs.glib
                    pkgs.grim
                  ]
                }
            '';

            meta = with pkgs.lib; {
              description = "Wayland desktop automation agent powered by visual LLMs";
              homepage = "https://github.com/waygent/waygent";
              license = licenses.mit;
              platforms = platforms.linux;
              mainProgram = "waygent";
            };
          };
        default = waygent;
      });

      devShells = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              gotools
              ydotool
              glib
              grim
            ];

            shellHook = ''
              echo "waygent dev shell"
              echo "  Go: $(go version)"
            '';
          };
        }
      );

      nixosModules.default = import ./nix/module.nix self;
    };
}
