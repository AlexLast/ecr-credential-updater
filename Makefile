darwin:
	GO111MODULE=on GOOS=darwin GOARCH=amd64 go build -a -o build/updater github.com/alexlast/ecr-credential-updater/cmd/updater
test:
	go test ./internal/... ./cmd/... -coverprofile cover.out -timeout 20m