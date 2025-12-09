#!/bin/bash

HOOKS_DIR=".githooks"
GIT_HOOKS_DIR=".git/hooks"

echo "🔧 Setting up git hooks..."

git config core.hooksPath $HOOKS_DIR

chmod +x $HOOKS_DIR/*

echo "✅ Git hooks configured"
echo "   Pre-commit hook will now validate routes before each commit"
