#!/bin/bash

# Template Refactoring Helper Script
# Extracts content section from old template format for conversion to base template

if [ $# -ne 1 ]; then
    echo "Usage: $0 <template-file.html>"
    exit 1
fi

FILE="$1"
BASENAME=$(basename "$FILE")

echo "Analyzing: $BASENAME"
echo ""

# Find line numbers for key sections
SIDEBAR_START=$(grep -n "admin-sidebar" "$FILE" | head -1 | cut -d: -f1)
ADMIN_MAIN_START=$(grep -n "admin-main" "$FILE" | head -1 | cut -d: -f1)
ADMIN_CONTENT_START=$(grep -n "admin-content" "$FILE" | head -1 | cut -d: -f1)
FIRST_SCRIPT=$(grep -n "<script" "$FILE" | head -1 | cut -d: -f1)

echo "Structure:"
echo "  - Sidebar starts: line $SIDEBAR_START"
echo "  - Main starts: line $ADMIN_MAIN_START"
echo "  - Content starts: line $ADMIN_CONTENT_START"
echo "  - Scripts start: line $FIRST_SCRIPT"
echo ""

TOTAL_LINES=$(wc -l < "$FILE")
BOILERPLATE=$((ADMIN_CONTENT_START - 1))
CONTENT_LINES=$((FIRST_SCRIPT - ADMIN_CONTENT_START - 2))
SCRIPT_LINES=$((TOTAL_LINES - FIRST_SCRIPT))

echo "Breakdown:"
echo "  - Boilerplate (sidebar/header): $BOILERPLATE lines (will be removed)"
echo "  - Content section: $CONTENT_LINES lines (will be kept)"
echo "  - Scripts section: $SCRIPT_LINES lines (will be kept)"
echo "  - Total: $TOTAL_LINES lines"
echo ""
echo "After refactoring: ~$((CONTENT_LINES + SCRIPT_LINES + 20)) lines ($(echo "scale=1; 100 * ($BOILERPLATE) / $TOTAL_LINES" | bc)% reduction)"
