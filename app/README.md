### Build and run the DojoHub app

1. Install `govendor`:
```
go get -u github.com/kardianos/govendor
```
2. Pull dependencies into `vendor/` directory:
```
$GOPATH/bin/govendor sync
```
3. Build and run DojoHub app:
```
go run dojohub.go
```
**Note:** See `-help` for more usage options.
