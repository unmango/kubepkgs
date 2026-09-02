"""Validate packages.json against itself, the tree, and the docs.

Run standalone from the repo root:

    python3 nix/check-consistency.py .
"""

import json
import os
import re
import sys

FAKE_VENDOR_HASH = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="


def check_structure(data, root, fail):
    supported = data["supported"]

    if not supported:
        fail("supported is empty")
    if data["latest"] not in supported:
        fail(f"latest {data['latest']!r} is not in supported")

    for minor in supported:
        entry = data["kubernetes"].get(minor)
        if entry is None:
            fail(f"kubernetes has no entry for supported minor {minor}")
            continue
        for field in ("version", "srcHash", "commit"):
            if not entry.get(field):
                fail(f"kubernetes {minor} has an empty {field}")

    for extra in set(data["kubernetes"]) - set(supported):
        fail(f"kubernetes {extra} is not a supported minor")

    for sig in data["sigs"]:
        name = sig["name"]

        definition = os.path.join(root, "sigs", sig["path"], "default.nix")
        if not os.path.isfile(definition):
            fail(f"{name} path {sig['path']!r} has no default.nix")

        for minor in supported:
            if minor not in sig["minors"]:
                fail(f"{name} pins no version for supported minor {minor}")
        for extra in set(sig["minors"]) - set(supported):
            fail(f"{name} pins a version for unsupported minor {extra}")

        pinned = set(sig["minors"].values())
        for version in pinned:
            record = sig["versions"].get(version)
            if record is None:
                fail(f"{name} {version} is pinned but has no versions record")
                continue
            for field in ("commit", "srcHash", "vendorHash"):
                if not record.get(field):
                    fail(f"{name} {version} has an empty {field}")
            if record.get("vendorHash") == FAKE_VENDOR_HASH:
                fail(f"{name} {version} still has the placeholder vendorHash; run vendor-hashes")

        for orphan in set(sig["versions"]) - pinned:
            fail(f"{name} {orphan} has a versions record but is pinned by no minor")


def check_readme(data, root, fail):
    """The supported-versions table in the README states the same pins."""
    text = open(os.path.join(root, "README.md")).read()

    rows = re.findall(r"^\|.*\|$", text, re.MULTILINE)
    if len(rows) < 3:
        fail("README has no supported-versions table")
        return

    def cells(row):
        return [c.strip() for c in row.strip("|").split("|")]

    header = cells(rows[0])[1:]
    expected_header = [sig["name"] for sig in data["sigs"]]
    if header != expected_header:
        fail(f"README table columns {header} do not match packages.json sigs {expected_header}")
        return

    documented = {}
    for row in rows[2:]:
        values = cells(row)
        match = re.match(r"\*\*(\S+?)\*\*(?: \(latest\))?$", values[0])
        if not match:
            fail(f"README table row {values[0]!r} is not formatted as **<minor>** [(latest)]")
            continue
        documented[match.group(1)] = (values[0], values[1:])

    if sorted(documented) != sorted(data["supported"]):
        fail(f"README table rows {sorted(documented)} do not match supported {sorted(data['supported'])}")

    for minor, (label, values) in documented.items():
        if minor not in data["kubernetes"]:
            continue
        marked_latest = "(latest)" in label
        if marked_latest != (minor == data["latest"]):
            fail(f"README marks {minor} as latest" if marked_latest
                 else f"README does not mark {minor} as latest")
        for sig, documented_version in zip(data["sigs"], values):
            pinned = sig["minors"][minor]
            series = ".".join(pinned.split(".")[:2])
            if documented_version != series:
                fail(f"README says {sig['name']} {documented_version} for {minor}, "
                     f"packages.json pins {pinned}")


# The two ways the docs name a Kubernetes minor: an attribute path and a flake
# check name. Both stop resolving once that minor leaves the supported window.
EXAMPLE_PATTERNS = [
    r'kubernetes\."(\d+\.\d+)"',
    r"core-(\d+\.\d+)-",
]

EXAMPLE_DOCS = [
    "README.md",
    "CLAUDE.md",
    ".github/copilot-instructions.md",
]


GO_PIN = re.compile(r"^\d+\.\d+$")


def check_go_pins(data, fail):
    """An optional go pin names a Go minor series, not a full version."""
    for minor, entry in data["kubernetes"].items():
        pinned = entry.get("go")
        if pinned is not None and not GO_PIN.match(pinned):
            fail(f"kubernetes {minor} pins go {pinned!r}, expected a minor series like \"1.26\"")

    for sig in data["sigs"]:
        pinned = sig.get("go")
        if pinned is not None and not GO_PIN.match(pinned):
            fail(f"{sig['name']} pins go {pinned!r}, expected a minor series like \"1.26\"")


def check_docs(data, root, fail):
    """Version-pinned examples in the docs name a minor that still exists."""
    supported = set(data["supported"])

    for name in EXAMPLE_DOCS:
        path = os.path.join(root, name)
        if not os.path.exists(path):
            fail(f"{name} is missing")
            continue
        text = open(path).read()
        for pattern in EXAMPLE_PATTERNS:
            for minor in sorted(set(re.findall(pattern, text))):
                if minor not in supported:
                    fail(f"{name} has an example for unsupported minor {minor}")


def main(root):
    data = json.load(open(os.path.join(root, "packages.json")))

    errors = []
    fail = errors.append

    check_structure(data, root, fail)
    check_readme(data, root, fail)
    check_docs(data, root, fail)
    check_go_pins(data, fail)

    if errors:
        for error in errors:
            print(f"packages.json: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1] if len(sys.argv) > 1 else "."))
