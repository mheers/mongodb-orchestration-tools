# Run locally
```
export MONGODB_INITDBS='[{"name": "demo", "username": "demouser", "password": "demopassword"},{"name": "demo2", "username": "demouser2", "password": "demopassword2"}]'
go run cmd/k8s-mongodb-initiator/main.go
```

# Test locally
```
make test-full-prepare
GO_TEST_PATH=./controller/user/ make test-full
GO_TEST_PATH=./controller/replset/ make test-full
make test-full-clean
```