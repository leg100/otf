#!/bin/bash
# Script to migrate web handlers to ui package with proper naming to avoid conflicts

set -e

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "Usage: $0 <service-name>"
    echo "Example: $0 workspace"
    exit 1
fi

SERVICE_DIR="internal/$SERVICE"
UI_DIR="internal/ui"
SERVICE_GO="$UI_DIR/${SERVICE}.go"
SERVICE_TEMPL="$UI_DIR/${SERVICE}_view.templ"
SERVICE_HELPERS="$UI_DIR/${SERVICE}_helpers.go"

echo "==> Migrating $SERVICE handlers to ui package"

# Step 1: Rename all generic function/type names with service prefix in .go file
if [ -f "$SERVICE_GO" ]; then
    echo "  - Renaming functions and types in $SERVICE_GO"

    # Rename types (avoid double prefixing)
    sed -i "/^type ${SERVICE}ListProps/!s/^type listProps/type ${SERVICE}ListProps/g" "$SERVICE_GO"
    sed -i "/^type ${SERVICE}GetProps/!s/^type getProps/type ${SERVICE}GetProps/g" "$SERVICE_GO"
    sed -i "/^type ${SERVICE}EditProps/!s/^type editProps/type ${SERVICE}EditProps/g" "$SERVICE_GO"
    sed -i "/^type ${SERVICE}NewProps/!s/^type newProps/type ${SERVICE}NewProps/g" "$SERVICE_GO"
    sed -i "/^type ${SERVICE}Table/!s/^type table struct/type ${SERVICE}Table struct/g" "$SERVICE_GO"
    sed -i "/^type ${SERVICE}Perm/!s/^type perm struct/type ${SERVICE}Perm struct/g" "$SERVICE_GO"

    # Rename functions
    sed -i "s/^func list(/func ${SERVICE}List(/g" "$SERVICE_GO"
    sed -i "s/^func get(/func ${SERVICE}Get(/g" "$SERVICE_GO"
    sed -i "s/^func new(/func ${SERVICE}New(/g" "$SERVICE_GO"
    sed -i "s/^func edit(/func ${SERVICE}Edit(/g" "$SERVICE_GO"
    sed -i "s/^func listActions(/func ${SERVICE}ListActions(/g" "$SERVICE_GO"

    # Fix references to renamed items
    sed -i "s/\\blistProps\\b/${SERVICE}ListProps/g" "$SERVICE_GO"
    sed -i "s/\\bgetProps\\b/${SERVICE}GetProps/g" "$SERVICE_GO"
    sed -i "s/\\beditProps\\b/${SERVICE}EditProps/g" "$SERVICE_GO"
    sed -i "s/\\bnewProps\\b/${SERVICE}NewProps/g" "$SERVICE_GO"
    sed -i "s/\\bperm\\b/${SERVICE}Perm/g" "$SERVICE_GO"

    # Fix function calls
    sed -i "s/html\\.Render(list(/html.Render(${SERVICE}List(/g" "$SERVICE_GO"
    sed -i "s/html\\.Render(get(/html.Render(${SERVICE}Get(/g" "$SERVICE_GO"
    sed -i "s/html\\.Render(new(/html.Render(${SERVICE}New(/g" "$SERVICE_GO"
    sed -i "s/html\\.Render(edit(/html.Render(${SERVICE}Edit(/g" "$SERVICE_GO"

    # Fix table references
    sed -i "s/&table{/${SERVICE}Table{/g" "$SERVICE_GO"
    sed -i "s/\\*table\\*/${SERVICE}Table/g" "$SERVICE_GO"

    echo "  ✓ Renamed functions and types"
fi

# Step 2: Update templ file if it exists
if [ -f "$SERVICE_TEMPL" ]; then
    echo "  - Updating $SERVICE_TEMPL"

    # Ensure package is ui
    sed -i "1s/^package .*/package ui/" "$SERVICE_TEMPL"

    # Add import for the service package if not present
    if ! grep -q "github.com/leg100/otf/internal/$SERVICE" "$SERVICE_TEMPL"; then
        # Find the import block and add the service import
        sed -i "/^import (/a\\    \"github.com/leg100/otf/internal/$SERVICE\"" "$SERVICE_TEMPL"
    fi

    # Rename templ components with service prefix
    sed -i "s/^templ list(/templ ${SERVICE}List(/g" "$SERVICE_TEMPL"
    sed -i "s/^templ get(/templ ${SERVICE}Get(/g" "$SERVICE_TEMPL"
    sed -i "s/^templ new(/templ ${SERVICE}New(/g" "$SERVICE_TEMPL"
    sed -i "s/^templ edit(/templ ${SERVICE}Edit(/g" "$SERVICE_TEMPL"
    sed -i "s/^templ listActions(/templ ${SERVICE}ListActions(/g" "$SERVICE_TEMPL"

    # Update type references in templ parameters
    sed -i "s/(props listProps)/(props ${SERVICE}ListProps)/g" "$SERVICE_TEMPL"
    sed -i "s/(props getProps)/(props ${SERVICE}GetProps)/g" "$SERVICE_TEMPL"
    sed -i "s/(props editProps)/(props ${SERVICE}EditProps)/g" "$SERVICE_TEMPL"
    sed -i "s/(props newProps)/(props ${SERVICE}NewProps)/g" "$SERVICE_TEMPL"

    # Also update type declarations in templ
    sed -i "/^type ${SERVICE}ListProps/!s/^type listProps/type ${SERVICE}ListProps/g" "$SERVICE_TEMPL"
    sed -i "/^type ${SERVICE}GetProps/!s/^type getProps/type ${SERVICE}GetProps/g" "$SERVICE_TEMPL"
    sed -i "/^type ${SERVICE}EditProps/!s/^type editProps/type ${SERVICE}EditProps/g" "$SERVICE_TEMPL"
    sed -i "/^type ${SERVICE}Table/!s/^type table/type ${SERVICE}Table/g" "$SERVICE_TEMPL"
    sed -i "/^type ${SERVICE}Perm/!s/^type perm/type ${SERVICE}Perm/g" "$SERVICE_TEMPL"

    echo "  ✓ Updated template file"

    # Regenerate template
    echo "  - Regenerating template..."
    go tool templ generate "$SERVICE_TEMPL" 2>&1 | grep -v "^$" || true
    echo "  ✓ Template regenerated"
fi

# Step 3: Update helpers file if it exists
if [ -f "$SERVICE_HELPERS" ]; then
    echo "  - Checking helpers file"
    # Ensure package is ui
    sed -i "1s/^package .*/package ui/" "$SERVICE_HELPERS"
    echo "  ✓ Updated helpers file"
fi

echo "==> Migration complete for $SERVICE"
echo ""
echo "Next steps:"
echo "1. Manually review and fix any remaining type references"
echo "2. Update internal/$SERVICE/service.go to remove web handler registration"
echo "3. Test: go build ./internal/ui"
