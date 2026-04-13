self:
{
  config,
  lib,
  pkgs,
  ...
}:

let
  cfg = config.services.waygent;
in
{
  options.services.waygent = {
    enable = lib.mkEnableOption "waygent - Wayland desktop automation agent";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.system}.default;
      description = "The waygent package to use.";
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];

    # ydotool daemon provides input injection capability on Wayland
    services.ydotool.enable = true;
  };
}
