package tui

import (
	"fmt"
	"slices"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/tuit/utils"
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
		isBot = " · Automated"
	}

	metaWin.Println(
		0,
		vaxis.Segment{
			Text:  displayStatus.Account.DisplayName,
			Style: vaxis.Style{Attribute: vaxis.AttrBold},
		},
		vaxis.Segment{
			Text:  " (" + displayStatus.Account.Acct + ")",
			Style: vaxis.Style{Attribute: vaxis.AttrBold},
		},
		vaxis.Segment{Text: isBot},
	)

	timeLine := fmt.Sprintf("%s · %s", utils.FormatTimeSince(displayStatus.CreatedAt.Local()), utils.TitleCase(displayStatus.Visibility))
	metaWin.Println(1, vaxis.Segment{Text: timeLine})

	statsLine := fmt.Sprintf("%d replies · %d boosts · %d favorites", displayStatus.RepliesCount, displayStatus.ReblogsCount, displayStatus.FavouritesCount)
	metaWin.Println(2, vaxis.Segment{Text: statsLine})

	y = headerY + avatarHeight + 1

	contentHeight := height - y
	if contentHeight <= 0 {
		return
	}

	if displayStatus.Sensitive {
		win.Println(y, vaxis.Segment{Text: "⚠ Sensitive"})
		y += 2
	}

	contentWin := win.New(0, y, width, contentHeight)
	content := utils.ParseStatus(displayStatus.Content, displayStatus.Tags)
	_, rows := contentWin.Wrap(content...)

	contentY := rows

	if displayStatus.Poll != nil {
		poll := displayStatus.Poll

		contentWin.Println(contentY, vaxis.Segment{Text: "Poll"})
		contentY++

		indicator := "○"
		if poll.Multiple {
			indicator = "☐"
		}

		for i, option := range poll.Options {
			prefix := indicator
			if slices.Contains(poll.OwnVotes, i) {
				prefix = "✓"
			}

			var votes string
			if poll.Voted {
				percentage := 0.0
				if poll.VotersCount > 0 {
					percentage = float64(option.VotesCount) / float64(poll.VotersCount) * 100
				}
				votes = fmt.Sprintf("%.0f%%", percentage)
			}

			formatted := fmt.Sprintf("%s %s %s", prefix, votes, option.Title)
			contentWin.Println(contentY, vaxis.Segment{Text: formatted})
			contentY++
		}

		if poll.Voted {
			contentWin.Println(contentY, vaxis.Segment{Text: fmt.Sprintf("Total: %d", poll.VotersCount)})
			contentY++
		}

		contentY += 2
	}

	if displayStatus.Card != nil {
		card := displayStatus.Card

		if card.URL != "" {
			cardText := card.Title
			if cardText == "" {
				cardText = card.URL
			}

			contentWin.PrintTruncate(
				contentY,
				vaxis.Segment{Text: "↗ "},
				vaxis.Segment{
					Text: cardText,
					Style: vaxis.Style{
						Hyperlink:      card.URL,
						UnderlineStyle: vaxis.UnderlineSingle,
					},
				},
			)
			contentY += 2
		}

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
