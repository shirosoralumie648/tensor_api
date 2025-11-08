# tensor_api

 A Go-based AI API management service with a modern web app.

 - Full Chinese documentation and screenshots: [README_zh-CN.md](./README_zh-CN.md)

 ## Quick start (development)

 Go
 - Go 1.24+
 - Build: `go build ./...`
 - Test: `go test ./...`

 Web (app)
 - Node.js 20+ and pnpm
 - In `app/`: `pnpm install && pnpm build`

 ## CI/CD

 - GitHub Actions validates Go build/tests and app build on pushes and pull requests.
 - Dependabot maintains Go modules, npm packages, and GitHub Actions.

 ## Contributing

 See CONTRIBUTING.md. Code of Conduct: admin@shirosora.cn

 ## License

 Apache-2.0