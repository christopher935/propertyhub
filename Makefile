.PHONY: build validate routes check pre-commit setup-hooks

build:
	@echo "🔨 Building PropertyHub server..."
	@go build -o propertyhub-server ./cmd/server
	@echo "✅ Build complete"

validate: build
	@echo "🔍 Validating routes..."
	@./propertyhub-server --dry-run
	@echo "✅ Validation complete"

routes: build
	@echo "📋 Registered routes:"
	@./propertyhub-server --list-routes 2>/dev/null || echo "Note: --list-routes not implemented yet"

check: validate
	@echo "✅ All checks passed"

setup-hooks:
	@chmod +x scripts/setup-hooks.sh
	@./scripts/setup-hooks.sh

pre-commit: validate
