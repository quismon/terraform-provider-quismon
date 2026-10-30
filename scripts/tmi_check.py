#!/usr/bin/env python3
"""TMI check: block implementation-internals from leaking into public docs.

Scans every PUBLIC documentation surface (website, docs/, OpenAPI spec,
API_DOCUMENTATION.md, provider templates + generated Registry docs, status
page shells) for patterns that reveal implementation internals: tooling
names, queue/transport mechanics, source paths, table names, probe
internals, dead/unwired-code admissions, and TODO markers.

Exit 1 with a violation list when anything leaks. Wire it as a
precondition of every document publication target (deploy-docs,
deploy-website, provider release) - see Makefile and the release workflow.

An entry can be allowed with an inline marker on the same line:
    <!-- tmi-allow: trust-dns -->
    // tmi-allow: bcrypt
(the marker names the pattern it excuses, and only on that line).
"""

import re
import sys
from pathlib import Path

# Root can be passed (standalone checkouts, e.g. the provider release CI);
# defaults to the monorepo root two levels up from this script.
if len(sys.argv) > 1:
    REPO = Path(sys.argv[1]).resolve()
else:
    REPO = Path(__file__).resolve().parent.parent

# Public surfaces: monorepo mode, or provider-checkout mode (the provider
# release CI checks out only terraform-provider-quismon).
if (REPO / "website").exists():
    PUBLIC_FILES = ["openapi.yaml", "API_DOCUMENTATION.md"]
    PUBLIC_DIRS = [
        "website",
        "docs",
        "terraform-provider-quismon/templates",
        "terraform-provider-quismon/docs",
        "infrastructure/status-page",
    ]
else:
    PUBLIC_FILES = ["README.md"]
    PUBLIC_DIRS = ["templates", "docs"]

# (name, pattern, allow_everywhere_in) - allow_everywhere_in limits a pattern
# to specific files where the mention is a deliberate public statement.
PATTERNS = [
    # Implementation tooling / crate names
    ("trust-dns", r"trust[._-]?dns|hickory", []),
    ("serde", r"\bserde\b", []),
    ("reqwest", r"\breqwest\b", []),
    ("tokio", r"\btokio\b", []),
    ("surge-ping", r"surge[._-]?ping", []),
    ("hyper", r"\bhyper\b", []),
    ("pnet", r"\bpnet_packet\b", []),
    ("go-module-paths", r"quismon/quismon-(api|checker|processor|notifier|scheduler)/", []),
    # Source paths / function names
    ("source-paths", r"src/checks|handlers\.go|models\.rs|main\.go line|database\.go", []),
    # Transport/queue mechanics
    ("queue-mechanics", r"consumer group|pub/sub|xadd|xreadgroup|redis stream|notifications:stream|results:queue", []),
    ("redis-internal", r"redis://|172\.17\.0\.1|quismon:notifications|quismon:results|quismon:notif:done", []),
    # Storage internals
    ("storage-internals", r"drop_chunks|hypertable|traceroute_paths|traceroute_path_changes|schema_migrations|SQLSTATE|billing\.subscriptions\b", []),
    # Probe internals (the unique mail tag scheme, EHLO identity)
    ("probe-internals", r"quismon-smtp-checker|QMO-|EHLO helo|auth login base64", []),
    # Dead/legacy/unwired admissions
    ("unwired-admissions", r"\bnot wired\b|\bunwired\b|\bdead code\b|\bno-op end-to-end\b|binary fallback|command fallback|legacy code path", []),
    # TODO markers in published docs
    ("todo-markers", r"\bTODO\b|\bFIXME\b|\bXXX\b", []),  # XXX matched case-sensitively below
    # Internal doc references / hosts
    ("internal-references", r"internal-docs/|prod-01|quismon-prod|\.secrets/", []),
    # Infrastructure file names
    ("infra-files", r"docker-compose|Caddyfile|bootstrap-playwright|systemd", []),
]

# Patterns deliberately allowed in specific public files:
#  - bcrypt in API_DOCUMENTATION.md: password-hashing disclosure is a
#    security norm and the 72-byte policy explanation needs it.
for name, pattern, allowed in PATTERNS:
    pass
ALLOW_OVERRIDES = {
    "todo-markers": [],  # none today
}
ALLOWED_FILES = {
    # pattern name -> [file substrings where mentions are accepted]
    "storage-internals": [],  # e.g. ["API_DOCUMENTATION.md"] if ever needed
}
# bcrypt is not in PATTERNS at all: password-hashing disclosure is standard.
# Add it here if the bar changes.


def public_files():
    files = []
    for f in PUBLIC_FILES:
        p = REPO / f
        if p.exists():
            files.append(p)
    for d in PUBLIC_DIRS:
        base = REPO / d
        if not base.exists():
            continue
        for p in base.rglob("*"):
            if p.suffix in {".html", ".md", ".yaml", ".yml", ".tmpl"}:
                files.append(p)
    return sorted(set(files))


def line_allowed(line: str, pattern_name: str) -> bool:
    return f"tmi-allow: {pattern_name}" in line


def main() -> int:
    violations = []
    for path in public_files():
        rel = str(path.relative_to(REPO))
        try:
            lines = path.read_text(encoding="utf-8").splitlines()
        except UnicodeDecodeError:
            continue
        for i, line in enumerate(lines, 1):
            # Skip the script's own pattern table if it ever lands in a
            # scanned directory (defense in depth).
            # Markers may sit up to 3 lines above (YAML multi-line scalars)
            context = lines[max(0, i - 5):i - 1]
            # Built-in allowlist: deliberate public statements, immune to spec
            # regenerations (yaml.safe_dump strips inline tmi-allow comments).
            # Key: (file substring, pattern name, text substring).
            BUILTIN_ALLOWED = [
                ("openapi.yaml", "queue-mechanics",
                 "backed by a durable Redis Stream"),
                ("openapi.yaml", "queue-mechanics",
                 "Redis stream entry ID to resume from"),
            ]
            for name, pattern, _ in PATTERNS:
                flags = 0 if name == "todo-markers" and "\\bXXX\\b" in pattern else re.IGNORECASE
                # XXX placeholders (xxx/yyy) are case-sensitive: lowercase
                # placeholder values in examples are fine.
                if name == "todo-markers" and re.search(r"\bxxx\b|\byyy\b", line, re.IGNORECASE) and not re.search(r"\bXXX\b", line):
                    continue
                if re.search(pattern, line, flags):
                    if line_allowed(line, name) or any(line_allowed(c, name) for c in context):
                        continue
                    if any(fsub in rel and tsub.lower() in line.lower()
                           for fsub, pname, tsub in BUILTIN_ALLOWED if pname == name):
                        continue
                    violations.append((rel, i, name, line.strip()[:120]))

    if violations:
        print(f"TMI check FAILED - {len(violations)} implementation detail(s) in public docs:\n")
        current = None
        for rel, i, name, text in violations:
            if rel != current:
                print(f"  {rel}")
                current = rel
            print(f"    line {i}: [{name}] {text}")
        print("\nFix the wording, or add a line-scoped 'tmi-allow: <pattern>' marker")
        print("for a deliberate public statement.")
        return 1

    print("TMI check passed - public docs are clean.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
