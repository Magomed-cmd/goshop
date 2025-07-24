run:
	go run cmd/goshop/main.go
test:
	gotestsum --format testdox ./... -v -- -count=1

clean-mocks:
	@echo "🧹 Cleaning all mocks directories..."
	find . -type d -name "mocks" -exec rm -rf {} + 2>/dev/null || true
	@echo "✅ All mocks cleaned!"

regenerate-mocks: clean-mocks
	@echo "🔄 Regenerating all mocks..."
	mockery --all
	@echo "✅ All mocks regenerated!"

fresh-mocks: clean-mocks regenerate-mocks test