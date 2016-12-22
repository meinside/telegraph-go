# telegraph-go

Go library for using [Telegraph API](http://telegra.ph/api).

## Install

```bash
$ go get -u github.com/meinside/telegraph-go/...
```

## Usage

### Sample code 1

```go
package main

import (
	"log"

	tg "github.com/meinside/telegraph-go"
)

func main() {
	//tg.IsVerbose = true // XXX - for verbose output

	// CreateAccount
	if account, err := tg.CreateAccount("telegraph-test", "Telegraph Test", ""); err == nil {
		log.Printf("> CreateAccount result: %#+v\n", account)

		accessToken := account.AccessToken

		// GetAccountInfo
		if account, err := tg.GetAccountInfo(accessToken, nil); err == nil {
			log.Printf("> GetAccountInfo result: %#+v\n", account)
		} else {
			log.Printf("* GetAccountInfo error: %s\n", err)
		}

		// EditAccountInfo
		if account, err := tg.EditAccountInfo(accessToken, "telegraph-test", "Telegraph API Test", ""); err == nil {
			log.Printf("> EditAccountInfo result: %#+v\n", account)
		} else {
			log.Printf("* EditAccountInfo error: %s\n", err)
		}

		/*
			THIS IS THE FIRST LINE.
			<p>
				<img src="..." alt="...">
				<b>THIS IS AN IMAGE.</b>
			</p>
		*/
		content := []tg.Node{
			tg.NewNodeWithString("THIS IS THE FIRST LINE."),
			tg.NewNodeWithElement(
				"p",
				nil,
				[]tg.Node{
					tg.NewNodeWithElement(
						"img",
						map[string]string{
							"src": "http://i2.cdn.cnn.com/cnnnext/dam/assets/160927210830-tk-ah0927-exlarge-169.jpg",
							"alt": "Pepe the frog",
						},
						nil,
					),
					tg.NewNodeWithElement(
						"b",
						nil,
						[]tg.Node{
							tg.NewNodeWithString("THIS IS AN IMAGE."),
						},
					),
				},
			),
		}

		// CreatePage
		if page, err := tg.CreatePage(accessToken, "Test page", "Telegraph Test", "", content, true); err == nil {
			log.Printf("> CreatePage result: %#+v\n", page)

			log.Printf("> Created page url: %s\n", page.Url)

			// GetPage
			if page, err := tg.GetPage(page.Path, true); err == nil {
				log.Printf("> GetPage result: %#+v\n", page)
			} else {
				log.Printf("* GetPage error: %s\n", err)
			}

			// EditPage
			if page, err := tg.EditPage(accessToken, page.Path, "Test page (edited)", content, "", "http://www.google.com", true); err == nil {
				log.Printf("> EditPage result: %#+v\n", page)

				log.Printf("> Edited page url: %s\n", page.Url)
			} else {
				log.Printf("* EditPage error: %s\n", err)
			}
		} else {
			log.Printf("* CreatePage error: %s\n", err)
		}

		// GetPageList
		if pages, err := tg.GetPageList(accessToken, 0, 50); err == nil {
			log.Printf("> GetPageList result: %#+v\n", pages)

			for _, page := range pages.Pages {
				// GetViews
				if views, err := tg.GetViews(page.Path, 2016, 0, 0, -1); err == nil {
					log.Printf("> GetViews result for %s: %#+v\n", page.Path, views)
				} else {
					log.Printf("* GetViews error: %s\n", err)
				}
			}
		} else {
			log.Printf("* GetPageList error: %s\n", err)
		}

		// RevokeAccessToken
		if account, err := tg.RevokeAccessToken(accessToken); err == nil {
			log.Printf("> RevokeAccessToken result: %#+v\n", account)
		} else {
			log.Printf("* RevokeAccessToken error: %s\n", err)
		}
	} else {
		log.Printf("* CreateAccount error: %s\n", err)
	}
}
```

### Sample code 2

A little more convenient way using a wrapper client:

```go
package main

import (
	"log"

	tg "github.com/meinside/telegraph-go"
	tgcl "github.com/meinside/telegraph-go/client"
)

func main() {
	// (1) XXX - create a new account,
	//if client, err := tgcl.Create("telegraph-test", "Telegraph Test", ""); err == nil {
	// (2) XXX - or load an existing account with your access token
	if client, err := tgcl.Load("abcdefghijklmnopqrstuvwxyz0123456789" /* access token */); err == nil {
		log.Printf("> Created/loaded client: %#+v\n", client)

		// GetAccountInfo
		if account, err := client.GetAccountInfo(nil); err == nil {
			log.Printf("> GetAccountInfo result: %#+v\n", account)
		} else {
			log.Printf("* GetAccountInfo error: %s\n", err)
		}

		// EditAccountInfo
		if account, err := client.EditAccountInfo("telegraph-test", "Telegraph API Test", ""); err == nil {
			log.Printf("> EditAccountInfo result: %#+v\n", account)
		} else {
			log.Printf("* EditAccountInfo error: %s\n", err)
		}

		// HTML for page
		html := `This page is for <b>testing</b>.
THIS IS THE FIRST LINE.
<p>
	<img src="http://i2.cdn.cnn.com/cnnnext/dam/assets/160927210830-tk-ah0927-exlarge-169.jpg" alt="Pepe the frog">
</p>
<font color="#FF0000">THIS IS THE RED LINE.</font> <i>(coloring not supported?)</i>`

		// CreatePage
		if page, err := client.CreatePageWithHtml("Test page", "Telegraph Test", "", html, true); err == nil {
			log.Printf("> CreatePage result: %#+v\n", page)

			log.Printf("> Created page url: %s\n", page.Url)

			// GetPage
			if page, err := client.GetPage(page.Path, true); err == nil {
				log.Printf("> GetPage result: %#+v\n", page)
			} else {
				log.Printf("* GetPage error: %s\n", err)
			}

			content, _ := tg.NewNodesWithHtml(html)
			/*
				// => will be converted to:

				[]telegraph.Node{
					"This page is for ",
					telegraph.NodeElement{
						Tag:"b",
						Attrs:map[string]string{},
						Children:[]telegraph.Node{
							"testing",
						},
					},
					".\nTHIS IS THE FIRST LINE.\n",
					telegraph.NodeElement{
						Tag:"p",
						Attrs:map[string]string{},
						Children:[]telegraph.Node{
							"\n\t",
							telegraph.NodeElement{
								Tag:"img",
								Attrs:map[string]string{
									"src": "http://i2.cdn.cnn.com/cnnnext/dam/assets/160927210830-tk-ah0927-exlarge-169.jpg",
									"alt": "Pepe the frog",
								},
								Children:[]telegraph.Node{},
							},
							"\n",
						},
					},
					"\n",
					telegraph.NodeElement{
						Tag:"font",
						Attrs:map[string]string{
							"color":"#FF0000",
						},
						Children:[]telegraph.Node{
							"THIS IS THE RED LINE.",
						},
					},
					" ",
					telegraph.NodeElement{
						Tag:"i",
						Attrs:map[string]string{},
						Children:[]telegraph.Node{
							"(coloring not supported?)",
						},
					},
				}
			*/

			// EditPage
			if page, err := client.EditPage(page.Path, "Test page (edited)", content, "", "http://www.google.com", true); err == nil {
				log.Printf("> EditPage result: %#+v\n", page)

				log.Printf("> Edited page url: %s\n", page.Url)
			} else {
				log.Printf("* EditPage error: %s\n", err)
			}
		} else {
			log.Printf("* CreatePage error: %s\n", err)
		}

		// GetPageList
		if pages, err := client.GetPageList(0, 50); err == nil {
			log.Printf("> GetPageList result: %#+v\n", pages)

			for _, page := range pages.Pages {
				// GetViews
				if views, err := client.GetViews(page.Path, 2016, 0, 0, -1); err == nil {
					log.Printf("> GetViews result for %s: %#+v\n", page.Path, views)
				} else {
					log.Printf("* GetViews error: %s\n", err)
				}
			}
		} else {
			log.Printf("* GetPageList error: %s\n", err)
		}

		// RevokeAccessToken
		/*
			if account, err := client.RevokeAccessToken(); err == nil {
				log.Printf("> RevokeAccessToken result: %#+v\n", account)
			} else {
				log.Printf("* RevokeAccessToken error: %s\n", err)
			}
		*/
	} else {
		log.Printf("* Create/load error: %s\n", err)
	}
}
```

## Todo

- [X] Add a helper function for converting HTML strings into []telegraph.Node
- [ ] Add tests
- [ ] Add benchmarks

## License

MIT

