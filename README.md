![Visualization of the codebase](./diagram.svg)

## TODO:

- [ ] implement chat functionality using websockets

- [ ] basic rendering tests for ui

- [ ] use lib for mocking database 
https://pkg.go.dev/github.com/DATA-DOG/go-sqlmock 
https://tanutaran.medium.com/golang-unit-testing-with-gorm-and-sqlmock-postgresql-simplest-setup-67ccc7c056ef

- [ ] deploy to aws
    - [ ] using terraform
    - [ ] using cloudformation?

- [ ] use fiber auth middleware https://docs.gofiber.io/api/middleware/basicauth
- [ ] add auth https://medium.com/@abhinavv.singh/a-comprehensive-guide-to-authentication-and-authorization-in-go-golang-6f783b4cea18
- [ ] use JWT tokens
- [ ] in general writing a chat app
- [ ] write a signup in tdd fashion
- [ ] use tdd, coverage to >90%
https://medium.com/@engmiladkh1372/test-suite-for-unit-testing-gorm-in-go-47b6ea8d4ab0
https://pkg.go.dev/github.com/stretchr/testify/suite



- [ ] add photo storage: 2 options: postgresql, s3-like storage
implement 2 options with interchangibility option
maybe use cache for storing lately accessed images

- [ ] add cache (e.g. redis)
- [ ] add pubsub/message queue of some kind
- [ ] auth with permissions roles (user, admin, chat admin)
- [ ] setup swagger https://docs.gofiber.io/contrib/swagger_v1.x.x/swagger/
- [ ] try https://github.com/go-gorm/gen
- [ ] use protobufs
- [ ] setup monkey testing
- [ ] try BDD
- [ ] setup docker-compose
- [ ] setup ci/cd
- [ ] use all features of gorm
- [ ] use all features of fiber
- [ ] setup automatic backup for database and images
- [ ] use db hooks for something
- [ ] add chatgpt integration as bot
- [ ] learn to use os, os/exec, io, bytes libs
- [ ] learn to use tags
- [ ] use goroutines somewhere
- [ ] play around and debug internaly of fiber, gorm, docker etc
- [ ] use github copilot free trial
- [ ] use factories for tests. is it useful in golang? https://github.com/bluele/factory-go
- [ ] setting up db for tests - setup, teardown - in other projects
- [ ] setup database migrations
- [ ] try https://github.com/sqlc-dev/sqlc and maybe benchmark
- [ ] write raw SQL query
- [ ] review templates e.g https://github.com/create-go-app/fiber-go-template/tree/master and others
- [ ] try supabase
- [ ] add payments
- [ ] implement rate limiting and maybe some more features from distributed applications :)
- [ ] try planetscale database
- [ ] structure logging https://pkg.go.dev/golang.org/x/exp/slog
- [ ] setup sentry or some other monitoring
- [ ] try rpc, grpc, webrtc
- [ ] setup linter for function length and code complexity
- [ ] setup load testing https://github.com/tsenart/vegeta
- [ ] add api versioning
- [ ] try out many of fiber middlewares
- [ ] setup vscode's dev containers
- [ ] improve knowledge of Makefile, bash scripting 
- [x] setup linter for function length and code complexity
- [ ] try using https://github.com/KillianLucas/open-interpreter/
- [ ] can be interesting to add messaging library as alternative for manually implemented one https://github.com/centrifugal/centrifugo?tab=readme-ov-file
- [ ] setup tls
- [ ] find mockup to html tool e.g. https://github.com/SawyerHood/draw-a-ui