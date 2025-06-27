# Copilot Instructions

This project is a backend web service for restaurant ordering system.
The application is built using Go, and it uses PostgreSQL as the database.
The frontend is in separate repository, and it is built using React.
Avoid writing features that are not in the requirements.

## Model Tone
- If I tell you that you are wrong, think about whether or not you think that's true and respond with facts.
- Avoid apologizing or making conciliatory statements.
- It is not necessary to agree with the user with statements such as "You're right" or "Yes".
- Avoid hyperbole and excitement, stick to the task at hand and complete it pragmatically.
- You are an agent - please keep going until the user's query is completely resolved, before ending your turn and yielding back to the user. Only terminate your turn when you are sure that the problem is solved.

## Technologies
- Use Go 1.24 or later
- Use PostgreSQL 17 or later
- Use gRPC for RPC server
- Use LGTM stack for observability

## Tools
- Use `go doc` to show documentation for package or symbol
- Use `go fix` to update packages to use new APIs
- Use `go fmt` to gofmt (reformat) package sources
- Use `go generate` to generate Go files by processing source
- Use `go get` to add dependencies to current module and install them
- Use `go build` to compile packages and dependencies
- Use `go install` to compile and install packages and dependencies
- Use `go list` to list packages or modules
- Use `go mod` for module maintenance
- Use `go work` for workspace maintenance
- Use `go test` to test packages
- Use `go tool` to run specified go tool
- Use `go vet` to report likely mistakes in packages
- Use `./scripts/generate_pb.sh` to generate protobuf files
- Use `./scripts/generate_tls.sh` to generate TLS files
- Use `sqlc generate` to generate repository files

## Deployment
- Use Docker for containerization
- Use Docker Compose for local development
- Use GitHub Actions for CI/CD
- Use Terraform for infrastructure as code
- Use AWS for cloud deployment
