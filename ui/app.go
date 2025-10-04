package ui

import (
	"context"
	"log"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/mastty/auth"
	"github.com/AbeEstrada/mastty/config"
	"github.com/mattn/go-mastodon"
)

type App struct {
	vx        *vaxis.Vaxis
	views     map[string]View
	view      View
	header    *Header
	footer    *Footer
	quitModal *QuitModal
	showQuit  bool
	running   bool
	loading   bool
	config    *config.Config
	client    *mastodon.Client
}

func CreateApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		auth.SetupAuth()
		log.Fatalf("Authentication required; re-run after setup")
	}

	vx, err := vaxis.New(vaxis.Options{})
	if err != nil {
		return nil, err
	}

	views := make(map[string]View)
	views["home"] = CreateHomeView()

	app := &App{
		vx:        vx,
		views:     views,
		view:      views["home"],
		header:    CreateHeader(),
		footer:    CreateFooter(vx),
		quitModal: CreateQuitModal(),
		showQuit:  false,
		running:   true,
		loading:   false,
		config:    cfg,
	}

	for _, view := range views {
		view.SetApp(app)
	}

	return app, nil
}

func (app *App) SetView(name string) {
	if view, ok := app.views[name]; ok {
		app.view = view
		view.OnActivate()
	}
}

func (app *App) Run() error {
	go app.initClient()

	for app.running {
		app.draw()
		app.handleEvent()
	}
	return nil
}

func (app *App) Close() {
	app.vx.Close()
}

func (app *App) draw() {
	win := app.vx.Window()
	win.Clear()

	app.header.Draw(win)

	if app.view != nil {
		app.view.Draw(win)
	}

	if app.showQuit {
		app.quitModal.Draw(win)
	}

	app.footer.Draw(win)

	app.vx.Render()
}

func (app *App) SetLoading(loading bool) {
	app.loading = loading
	text := ""
	if loading {
		text = "Loading..."
	}
	app.footer.SetText(text)
	app.vx.PostEvent(vaxis.Redraw{})
}

func (app *App) initClient() {
	app.SetLoading(true)
	config := &mastodon.Config{
		Server:       app.config.Auth.Server,
		ClientID:     app.config.Auth.ClientID,
		ClientSecret: app.config.Auth.ClientSecret,
		AccessToken:  app.config.Auth.AccessToken,
	}

	client := mastodon.NewClient(config)

	if _, err := client.GetAccountCurrentUser(context.Background()); err != nil {
		app.Close()
		log.Fatalf("Failed to authenticate with Mastodon: %v", err)
	}

	app.client = client

	if app.view != nil {
		app.view.OnActivate()
	}
	app.SetLoading(false)
}

func (app *App) handleEvent() {
	event := app.vx.PollEvent()
	if event == nil {
		return
	}
	if key, ok := event.(vaxis.Key); ok {
		app.handleKeyEvent(key)
	}
}

func (app *App) handleKeyEvent(key vaxis.Key) {
	if app.showQuit {
		action := app.quitModal.HandleKey(key)
		switch action {
		case "quit":
			app.running = false
		case "close":
			app.showQuit = false
		}
		return
	}

	// Handle global keybindings
	if key.Matches('q') {
		app.showQuit = true
	} else {
		// Delegate keys to the current view
		if app.view != nil {
			app.view.HandleKey(key)
		}
	}
}
