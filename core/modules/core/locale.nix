# Locale, time, and console defaults. All overridable per host.
{ lib, ... }:
{
  time.timeZone = lib.mkDefault "UTC";
  i18n.defaultLocale = lib.mkDefault "en_US.UTF-8";
  console.keyMap = lib.mkDefault "us";
}
