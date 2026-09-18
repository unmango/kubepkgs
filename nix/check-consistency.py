"""Validate packages.json against itself, the tree, and the docs.

Run standalone from the repo root:

    python3 nix/check-consistency.py .
"""

import json
import os
import re
import sys

FAKE_VENDOR_HASH = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

ROW = re.compile(r"^\|.*\|$")


def rosters(data):
    """Every tracked package paired with the roster, and directory, it lives in."""
    for roster in ("sigs", "deps"):
        for sig in data.get(roster, []):
            yield roster, sig


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

    seen = {}
    for roster, sig in rosters(data):
        name = sig["name"]

        # The CLI's --target and --sig flags are a flat namespace, so a name
        # reused across rosters would make them ambiguous.
        if name in seen:
            fail(f"{name} is declared in both {seen[name]} and {roster}")
        seen[name] = roster

        definition = os.path.join(root, roster, sig["path"], "default.nix")
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


def find_tables(lines):
    """The [start, end] line index of each contiguous run of table rows."""
    tables = []
    i = 0
    while i < len(lines):
        if not ROW.match(lines[i]):
            i += 1
            continue
        j = i
        while j + 1 < len(lines) and ROW.match(lines[j + 1]):
            j += 1
        # A header, a separator and at least one data row.
        if j - i + 1 >= 3:
            tables.append((i, j))
        i = j + 1
    return tables


def cells(row):
    return [c.strip() for c in row.strip("|").split("|")]


def check_table(data, lines, bounds, roster, packages, fail):
    """One roster's table states the same pins as packages.json."""
    first, last = bounds
    rows = lines[first:last + 1]

    # GFM keeps consuming rows until a blank line, so prose directly under the
    # last row renders as a final row of the table. That is invisible to the
    # row regex, which only matches pipe-delimited lines.
    if last + 1 < len(lines) and lines[last + 1].strip():
        fail(f"README {roster} table is followed by {lines[last + 1].strip()[:40]!r} with no "
             "blank line, which renders as a trailing table row")

    header = cells(rows[0])[1:]
    expected_header = [sig["name"] for sig in packages]
    if header != expected_header:
        fail(f"README {roster} table columns {header} do not match packages.json {expected_header}")
        return

    documented = {}
    for row in rows[2:]:
        values = cells(row)
        match = re.match(r"\*\*(\S+?)\*\*(?: \(latest\))?$", values[0])
        if not match:
            fail(f"README {roster} table row {values[0]!r} is not formatted as **<minor>** [(latest)]")
            continue
        documented[match.group(1)] = (values[0], values[1:])

    if sorted(documented) != sorted(data["supported"]):
        fail(f"README {roster} table rows {sorted(documented)} do not match "
             f"supported {sorted(data['supported'])}")

    for minor, (label, values) in documented.items():
        if minor not in data["kubernetes"]:
            continue
        marked_latest = "(latest)" in label
        if marked_latest != (minor == data["latest"]):
            fail(f"README {roster} table marks {minor} as latest" if marked_latest
                 else f"README {roster} table does not mark {minor} as latest")
        for sig, documented_version in zip(packages, values):
            pinned = sig["minors"][minor]
            series = ".".join(pinned.split(".")[:2])
            if documented_version != series:
                fail(f"README says {sig['name']} {documented_version} for {minor}, "
                     f"packages.json pins {pinned}")


def check_readme(data, root, fail):
    """One README table per roster, each stating the same pins as packages.json."""
    lines = open(os.path.join(root, "README.md")).read().splitlines()

    expected = [("sigs", data["sigs"])]
    if data.get("deps"):
        expected.append(("deps", data["deps"]))

    tables = find_tables(lines)
    if len(tables) < len(expected):
        fail(f"README has {len(tables)} version table(s), expected {len(expected)}")
        return

    for bounds, (roster, packages) in zip(tables, expected):
        check_table(data, lines, bounds, roster, packages, fail)


# The core roster lives in core/default.nix rather than in packages.json,
# because core binaries carry their build configuration in Nix. CORE_ATTR
# matches a top-level attribute of the result set; everything nested inside one
# is indented further. The `let` bindings share that indentation, so only the
# part after the top-level `in` is read.
CORE_ATTR = re.compile(r"^  ([A-Za-z][\w-]*) = ", re.M)

# A backticked package name in an inventory list, ignoring any parenthesised
# note beside it.
ENTRY = re.compile(r"`([^`]+)`")

INVENTORIES = [
    ("### Available core packages", ""),
    ("### Available dependency packages", "deps."),
    ("### Available SIG packages", "sigs."),
]


def core_packages(root, fail):
    text = open(os.path.join(root, "core", "default.nix")).read()
    _, marker, body = text.rpartition("\nin\n")
    if not marker:
        fail("core/default.nix has no top-level `in`")
        return []
    names = CORE_ATTR.findall(body)
    if not names:
        fail("core/default.nix declares no packages")
    return names


def inventory(lines, heading, fail):
    """The packages listed in the paragraph under heading, or None."""
    for i, line in enumerate(lines):
        if line.strip() != heading:
            continue
        for candidate in lines[i + 1:]:
            if not candidate.strip():
                continue
            if candidate.startswith("#"):
                break
            return ENTRY.findall(candidate)
        fail(f"README {heading!r} has no package list")
        return None
    fail(f"README has no {heading!r} heading")
    return None


def check_inventories(data, root, fail):
    """Each roster's packages are all listed under its README heading."""
    lines = open(os.path.join(root, "README.md")).read().splitlines()

    rosters = [
        core_packages(root, fail),
        [sig["name"] for sig in data.get("deps", [])],
        [sig["name"] for sig in data.get("sigs", [])],
    ]

    for (heading, prefix), roster in zip(INVENTORIES, rosters):
        listed = inventory(lines, heading, fail)
        if listed is None:
            continue
        expected = {prefix + name for name in roster}
        for missing in sorted(expected - set(listed)):
            fail(f"README {heading!r} does not list {missing}; run sync-docs")
        for extra in sorted(set(listed) - expected):
            fail(f"README {heading!r} lists {extra}, which no roster declares")


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

    for _, sig in rosters(data):
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
    check_inventories(data, root, fail)
    check_docs(data, root, fail)
    check_go_pins(data, fail)

    if errors:
        for error in errors:
            print(f"packages.json: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1] if len(sys.argv) > 1 else "."))
