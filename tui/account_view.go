package tui

import (
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/tuit/utils"
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

	width, _ := win.Size()
	y := 0

	avatarWidth := 24
	avatarHeight := 12
	avatarURL := account.AvatarStatic

	vxImage, cached := utils.ImageCache.Get(avatarURL, avatarWidth, avatarHeight)
	if cached {
		if width > avatarWidth {
			imgWin := win.New(0, y, avatarWidth, avatarHeight)
			vxImage.Draw(imgWin)
		}
	} else {
		utils.ImageCache.LoadAsync(avatarURL)
	}

	metaX := avatarWidth + 1
	metaWin := win.New(metaX, y, width-metaX, avatarHeight)
	metaWin.Println(
		0,
		vaxis.Segment{
			Text:  account.DisplayName,
			Style: vaxis.Style{Attribute: vaxis.AttrBold},
		},
	)
	metaWin.Println(
		1,
		vaxis.Segment{
			Text:  fmt.Sprintf("@%s", account.Acct),
			Style: vaxis.Style{Attribute: vaxis.AttrBold},
		},
	)
	if account.Bot {
		metaWin.Println(
			2,
			vaxis.Segment{
				Text: "Automated",
			},
		)
	}

	y += avatarHeight

}

func (v *AccountView) HandleKey(key vaxis.Key) {}
