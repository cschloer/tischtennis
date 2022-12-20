.PHONY: build clean deploy

build:
	cd code/getHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/indexBin index.go && cd ../..
	cd code/getHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/adminBin admin.go && cd ../..
	cd code/getHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/personBin person.go && cd ../..
	cd code/postHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/adminPostBin admin.go && cd ../..
	cd code/postHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/personPostBin person.go && cd ../..
	cd code/postHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/gamePostBin game.go && cd ../..
	cd code/deleteHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/personDeleteBin person.go && cd ../..
	cd code/deleteHandlers && env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/gameDeleteBin game.go && cd ../..

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose
