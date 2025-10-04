package ui

import "git.sr.ht/~rockorager/vaxis"

type ProfileView struct {
	app *App
}

func CreateProfileView() *ProfileView {
	return &ProfileView{}
}

func (v *ProfileView) SetApp(app *App) {
	v.app = app
}

func (v *ProfileView) OnActivate() {}

func (v *ProfileView) Draw(win vaxis.Window) {
	win.Println(1, vaxis.Segment{
		Text: "Profile",
	})
}

func (v *ProfileView) HandleKey(key vaxis.Key) {}
