package server

import (
	"github.com/felipewom/go-url-shortener/internal/factory"
	"github.com/felipewom/go-url-shortener/internal/store"
	"github.com/felipewom/go-url-shortener/utils"
	"github.com/kataras/iris/v12"
	"html/template"
	"log"
	"os"
	"path"
)

func Startup() (*iris.Application, store.DB) {
	// Pass that db to our server, in order to be able to test the whole server with a different database later on.
	app, db := newApp()
	// release the "db" connection when server goes off.
	iris.RegisterOnInterrupt(db.Close)
	port := utils.GetEnvPort()
	err := app.Run(iris.Addr(":" + port))
	if err != nil {
		log.Fatalf("Error during startup %+v", err)
	}
	return app, db
}

func newApp() (*iris.Application, store.DB) {
	app := iris.Default() // or server := iris.New()
	// create our factory, which is the manager for the object creation.
	// between our web server and the db.
	// assign a variable to the DB so we can use its features later.

	db := store.NewDB()
	fct := factory.NewFactory(factory.DefaultGenerator, db)
	dirname, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	webFolder, err := os.Open(path.Join(dirname, "/web/resources/public"))
	if err != nil {
		panic(err)
	}

	// serve the "./web/resources/public" directory's "*.html" files with the HTML std view engine.
	tmpl := iris.HTML(webFolder.Name(), ".html").Reload(true)
	// register any template func(s) here.
	//
	// Look ./web/resources/public/index.html#L16
	tmpl.AddFunc("IsPositive", func(n int) bool {
		if n > 0 {
			return true
		}
		return false
	})

	app.RegisterView(tmpl)

	// Serve static files (css)
	app.HandleDir("/static", webFolder.Name())

	indexHandler := func(ctx iris.Context) {
		ctx.ViewData("URL_COUNT", db.Len())
		ctx.ViewData("URL_LIST", db.GetAll())
		ctx.View("index.html")
	}
	app.Get("/", indexHandler)

	// find and execute a short url by its key
	// used on http://localhost:8080/u/dsaoj41u321dsa
	execShortURL := func(ctx iris.Context, key string) {
		if key == "" {
			ctx.StatusCode(iris.StatusBadRequest)
			return
		}

		value := db.Get(key)
		if value == "" {
			ctx.StatusCode(iris.StatusNotFound)
			ctx.Writef("Short URL for key: '%s' not found", key)
			return
		}

		ctx.Redirect(value, iris.StatusTemporaryRedirect)
	}
	app.Get("/u/{shortkey}", func(ctx iris.Context) {
		execShortURL(ctx, ctx.Params().Get("shortkey"))
	})

	app.Post("/shorten", func(ctx iris.Context) {
		formValue := ctx.FormValue("url")
		if formValue == "" {
			ctx.ViewData("FORM_RESULT", "You need to a enter a URL")
			ctx.StatusCode(iris.StatusLengthRequired)
			return
		}
		key, err := fct.Gen(formValue)
		if err != nil {
			ctx.ViewData("FORM_RESULT", "Invalid URL")
			ctx.StatusCode(iris.StatusBadRequest)
		} else {
			if err = db.Set(key, formValue); err != nil {
				ctx.ViewData("FORM_RESULT", "Internal error while saving the URL")
				app.Logger().Infof("while saving URL: " + err.Error())
				ctx.StatusCode(iris.StatusInternalServerError)
			} else {
				ctx.StatusCode(iris.StatusOK)
				shortenURL := "/u/" + key
				ctx.ViewData("FORM_RESULT",
					template.HTML("<pre><a target='_new' href='"+shortenURL+"'>"+key+" </a></pre>"))
			}
		}

		indexHandler(ctx) // no redirect, we need the FORM_RESULT.
	})

	app.Post("/clear_cache", func(ctx iris.Context) {
		err := db.Clear()
		if err != nil {
			ctx.JSON(err)
			return
		}
		ctx.Redirect("/")
	})

	return app, *db
}
