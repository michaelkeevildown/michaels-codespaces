#!/bin/bash

# Test script for component selection

echo "Testing component selection system..."
echo ""

# Source required modules
source scripts/modules/components/registry.sh
source scripts/modules/components/simple-selector.sh

echo "=== Testing Component Registry ==="
echo ""
echo "Registered components:"
register_components
for data in "${COMPONENT_DATA[@]}"; do
    id=$(echo "$data" | cut -d'|' -f1)
    name=$(echo "$data" | cut -d'|' -f2)
    desc=$(echo "$data" | cut -d'|' -f3)
    echo "  - $id: $name - $desc"
done

echo ""
echo "=== Testing Simple Selector ==="
echo ""
echo "Running simple_select (type 'a' to select all)..."
selected=$(simple_select)
echo ""
echo "Selected components: $selected"

echo ""
echo "=== Testing Component Dependencies ==="
echo ""
for component in $selected; do
    deps=$(get_component_dependencies "$component")
    if [ -n "$deps" ]; then
        echo "$component depends on: $deps"
    else
        echo "$component has no dependencies"
    fi
done

echo ""
echo "=== Testing Installation Order ==="
echo ""
order=$(get_install_order $selected)
echo "Installation order (with dependencies resolved):"
for comp in $order; do
    echo "  - $comp"
done

echo ""
echo "Test complete!"