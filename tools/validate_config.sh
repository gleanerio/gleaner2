#!/bin/bash
# Validate a glcon configuration directory.
# Checks that config files parse correctly and contain required sections.
#
# Usage: ./tools/validate_config.sh configs/local
#        ./tools/validate_config.sh configs/local services.yaml sources.yaml glcon.yaml
#
# Exit codes:
#   0 = all checks pass
#   1 = validation errors found

set -euo pipefail

CONFIG_DIR="${1:?Usage: validate_config.sh <config-dir> [files...]}"
shift

# If specific files given, check those; otherwise check expected set
if [ $# -gt 0 ]; then
    FILES=("$@")
else
    FILES=(services.yaml sources.yaml glcon.yaml nabu.yaml)
fi

PASS=0
WARN=0
FAIL=0

check_file() {
    local f="$1"
    local path="$CONFIG_DIR/$f"

    if [ ! -f "$path" ]; then
        echo "  SKIP  $f (not found)"
        return
    fi

    # Basic YAML syntax check (python available on most systems)
    if command -v python3 &>/dev/null; then
        if python3 -c "import yaml; yaml.safe_load(open('$path'))" 2>/dev/null; then
            PASS=$((PASS + 1))
            echo "  PASS  $f — valid YAML"
        else
            FAIL=$((FAIL + 1))
            echo "  FAIL  $f — invalid YAML syntax"
            python3 -c "import yaml; yaml.safe_load(open('$path'))" 2>&1 | head -5
            return
        fi
    else
        # Fallback: just check it's non-empty
        if [ -s "$path" ]; then
            PASS=$((PASS + 1))
            echo "  PASS  $f — file exists and is non-empty"
        else
            FAIL=$((FAIL + 1))
            echo "  FAIL  $f — file is empty"
            return
        fi
    fi

    # Content-specific checks
    case "$f" in
        services.yaml)
            check_key "$path" "minio" "MinIO configuration"
            check_key "$path" "endpoints" "SPARQL endpoint configuration"
            check_secret_placeholder "$path"
            ;;
        sources.yaml)
            check_key "$path" "sources" "data source definitions"
            ;;
        glcon.yaml|nabu.yaml)
            check_key "$path" "objects" "object prefix configuration"
            ;;
    esac
}

check_key() {
    local path="$1"
    local key="$2"
    local desc="$3"

    if grep -q "^${key}:" "$path" 2>/dev/null; then
        PASS=$((PASS + 1))
        echo "  PASS  $(basename "$path"): has $desc ($key:)"
    else
        WARN=$((WARN + 1))
        echo "  WARN  $(basename "$path"): missing $desc ($key:)"
    fi
}

check_secret_placeholder() {
    local path="$1"

    # Warn if accessKey/secretKey look like they have placeholder values
    if grep -qE '(accesskey|secretkey):\s*""' "$path" 2>/dev/null; then
        WARN=$((WARN + 1))
        echo "  WARN  $(basename "$path"): credentials are empty — set via config or MINIO_ACCESS_KEY/MINIO_SECRET_KEY env vars"
    fi
}

echo "=== Config Validation: $CONFIG_DIR ==="
echo ""

for f in "${FILES[@]}"; do
    check_file "$f"
done

echo ""
echo "=== Results: $PASS passed, $WARN warnings, $FAIL failed ==="

if [ $FAIL -gt 0 ]; then
    exit 1
fi
