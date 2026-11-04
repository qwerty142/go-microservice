build-all:
	mkdir -p build

	mkdir -p build/cart
	GOOS=linux GOARCH=amd64 go build -o ./build/cart ./cart/...

	mkdir -p build/loms
	GOOS=linux GOARCH=amd64 go build -o ./build/cart ./cart/...


run-all: build-all
	docker-compose up --force-recreate --build -d


lint:
	golangci-lint --config=./.golangci.yml run ./cart/...


lint-fix:
	golangci-lint --config=./.golangci.yml run --fix ./cart/...


unit-test:
	go test -v -count=1 -timeout=30s -race -short ./cart/...
	go test -v -count=1 -timeout=30s -race -short ./loms/...

e2e-test:
	go test -tags e2e -v -count=1 -timeout=30s ./cart/e2e/...

pprof-profile:
	go tool pprof loms-profile.pb.gz http://localhost:8081/debug/pprof/profile
	go tool pprof cart-profile.pb.gz http://localhost:8082/debug/pprof/profile
