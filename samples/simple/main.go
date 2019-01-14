package main

// telegraph-go/samples/simple

import (
	"log"

	telegraph "github.com/meinside/telegraph-go"
)

const (
	//verbose = false
	verbose = true

	html = `This page is for <b>testing</b>.
THIS IS THE FIRST LINE.
<p>
	<img src="http://i2.cdn.cnn.com/cnnnext/dam/assets/160927210830-tk-ah0927-exlarge-169.jpg" alt="Pepe the frog">
</p>
<font color="#FF0000">THIS IS THE RED LINE.</font> <i>(coloring not supported?)</i>`
)

func main() {
	telegraph.Verbose = verbose

	var savedAccessToken string

	// (1) XXX - create a new account,
	if client, err := telegraph.Create("telegraph-test", "Telegraph Test", ""); err == nil {
		log.Printf("> Created client: %#+v", client)

		savedAccessToken = client.AccessToken

		// GetAccountInfo
		if account, err := client.GetAccountInfo(nil); err == nil {
			log.Printf("> GetAccountInfo result: %#+v", account)
		} else {
			log.Printf("* GetAccountInfo error: %s", err)
		}

		// EditAccountInfo
		if account, err := client.EditAccountInfo("telegraph-test", "Telegraph API Test", ""); err == nil {
			log.Printf("> EditAccountInfo result: %#+v", account)
		} else {
			log.Printf("* EditAccountInfo error: %s", err)
		}

		// CreatePage
		if page, err := client.CreatePageWithHTML("Test page", "Telegraph Test", "", html, true); err == nil {
			log.Printf("> CreatePage result: %#+v", page)
			log.Printf("> Created page url: %s", page.URL)

			// GetPage
			if page, err := client.GetPage(page.Path, true); err == nil {
				log.Printf("> GetPage result: %#+v", page)
			} else {
				log.Printf("* GetPage error: %s", err)
			}

			content, _ := telegraph.NewNodesWithHTML(html)

			// EditPage
			if page, err := client.EditPage(page.Path, "Test page (edited)", content, "", "http://www.google.com", true); err == nil {
				log.Printf("> EditPage result: %#+v", page)
				log.Printf("> Edited page url: %s", page.URL)
			} else {
				log.Printf("* EditPage error: %s", err)
			}
		} else {
			log.Printf("* CreatePage error: %s", err)
		}

		// GetPageList
		if pages, err := client.GetPageList(0, 50); err == nil {
			log.Printf("> GetPageList result: %#+v", pages)

			for _, page := range pages.Pages {
				// GetViews
				if views, err := client.GetViews(page.Path, 2016, 0, 0, -1); err == nil {
					log.Printf("> GetViews result for %s: %#+v", page.Path, views)
				} else {
					log.Printf("* GetViews error: %s", err)
				}
			}
		} else {
			log.Printf("* GetPageList error: %s", err)
		}
	} else {
		log.Printf("* Create error: %s", err)
	}

	if savedAccessToken == "" {
		log.Printf("* Couldn't save access token, exiting...")
		return
	}

	// (2) XXX - or load an existing account with your access token
	if client, err := telegraph.Load(savedAccessToken); err == nil {
		log.Printf("> Loaded client: %#+v", client)

		// GetAccountInfo
		if account, err := client.GetAccountInfo(nil); err == nil {
			log.Printf("> GetAccountInfo result: %#+v", account)
		} else {
			log.Printf("* GetAccountInfo error: %s", err)
		}

		// EditAccountInfo
		if account, err := client.EditAccountInfo("telegraph-test", "Telegraph API Test", ""); err == nil {
			log.Printf("> EditAccountInfo result: %#+v", account)
		} else {
			log.Printf("* EditAccountInfo error: %s", err)
		}

		// CreatePage
		if page, err := client.CreatePageWithHTML("Test page", "Telegraph Test", "", html, true); err == nil {
			log.Printf("> CreatePage result: %#+v", page)
			log.Printf("> Created page url: %s", page.URL)

			// GetPage
			if page, err := client.GetPage(page.Path, true); err == nil {
				log.Printf("> GetPage result: %#+v", page)
			} else {
				log.Printf("* GetPage error: %s", err)
			}

			content, _ := telegraph.NewNodesWithHTML(html)

			// EditPage
			if page, err := client.EditPage(page.Path, "Test page (edited)", content, "", "http://www.google.com", true); err == nil {
				log.Printf("> EditPage result: %#+v", page)

				log.Printf("> Edited page url: %s", page.URL)
			} else {
				log.Printf("* EditPage error: %s", err)
			}
		} else {
			log.Printf("* CreatePage error: %s", err)
		}

		// GetPageList
		if pages, err := client.GetPageList(0, 50); err == nil {
			log.Printf("> GetPageList result: %#+v", pages)

			for _, page := range pages.Pages {
				// GetViews
				if views, err := client.GetViews(page.Path, 2016, 0, 0, -1); err == nil {
					log.Printf("> GetViews result for %s: %#+v", page.Path, views)
				} else {
					log.Printf("* GetViews error: %s", err)
				}
			}
		} else {
			log.Printf("* GetPageList error: %s", err)
		}

		// RevokeAccessToken
		if account, err := client.RevokeAccessToken(); err == nil {
			log.Printf("> RevokeAccessToken result: %#+v", account)
		} else {
			log.Printf("* RevokeAccessToken error: %s", err)
		}
	} else {
		log.Printf("* Load error: %s", err)
	}
}
