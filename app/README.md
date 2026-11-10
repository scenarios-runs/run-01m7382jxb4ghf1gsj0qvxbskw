# Sample app: Go / net/http

A tiny notes API with no dependencies outside the standard library.

- `GET /healthz` returns `{"status":"ok"}`
- `GET /notes`, `POST /notes` keep notes in memory
- the port comes from `PORT` (default 8000)

```bash
go vet ./...
go test ./...
go run .
```
