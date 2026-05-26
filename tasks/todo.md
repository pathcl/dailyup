# dailyup — Task Tracker

## Done

- [x] Initialize Go module and directory structure
- [x] config package — TDD (config_test.go → config.go)
- [x] azdevops/client — TDD (client_test.go → client.go)
- [x] azdevops/workitems — TDD (workitems_test.go → workitems.go)
- [x] azdevops/pullrequests — TDD (pullrequests_test.go → pullrequests.go)
- [x] azdevops/commits — TDD (commits_test.go → commits.go)
- [x] report/markdown — TDD (markdown_test.go → markdown.go)
- [x] cmd/root.go + main.go — CLI wiring
- [x] go build ./... green
- [x] go test ./... green (16 tests)

## Usage

1. Install: `go install github.com/pathcl/dailyup@latest` or `go build -o dailyup .`
2. Create `~/.config/dailyup/config.toml`:
   ```toml
   organization = "myorg"
   project      = "myproject"
   tags         = ["sprint-23"]
   weeks        = 2
   email        = "me@example.com"   # optional, for commit filtering
   ```
3. Authenticate: `az login`
4. Run: `dailyup summary`
