build:
	gox -output="bin/{{.Dir}}_{{.OS}}_{{.Arch}}"

server:
	go run main.go --debug true

test:
	go test -v -cover ./...

docker:
	docker build -t fuellabs/proxy .

docker-push:
	docker push fuellabs/proxy

.PHONY: build server test docker docker-push
