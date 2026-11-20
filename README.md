# telegraph-go

Go library for using [Telegraph API](https://telegra.ph/api).

## Install

```bash
$ go get -u github.com/meinside/telegraph-go
```

## Usage

```go
ctx := context.Background()

// new account (keep client.AccessToken() to reuse it later with telegraph.Load)
client, err := telegraph.Create(ctx, telegraph.CreateAccountParams{ShortName: "my-blog"})

// publish HTML; unsupported tags and attributes are filtered out
page, err := client.CreatePageWithHTML(ctx, telegraph.CreatePageParams{Title: "Hello"}, "<p>Hello, <b>world</b></p>")

// fetched content decodes to string / telegraph.NodeElement and can be sent back as is
page, err = client.GetPage(ctx, page.Path, true)
html, err := telegraph.NodesToHTML(page.Content)
```

- Options: `WithHTTPClient`, `WithBaseURL`, `WithLogger` (see `telegraph.go`).
- Errors: `*APIError` (with `FloodWait()`), `*HTTPError`, and `ErrInvalidParameter` for
  parameters rejected before sending (see `errors.go`).

More in [./samples/](https://github.com/meinside/telegraph-go/tree/master/samples).

## Test

```bash
$ go test -short ./...   # offline tests only
$ go test ./...          # also TestLiveAPI: creates an account and pages on telegra.ph
```

## Todo

- [X] Add a helper function for converting HTML strings into []telegraph.Node
- [X] Support go module
- [X] Add tests
- [ ] Add benchmarks
