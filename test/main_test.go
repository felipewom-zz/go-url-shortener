package test

import (
	"github.com/felipewom/go-url-shortener/internal/server"
	"github.com/kataras/iris/v12/httptest"
	"testing"
)

func TestURLShortener(t *testing.T) {

	app, db := server.Startup()

	e := httptest.New(t, app)
	originalURL := "https://google.com"

	// save
	e.POST("/shorten").
		WithFormField("url", originalURL).
		Expect().
		Status(httptest.StatusOK).
		Body().
		Contains("<pre><a target='_new' href=")

	keys := db.GetByValue(originalURL)
	if got := len(keys); got != 1 {
		t.Fatalf("expected to have 1 key but saved %d short urls", got)
	}

	// get
	e.GET("/u/" + keys[0]).Expect().
		Status(httptest.StatusTemporaryRedirect).Header("Location").Equal(originalURL)

	// save the same again, it should add a new key
	e.POST("/shorten").
		WithFormField("url", originalURL).Expect().
		Status(httptest.StatusOK).Body().Contains("<pre><a target='_new' href=")

	keys2 := db.GetByValue(originalURL)
	if got := len(keys2); got != 1 {
		t.Fatalf("expected to have 1 keys even if we save the same original url but saved %d short urls", got)
	} // the key is the same, so only the first one matters.

	if keys[0] != keys2[0] {
		t.Fatalf("expected keys to be equal if the original url is the same, but got %s = %s ", keys[0], keys2[0])
	}

	// clear db
	e.POST("/clear_cache").Expect().Status(httptest.StatusOK)
	if got := db.Len(); got != 0 {
		t.Fatalf("expected database to have 0 registered objects after /clear_cache but has %d", got)
	}

	// give it some time to release the db connection
	defer db.Close()
}
