#!/bin/bash

echo "🔍 Analyzing route files for potential conflicts..."

ROUTE_FILES="cmd/server/routes_api.go cmd/server/routes_admin.go cmd/server/routes_consumer.go"
ERRORS=0

for file in $ROUTE_FILES; do
    if [ -f "$file" ]; then
        echo "📄 Checking $file..."
        
        grep -nE '\.(GET|POST|PUT|DELETE|PATCH)\s*\("[^"]+' "$file" | while read line; do
            echo "   $line"
        done
    fi
done

echo ""
echo "🔎 Checking for duplicate path strings..."

ALL_PATHS=$(grep -hroE '"\/(api\/)?v1\/[^"]+"|"\/[^"]+' $ROUTE_FILES 2>/dev/null | sort)
DUPLICATES=$(echo "$ALL_PATHS" | uniq -d)

if [ -n "$DUPLICATES" ]; then
    echo "⚠️  WARNING: Potential duplicate paths found:"
    echo "$DUPLICATES"
    echo ""
    echo "   Review these paths - they may be registered multiple times"
    ERRORS=1
else
    echo "✅ No obvious duplicate paths detected"
fi

echo ""
echo "🔎 Checking for v1 path registration conflicts..."

V1_DIRECT=$(grep -hE 'api\.(GET|POST|PUT|DELETE|PATCH)\s*\("/v1/' $ROUTE_FILES 2>/dev/null | wc -l)
V1_GROUP=$(grep -hE 'v1\s*:=\s*api\.Group' $ROUTE_FILES 2>/dev/null | wc -l)

if [ "$V1_DIRECT" -gt 0 ] && [ "$V1_GROUP" -gt 0 ]; then
    echo "⚠️  WARNING: Found both direct /v1/ routes AND v1 group!"
    echo "   Direct /v1/ registrations: $V1_DIRECT"
    echo "   v1 Group definitions: $V1_GROUP"
    echo "   This is likely to cause duplicate route conflicts!"
    ERRORS=1
else
    echo "✅ No v1 registration conflicts detected"
fi

echo ""
if [ $ERRORS -gt 0 ]; then
    echo "❌ Route linting found potential issues"
    exit 1
else
    echo "✅ Route linting passed"
    exit 0
fi
