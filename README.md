# Go - SnippetBox

## Usage:

### 1. Clone Repository
```
git clone https://github.com/aesuhaendi/go-snippetbox.git
```

### 2. Generate Self-Signed TLS Certificate
```
cd %THIS_PROJECT%\tls
go run %GOROOT%\src\crypto\tls\generate_cert.go --rsa-bits=2048 --host=localhost
```

### 3. Create DB, Table, and insert some data
```
--- Please check sql dir ---
```

### 4. Running HTTPS Server
```
go run ./cmd/web
```
If you are using air development tool, just run air in project dir
```
air
```
NOTE: Please check 'cmd/web/main.go' for command-line arguments or just run code below:
```
go run ./cmd/web --help
```

### 5. Try to create user accounts and snippets