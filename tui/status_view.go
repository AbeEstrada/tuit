package tui

import (
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/mastty/utils"
	"github.com/mattn/go-mastodon"
)

type StatusView struct {
	app      *App
	statusID mastodon.ID
}

func CreateStatusView() *StatusView {
	return &StatusView{}
}

func (v *StatusView) SetApp(app *App) {
	v.app = app
}

func (v *StatusView) Draw(win vaxis.Window, focused bool, status *mastodon.Status) {
	if status == nil {
		win.Println(0, vaxis.Segment{Text: ""})
		return
	}

	if v.statusID != status.ID {
		v.statusID = status.ID
	}

	width, height := win.Size()
	y := 0
	displayStatus := status

	if status.Reblog != nil {
		win.Println(y, vaxis.Segment{
			Text: fmt.Sprintf("Boosted by @%s", status.Account.Acct),
		})
		y += 2
		displayStatus = status.Reblog
	} else if status.InReplyToID != nil {
		win.Println(y, vaxis.Segment{Text: "Continued thread"})
		y += 2
	}

	headerY := y
	avatarWidth := 6
	avatarHeight := 3
	avatarURL := displayStatus.Account.AvatarStatic

	vxImage, cached := utils.ImageCache.Get(avatarURL)
	if cached {
		if width > avatarWidth {
			imgWin := win.New(0, headerY, avatarWidth, avatarHeight)
			vxImage.Draw(imgWin)
		}
	} else {
		utils.ImageCache.LoadAsync(avatarURL, avatarWidth, avatarHeight)
	}

	metaX := avatarWidth + 1
	metaWin := win.New(metaX, headerY, width-metaX, avatarHeight)

	var isBot string
	if displayStatus.Account.Bot {
		isBot = "Automated: "
	}

	userLine := fmt.Sprintf("%s%s (@%s)", isBot, displayStatus.Account.DisplayName, displayStatus.Account.Acct)
	metaWin.Println(0, vaxis.Segment{Text: userLine, Style: vaxis.Style{Attribute: vaxis.AttrBold}})

	timeLine := fmt.Sprintf("%s · %s", utils.FormatTimeSince(displayStatus.CreatedAt.Local()), utils.TitleCase(displayStatus.Visibility))
	metaWin.Println(1, vaxis.Segment{Text: timeLine})

	statsLine := fmt.Sprintf("%d replies · %d boosts · %d favorites", displayStatus.RepliesCount, displayStatus.ReblogsCount, displayStatus.FavouritesCount)
	metaWin.Println(2, vaxis.Segment{Text: statsLine})

	y = headerY + avatarHeight + 1

	contentHeight := height - y
	if contentHeight <= 0 {
		return
	}

	contentWin := win.New(0, y, width, contentHeight)
	content := utils.ParseStatus(displayStatus.Content, displayStatus.Tags)
	_, rows := contentWin.Wrap(content...)

	contentY := rows

	if displayStatus.Card != nil {
		card := displayStatus.Card

		if card.Image != "" {
			imageURL := card.Image
			mediaWidth := width

			vxImage, cached := utils.ImageCache.Get(imageURL)
			if cached {
				_, mediaHeight := vxImage.CellSize()

				if contentY+mediaHeight < contentHeight {
					imgWin := contentWin.New(0, contentY, mediaWidth, mediaHeight)
					vxImage.Draw(imgWin)

					contentY += mediaHeight
				}
			} else {
				var calculatedHeight int

				if card.Width > 0 {
					aspectRatio := float64(card.Height) / float64(card.Width)
					calculatedHeight = int(float64(mediaWidth) * aspectRatio * 0.5)
					calculatedHeight = max(calculatedHeight, 1)
				} else {
					calculatedHeight = 10
				}

				utils.ImageCache.LoadAsync(imageURL, mediaWidth, calculatedHeight)
			}
		}

		contentY++
	}

	if len(displayStatus.MediaAttachments) > 0 {
		mediaWidth := width

		for i, media := range displayStatus.MediaAttachments {
			if i > 0 {
				contentY++
			}

			imageURL := media.PreviewURL
			if imageURL == "" {
				imageURL = media.URL
			}
			if imageURL == "" {
				continue
			}

			vxImage, cached := utils.ImageCache.Get(imageURL)
			if cached {
				_, mediaHeight := vxImage.CellSize()

				imgWin := contentWin.New(0, contentY, mediaWidth, mediaHeight)
				vxImage.Draw(imgWin)

				contentY += mediaHeight
			} else {
				var calculatedHeight int
				imageMeta := media.Meta.Original

				if imageMeta.Width > 0 {
					aspectRatio := float64(imageMeta.Height) / float64(imageMeta.Width)
					calculatedHeight = int(float64(mediaWidth) * aspectRatio * 0.5)
					calculatedHeight = max(calculatedHeight, 1)
				} else {
					calculatedHeight = 10
				}

				utils.ImageCache.LoadAsync(imageURL, mediaWidth, calculatedHeight)
			}
		}
		contentY++
	}

}

func (v *StatusView) HandleKey(key vaxis.Key) {}
