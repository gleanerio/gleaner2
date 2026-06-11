#!/bin/bash
# Smoke test for the glcon/nabu CLI.
# Verifies that every subcommand can print help without crashing.
# No config or services required — tests CLI wiring only.
#
# Usage: ./tools/smoke_test.sh [path/to/binary]

set -euo pipefail

BINARY="${1:-./nabu}"

if [ ! -x "$BINARY" ]; then
    echo "Binary not found at $BINARY — building..."
    (cd "$(dirname "$0")/.." && go build -o nabu ./cmd/nabu/) || {
        echo "FAIL: build failed"
        exit 1
    }
    BINARY="$(dirname "$0")/../nabu"
fi

PASS=0
FAIL=0
ERRORS=""

run_test() {
    local desc="$1"
    shift
    if "$@" >/dev/null 2>&1; then
        PASS=$((PASS + 1))
        echo "  PASS  $desc"
    else
        FAIL=$((FAIL + 1))
        ERRORS="${ERRORS}\n  FAIL  $desc ($*)"
        echo "  FAIL  $desc"
    fi
}

echo "=== CLI Smoke Tests ==="
echo ""

# Root help
run_test "nabu --help" "$BINARY" --help

# Every subcommand's --help
run_test "prefix --help" "$BINARY" prefix --help
run_test "bulk --help" "$BINARY" bulk --help
run_test "release --help" "$BINARY" release --help
run_test "prune --help" "$BINARY" prune --help
run_test "object --help" "$BINARY" object --help

# Graph subcommands
run_test "graph --help" "$BINARY" graph --help
run_test "graph clear --help" "$BINARY" graph clear --help
run_test "graph drop --help" "$BINARY" graph drop --help

# Config subcommands
run_test "config --help" "$BINARY" config --help
run_test "config init --help" "$BINARY" config init --help

# Version flag (if implemented)
if "$BINARY" --help 2>&1 | grep -q -- '--version'; then
    run_test "nabu --version" "$BINARY" --version
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="

if [ $FAIL -gt 0 ]; then
    echo -e "\nFailures:$ERRORS"
    exit 1
fi
