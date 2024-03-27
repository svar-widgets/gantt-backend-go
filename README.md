Backend for SVAR Data Gantt
===========================

### How to start

- create config.yml with DB access config

```yaml
db:
  path: db.sqlite
  resetonstart: true
server:
  url: "http://localhost:3000"
  port: ":3000"
  cors:
    - "*"
```

- start the backend

```shell script
go build
./gantt-backend-go
```
