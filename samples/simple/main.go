package main

// telegraph-go/samples/simple
//
// Creates a throwaway Telegraph account, publishes a page and reads it back.

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"time"

	telegraph "github.com/meinside/telegraph-go"
)

const html = `This page is for <b>testing</b>.
THIS IS THE FIRST LINE.
<p>
	<img src="https://telegra.ph/images/logo.png" alt="Telegraph logo">
</p>
<font color="#FF0000">THIS IS THE RED LINE.</font> <i>(coloring not supported)</i>`

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// log every request at debug level (the access token is redacted)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// create a new account; to reuse one, keep client.AccessToken() and call telegraph.Load
	client, err := telegraph.Create(ctx, telegraph.CreateAccountParams{
		ShortName:  "telegraph-test",
		AuthorName: "Telegraph Test",
	}, telegraph.WithLogger(logger))
	if err != nil {
		log.Fatalf("Create: %v", err)
	}

	account, err := client.GetAccountInfo(ctx, telegraph.FieldShortName, telegraph.FieldAuthURL)
	if err != nil {
		log.Fatalf("GetAccountInfo: %v", err)
	}
	log.Printf("account %q, browser login (valid 5 minutes): %s", account.ShortName, account.AuthURL)

	page, err := client.CreatePageWithHTML(ctx, telegraph.CreatePageParams{Title: "Test page"}, html)
	if err != nil {
		var apiErr *telegraph.APIError
		if errors.As(err, &apiErr) {
			if wait, ok := apiErr.FloodWait(); ok {
				log.Fatalf("rate limited, retry after %s", wait)
			}
		}
		log.Fatalf("CreatePageWithHTML: %v", err)
	}
	log.Printf("created %s", page.URL)

	page, err = client.GetPage(ctx, page.Path, true)
	if err != nil {
		log.Fatalf("GetPage: %v", err)
	}
	rendered, err := telegraph.NodesToHTML(page.Content)
	if err != nil {
		log.Fatalf("NodesToHTML: %v", err)
	}
	log.Printf("content as HTML: %s", rendered)

	// append a paragraph to the fetched content and save it
	page.Content = append(page.Content, telegraph.NodeElement{Tag: "p", Children: telegraph.Nodes{"(edited)"}})
	if _, err := client.EditPage(ctx, telegraph.EditPageParams{
		Path:    page.Path,
		Title:   "Test page (edited)",
		Content: page.Content,
	}); err != nil {
		log.Fatalf("EditPage: %v", err)
	}

	views, err := client.GetViews(ctx, page.Path, telegraph.GetViewsParams{Year: time.Now().Year()})
	if err != nil {
		log.Fatalf("GetViews: %v", err)
	}
	log.Printf("views this year: %d", views.Views)
}
