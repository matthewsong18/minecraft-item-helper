{
  pkgs,
  lib,
  config,
  ...
}:
{
  languages.go.enable = true;
  packages = [ pkgs.gotestsum ];
}
