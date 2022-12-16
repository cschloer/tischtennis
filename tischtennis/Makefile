.PHONY: build clean deploy

build:
	cd code/getHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/indexBin index.go && cd ../..
	cd code/getHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/adminBin admin.go && cd ../..
	cd code/getHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/personBin person.go && cd ../..
	cd code/postHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/adminPostBin admin.go && cd ../..

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose
