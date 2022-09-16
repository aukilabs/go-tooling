test: go-normalize
	go test ./...

go-normalize:
	@go fmt ./...
	@go vet ./...

release: 
ifdef VERSION
	@echo "\033[94m\n• Releasing ${VERSION}\033[00m"
	@git tag ${VERSION}
	@git push origin ${VERSION}

else
	@echo "\033[94m\n• Releasing version\033[00m"
	@echo "\033[91mVERSION is not defided\033[00m"
	@echo "~> make VERSION=\033[90mv0.0.x\033[00m release"
endif