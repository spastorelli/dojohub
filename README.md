### Requirements

* [Go](http://golang.org/doc/install)
* [Heroku Toolbelt](https://toolbelt.heroku.com/) installed.

### Build and run locally

1. Install `govendor`:
  ```sh
  go get -u github.com/kardianos/govendor
  ```
2. Pull dependencies into `vendor/` directory:
  ```sh
  $GOPATH/bin/govendor sync
  ```
3. Run the app:
  * using heroku CLI:
  ```sh
  heroku local dev
  ```
  * or, alternatively:
  ```sh
  go run dojohub.go -host 127.0.0.1
  ```
  **Note:** See `-help` for more usage options.

## Deploying to Heroku

```sh
heroku create
git push heroku master
heroku open
```

or

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)
