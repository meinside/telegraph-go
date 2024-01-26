package telegraph

import (
	"os"
	"testing"
)

const (
	htmlForTest = `This page is for <b>testing</b>.
THIS IS THE FIRST LINE.
<p>
	<img src="http://i2.cdn.cnn.com/cnnnext/dam/assets/160927210830-tk-ah0927-exlarge-169.jpg" alt="Pepe the frog">
</p>
<font color="#FF0000">THIS IS THE RED LINE.</font> <i>(coloring not supported?)</i>`
)

func TestMethods(t *testing.T) {
	Verbose = os.Getenv("VERBOSE") == "true"

	var savedAccessToken string

	// (1) XXX - create a new account,
	if client, err := Create("telegraph-test", "Telegraph Test", ""); err == nil {
		savedAccessToken = client.AccessToken

		// GetAccountInfo
		if _, err := client.GetAccountInfo(nil); err != nil {
			t.Errorf("failed to get account info: %s", err)
		}

		// EditAccountInfo
		if _, err := client.EditAccountInfo("telegraph-test", "Telegraph API Test", ""); err != nil {
			t.Errorf("failed to edit account info: %s", err)
		}

		// CreatePage
		if page, err := client.CreatePageWithHTML("Test page", "Telegraph Test", "", htmlForTest, true); err == nil {
			// GetPage
			if _, err := client.GetPage(page.Path, true); err != nil {
				t.Errorf("failed to get created page: %s", err)
			}

			content, _ := NewNodesWithHTML(htmlForTest)

			// EditPage
			if _, err := client.EditPage(page.Path, "Test page (edited)", content, "", "http://www.google.com", true); err != nil {
				t.Errorf("failed to edit created page: %s", err)
			}
		} else {
			t.Errorf("failed to create page: %s", err)
		}

		// GetPageList
		if pages, err := client.GetPageList(0, 50); err == nil {
			for _, page := range pages.Pages {
				// GetViews
				if _, err := client.GetViews(page.Path, 2016, 0, 0, -1); err != nil {
					t.Errorf("failed to get views: %s", err)
				}
			}
		} else {
			t.Errorf("failed to get page list: %s", err)
		}
	} else {
		t.Errorf("failed to create a new client: %s", err)
	}

	if savedAccessToken == "" {
		t.Fatal("failed to save access token")
	}

	// (2) XXX - or load an existing account with your access token
	if client, err := Load(savedAccessToken); err == nil {
		// GetAccountInfo
		if _, err := client.GetAccountInfo(nil); err != nil {
			t.Errorf("failed to get account info with existing access token: %s", err)
		}

		// EditAccountInfo
		if _, err := client.EditAccountInfo("telegraph-test", "Telegraph API Test", ""); err != nil {
			t.Errorf("failed to edit account info with existing access token: %s", err)
		}

		// CreatePage
		if page, err := client.CreatePageWithHTML("Test page", "Telegraph Test", "", htmlForTest, true); err == nil {
			// GetPage
			if _, err := client.GetPage(page.Path, true); err != nil {
				t.Errorf("failed to get created page with existing access token: %s", err)
			}

			content, _ := NewNodesWithHTML(htmlForTest)

			// EditPage
			if _, err := client.EditPage(page.Path, "Test page (edited)", content, "", "http://www.google.com", true); err != nil {
				t.Errorf("failed to edit created page with existing access token: %s", err)
			}
		} else {
			t.Errorf("failed to create page with existing access token: %s", err)
		}

		// GetPageList
		if pages, err := client.GetPageList(0, 50); err == nil {
			for _, page := range pages.Pages {
				// GetViews
				if _, err := client.GetViews(page.Path, 2016, 0, 0, -1); err != nil {
					t.Errorf("failed to get views with existing access token: %s", err)
				}
			}
		} else {
			t.Errorf("failed to get page list with existing access token: %s", err)
		}

		// RevokeAccessToken
		if _, err := client.RevokeAccessToken(); err != nil {
			t.Errorf("failed to revoke access token: %s", err)
		}
	} else {
		t.Errorf("failed to load client with existing access token: %s", err)
	}
}
