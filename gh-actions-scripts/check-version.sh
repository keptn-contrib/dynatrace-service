#!/usr/bin/env bash

function printUsage() {
    echo "Usage: $0 <version-number> <pre-release | release-3 | release-2> >"
    echo "  example: $0 1.2.3-next-0 pre-release"
    echo "  example: $0 1.2.3 release-3"
    echo "  example: $0 1.234 release-2"
    exit 1
}

function checkPreReleaseVersion() {
  if [[ "$1" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))(\+([0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*))?$ ]]; then
    exit 0
  else
    echo "not a valid pre-release version number"
    exit 2
  fi
}

function checkMajorMinorReleaseVersion() {
  if [[ "$1" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
    exit 0
  else
    echo "not a valid major.minor release version number"
    exit 3
  fi
}

function checkFullReleaseVersion() {
  if [[ "$1" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)?$ ]]; then
    exit 0
  else
    echo "not a valid major.minor.patch release version number"
    exit 3
  fi
}

### start

if [ $# -ne 2 ]; then
    printUsage
fi

VERSION="$1"
if [ -z "$1" ] || [ -z "$2" ]; then
  printUsage
fi

if [ "$2" = "pre-release" ]; then
  checkPreReleaseVersion "$VERSION"
elif [ "$2" = "release-3" ]; then
  checkFullReleaseVersion "$VERSION"
elif [ "$2" = "release-2" ]; then
  checkMajorMinorReleaseVersion "$VERSION"
else
  echo "only 'pre-release', 'release-3' or 'release-2' are allowed"
  printUsage
fi
