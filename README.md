# Portfolio API
1. Jalankan docker postgresql

```
docker run --name postgresql -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password -e POSTGRES_DB=portfolio -d -p 5432:5432 postgres:16
```

2. Tambahkan dependency yang dibutuhkan

```
https://github.com/gin-gonic/gin
```

```
go get github.com/jackc/pgx/v5/stdlib
```

```
https://github.com/ivanauliaa/response-formatter
```

```
https://github.com/google/uuid
```

```
https://pkg.go.dev/golang.org/x/crypto
```

3. Export Variable yang dibutuhkan

```
postgres://user:password@localhost:5432/portfolio?sslmode=disable
```