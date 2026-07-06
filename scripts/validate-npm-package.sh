#!/usr/bin/env bash
set -euo pipefail

package_dir="npm"
version=""
package_name="@delicious233/dida-cli"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      if [[ $# -lt 2 ]]; then
        echo "--version requires a value" >&2
        exit 1
      fi
      version="${2#v}"
      shift 2
      ;;
    --package-dir)
      if [[ $# -lt 2 ]]; then
        echo "--package-dir requires a value" >&2
        exit 1
      fi
      package_dir="$2"
      shift 2
      ;;
    --package-name)
      if [[ $# -lt 2 ]]; then
        echo "--package-name requires a value" >&2
        exit 1
      fi
      package_name="$2"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$version" ]]; then
  echo "npm package version is required" >&2
  exit 1
fi

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "npm package version must use X.Y.Z; got $version" >&2
  exit 1
fi

if [[ ! -d "$package_dir" ]]; then
  echo "npm package directory not found: $package_dir" >&2
  exit 1
fi

(
  cd "$package_dir"
  npm version "$version" --no-git-tag-version --allow-same-version >/dev/null
  pack_json="$(npm pack --dry-run --json)"
  echo "$pack_json"
  echo "$pack_json" | EXPECTED_NAME="$package_name" EXPECTED_VERSION="$version" node -e '
    const fs = require("fs");
    const data = JSON.parse(fs.readFileSync(0, "utf8"));
    const pkg = data[0];
    if (!pkg) {
      console.error("npm pack did not return package metadata");
      process.exit(1);
    }
    if (pkg.name !== process.env.EXPECTED_NAME) {
      console.error(`npm package name mismatch: got ${pkg.name}, want ${process.env.EXPECTED_NAME}`);
      process.exit(1);
    }
    if (pkg.version !== process.env.EXPECTED_VERSION) {
      console.error(`npm package version mismatch: got ${pkg.version}, want ${process.env.EXPECTED_VERSION}`);
      process.exit(1);
    }
    const files = new Set(pkg.files.map((file) => file.path));
    for (const path of ["bin/dida", "scripts/install.js", "package.json"]) {
      if (!files.has(path)) {
        console.error(`npm package is missing ${path}`);
        process.exit(1);
      }
    }
  '
)

echo "npm package metadata valid for ${package_name}@${version}"
