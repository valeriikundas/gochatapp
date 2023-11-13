run:
	go run .

test:
	go test

cover:
	go test -coverprofile=coverage.out

report:
	go tool cover -html=coverage.out

redis:
	docker run --name redis -p 6379:6379 -v ./data/redis:/data -d redis

redis-ui:
	docker run --name redis-ui -v ./data/redis:/db -p 8001:8001 -d redislabs/redisinsight:latest

redis-cli:
	docker run --name redis-cli --rm -it goodsmileduck/redis-cli redis-cli -h host.docker.internal -p 6379
