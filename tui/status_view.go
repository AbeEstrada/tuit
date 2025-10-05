package tui

import (
	"fmt"
	"strings"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/mastty/utils"
	"github.com/mattn/go-mastodon"
)

type StatusView struct {
	app      *App
	headerH  int
	contentH int
	scrollY  int
	statusID mastodon.ID
}

func CreateStatusView() *StatusView {
	return &StatusView{
		scrollY: 0,
	}
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
		v.scrollY = 0
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

	headerStartRow := y
	avatarWidth := 6
	avatarHeight := 3
	avatarURL := displayStatus.Account.AvatarStatic

	vxImage, cached := utils.ImageCache.Get(avatarURL)
	if cached {
		if width > avatarWidth {
			imgWin := win.New(0, headerStartRow, avatarWidth, avatarHeight)
			vxImage.Draw(imgWin)
		}
	} else {
		utils.ImageCache.LoadAsync(avatarURL, avatarWidth, avatarHeight)
	}

	metaCol := avatarWidth + 1
	if width > metaCol {
		metaWin := win.New(metaCol, headerStartRow, width-metaCol, avatarHeight)

		userLine := fmt.Sprintf("%s (@%s)", displayStatus.Account.DisplayName, displayStatus.Account.Acct)
		metaWin.Println(0, vaxis.Segment{Text: userLine, Style: vaxis.Style{Attribute: vaxis.AttrBold}})

		timeLine := fmt.Sprintf("%s · %s", utils.FormatTimeSince(displayStatus.CreatedAt.Local()), utils.TitleCase(displayStatus.Visibility))
		metaWin.Println(1, vaxis.Segment{Text: timeLine})

		statsLine := fmt.Sprintf("%d replies · %d boosts · %d favorite", displayStatus.RepliesCount, displayStatus.ReblogsCount, displayStatus.FavouritesCount)
		metaWin.Println(2, vaxis.Segment{Text: statsLine})
	}

	y = headerStartRow + avatarHeight + 1
	v.headerH = y

	contentAreaHeight := height - v.headerH
	if contentAreaHeight <= 0 {
		return
	}

	contentWin := win.New(0, v.headerH, width, contentAreaHeight)
	contentY := 0

	content := utils.HTMLToPlainText(displayStatus.Content)

	for paragraph := range strings.SplitSeq(content, "\n") {
		if strings.TrimSpace(paragraph) == "" {
			contentY++
			continue
		}
		wrapped := utils.WrapText(paragraph, width)
		for _, line := range wrapped {
			if contentY >= v.scrollY && contentY-v.scrollY < contentAreaHeight {
				contentWin.Println(contentY-v.scrollY, vaxis.Segment{Text: line})
			}
			contentY++
		}
	}

	if len(displayStatus.MediaAttachments) > 0 {
		contentY++
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

				if contentY >= v.scrollY && contentY-v.scrollY < contentAreaHeight {
					imgWin := contentWin.New(0, contentY-v.scrollY, mediaWidth, mediaHeight)
					vxImage.Draw(imgWin)
				}

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

	v.contentH = contentY
}

func (v *StatusView) HandleKey(key vaxis.Key) {
	_, height := v.app.vx.Window().Size()
	contentAreaHeight := height - v.headerH

	if key.Matches('j') {
		if v.scrollY+contentAreaHeight < v.contentH {
			v.scrollY++
			v.app.vx.PostEvent(vaxis.Redraw{})
		}
	} else if key.Matches('k') {
		if v.scrollY > 0 {
			v.scrollY--
			v.app.vx.PostEvent(vaxis.Redraw{})
		}
	}
}
