package tui

import (
	"fmt"
	"strings"

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

	width, height := win.Size()
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
	metaY := 0
	metaWin := win.New(metaX, metaY, width-metaX, avatarHeight)
	metaWin.Println(
		metaY,
		vaxis.Segment{
			Text:  account.DisplayName,
			Style: vaxis.Style{Attribute: vaxis.AttrBold},
		},
	)
	metaY += 1
	metaWin.Println(
		metaY,
		vaxis.Segment{
			Text:  fmt.Sprintf("@%s", account.Acct),
			Style: vaxis.Style{Attribute: vaxis.AttrBold},
		},
	)
	metaY += 1
	if account.Bot {
		metaWin.Println(
			metaY,
			vaxis.Segment{
				Text: "Automated",
			},
		)
		metaY += 1
	}
	metaWin.Println(
		metaY,
		vaxis.Segment{
			Text: fmt.Sprintf("Joined %s", account.CreatedAt.Local().Format("Jan 2, 2006")),
		},
	)
	metaY += 2
	metaWin.Println(
		metaY,
		vaxis.Segment{
			Text: fmt.Sprintf("%s posts", utils.FormatNumber(account.StatusesCount)),
		},
	)
	metaY += 1
	metaWin.Println(
		metaY,
		vaxis.Segment{
			Text: fmt.Sprintf("%s following", utils.FormatNumber(account.FollowingCount)),
		},
	)
	metaY += 1
	metaWin.Println(
		metaY,
		vaxis.Segment{
			Text: fmt.Sprintf("%s followers", utils.FormatNumber(account.FollowersCount)),
		},
	)
	metaY += 1

	y += avatarHeight

	fieldsWin := win.New(0, y, width, len(account.Fields))
	for i, field := range account.Fields {
		valueSegments := utils.ParseStatus(field.Value, nil)

		var flatText strings.Builder
		for _, seg := range valueSegments {
			clean := strings.ReplaceAll(seg.Text, "\n", " ")
			clean = strings.TrimSpace(clean)
			if clean != "" {
				flatText.WriteString(clean + " ")
			}
		}
		flatValue := strings.TrimSpace(flatText.String())

		verified := ""
		verifiedStyle := vaxis.Style{}
		if !field.VerifiedAt.IsZero() {
			verified = "✓ "
			verifiedStyle = vaxis.Style{Foreground: vaxis.IndexColor(2)}
		}

		valueStyle := vaxis.Style{}
		if utils.IsValidURL(flatValue) {
			valueStyle = vaxis.Style{
				Hyperlink:      flatValue,
				UnderlineStyle: vaxis.UnderlineSingle,
			}
		}

		fieldsWin.Println(
			i,
			vaxis.Segment{Text: verified, Style: verifiedStyle},
			vaxis.Segment{
				Text:  fmt.Sprintf("%s: ", field.Name),
				Style: vaxis.Style{Attribute: vaxis.AttrBold},
			},
			vaxis.Segment{Text: flatValue, Style: valueStyle},
		)
		y += 1
	}
	y += 1

	contentHeight := height - y
	contentWin := win.New(0, y, width, contentHeight)
	content := utils.ParseStatus(account.Note, nil)
	_, rows := contentWin.Wrap(content...)

	y += rows

}

func (v *AccountView) HandleKey(key vaxis.Key) {}
