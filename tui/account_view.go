package tui

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/mattn/go-mastodon"
)

type AccountView struct {
	app *App
}

func CreateAccountView() *AccountView {
	return &AccountView{}
}

func (v *AccountView) SetApp(app *App) {
	v.app = app
}

func (v *AccountView) Draw(win vaxis.Window, focused bool, account *mastodon.Account) {
	if account == nil {
		win.Println(0, vaxis.Segment{Text: ""})
		return
	}

	win.Println(0, vaxis.Segment{Text: account.Acct})

}

func (v *AccountView) HandleKey(key vaxis.Key) {}
