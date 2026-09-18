{
  pkgs,
  ...
}:
{
  languages.go = {
    enable = true;
    package = pkgs.go_latest;
  };
  packages = [ pkgs.gotestsum ];
}
