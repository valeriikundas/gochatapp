setup:
	docker-compose up -d redis postgres

dev: setup
	air

run:
	go run .

test:
	go test .

cov-test:
	go test -coverprofile=tmp/coverage.out

cov-report: cov-test
	go tool cover -html=tmp/coverage.out -o tmp/coverage.html

cov-open: cov-test cov-report
	open tmp/coverage.html

redis:
	docker run --name redis -p 6379:6379 -v ./data/redis:/data -d redis

redis-ui:
	docker run --name redis-ui -v ./data/redis:/db -p 8001:8001 -d redislabs/redisinsight:latest

redis-cli:
	docker run --name redis-cli --rm -it goodsmileduck/redis-cli redis-cli -h host.docker.internal -p 6379

test-docker:
	# FIXME: check that database is ready
	docker-compose up -d redis postgres
	docker-compose exec -d postgres dropdb --user postgres --if-exists chatapp_test
	docker-compose exec -d postgres createdb --user postgres chatapp_test
	make test
	docker-compose down redis postgres

deploy:
	fly deploy
