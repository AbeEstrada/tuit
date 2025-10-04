package ui

import "git.sr.ht/~rockorager/vaxis"

type View interface {
	SetApp(app *App)
	OnActivate()
	Draw(win vaxis.Window)
	HandleKey(key vaxis.Key)
}
