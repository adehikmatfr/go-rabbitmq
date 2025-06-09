test:
	@go test -v -count=1 -coverprofile=coverage.out -covermode=atomic ./pkg/...

coverage-html:
	@go tool cover -html=coverage.out -o coverage.html