package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nollbit/musikmaskinen/controller"

	"github.com/lukesampson/figlet/figletlib"
	log "github.com/sirupsen/logrus"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/nollbit/musikmaskinen/fonts"
	"github.com/nollbit/musikmaskinen/spotify"

	mmwidgets "github.com/nollbit/musikmaskinen/widgets"
	sp "github.com/nollbit/spotify"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	command      = kingpin.Command("run", "Run the player").Default()
	maxQueueSize = command.Flag("max-queue-size", "How many tracks can be enqueued?").Default("5").Int()
)

func formatLength(l int) string {
	mins := int(l / 60.0)
	secs := int(l) % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}

func titles(tracks []sp.FullTrack) []string {
	titles := make([]string, 0, len(tracks))
	for _, track := range tracks {
		title := fmt.Sprintf(" %s - %s (%s) ", track.Artists[0].Name, track.Name, formatLength(track.Duration/1000))
		titles = append(titles, title)
	}
	return titles
}

func main() {
	kingpin.Parse()

	font, err := figletlib.ReadFontFromBytes([]byte(fonts.AnsiShadow))
	if err != nil {
		log.Fatalf("Unable read font: %v", err)
	}

	file, err := os.OpenFile("mm.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Unable to create log file: %v", err)
	}
	log.SetLevel(log.DebugLevel)
	log.SetOutput(file)

	spotifyClient, err := spotify.GetClient()

	if err != nil {
		log.Fatalf("Unable to login: %v", err)
	}

	player, err := spotify.NewPlayer(spotifyClient, *maxQueueSize)
	if err != nil {
		log.Fatalf("Unable to create spotify player: %v", err)
	}

	devices, err := spotifyClient.PlayerDevices()
	if err != nil {
		log.Fatalf("Unable to get user devices: %v", err)
	}

	hasActiveDevice := false

	fmt.Println("Available devices: ")
	for i, device := range devices {
		active := ""
		if device.Active {
			hasActiveDevice = true
			active = "[active]"
		}
		fmt.Printf("(%d) %s (%s) %s\n", i, device.Name, device.Type, active)
	}

	if !hasActiveDevice {
		fmt.Println("No active spotify device")
		os.Exit(1)
	}

	// stop any current playback, ignore error
	spotifyClient.Pause()

	curatedPlaylist, err := spotify.NewCuratedPlaylist(spotifyClient, sp.ID(*spotify.SpotifyCuratedPlaylistID))
	if err != nil {
		log.WithError(err).Fatal("Unable to watch playlist")
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	cntrl, err := controller.NewController()
	if err != nil {
		log.WithError(err).Warn("Unable to open controller, disabling")
		// create dummy controller
		cntrl = controller.NewDummyController()
	}

	termWidth, termHeight := ui.TerminalDimensions()

	uiHeader := mmwidgets.NewFigletBanner()
	uiHeader.FigletFont = font
	uiHeader.TextStyle = ui.NewStyle(40)
	uiHeader.Border = false
	uiHeader.SetRect(0, 0, termWidth, 8)
	ui.Render(uiHeader)

	uiUsage := widgets.NewParagraph()
	uiUsage.Title = "Instruction"
	uiUsage.TitleStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	uiUsage.TextStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)

	uiTrackList := widgets.NewList()
	uiTrackList.Title = "Tracks"
	uiTrackList.TitleStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	uiTrackList.TextStyle = ui.NewStyle(ui.ColorYellow)
	uiTrackList.SelectedRowStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow, ui.ModifierBold)
	uiTrackList.WrapText = false

	uiQueueTable := widgets.NewTable()
	uiQueueTable.Rows = [][]string{
		[]string{"   ", " Track", " Dur.", " Wait"},
	}
	uiQueueTable.TextStyle = ui.NewStyle(ui.ColorWhite)
	uiQueueTable.RowSeparator = true
	uiQueueTable.FillRow = true
	uiQueueTable.Title = "Queue"
	uiQueueTable.TitleStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)

	uiQueueTable.ColumnResizer = func() {
		widthLeft := uiQueueTable.Inner.Dx() - 20
		uiQueueTable.ColumnWidths = []int{3, widthLeft, 6, 7}
	}

	uiTrackInfo := widgets.NewParagraph()
	uiTrackInfo.Title = "Current Track"
	uiTrackInfo.TitleStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	uiTrackInfo.Text = ""
	uiTrackInfo.WrapText = false

	uiTrackPlayerGauge := widgets.NewGauge()
	uiTrackPlayerGauge.Title = "Playing"
	uiTrackPlayerGauge.TitleStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	uiTrackPlayerGauge.Percent = 0
	uiTrackPlayerGauge.LabelStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack)
	uiTrackPlayerGauge.Label = "<3!"
	uiTrackPlayerGauge.BarColor = ui.ColorBlue

	grid := ui.NewGrid()
	grid.SetRect(1, 8, termWidth-1, termHeight-1)

	grid.Set(
		ui.NewRow(1.0,
			// left UI column
			ui.NewCol(0.5,
				ui.NewRow(0.2, uiUsage),
				ui.NewRow(0.8, uiTrackList),
			),
			// right UI column
			ui.NewCol(0.5,
				ui.NewRow(0.2, uiTrackInfo),
				ui.NewRow(0.1, uiTrackPlayerGauge), // progress bar for current song
				ui.NewRow(0.7, uiQueueTable),
			),
		),
	)

	ticker := time.NewTicker(time.Second / 30).C
	queueTicker := time.NewTicker(time.Second / 10).C
	bannerColorTicker := time.NewTicker(time.Second / 5).C
	bannerTextTicker := time.NewTicker(time.Second * 15).C
	curatedPlaylistTicker := time.NewTicker(time.Second * 15).C

	ui.Render(grid)

	// updates the instructions box based on queue status
	updateInstructions := func() {
		var sb strings.Builder

		sb.WriteString(" How to select a song:\n")
		sb.WriteString("  1. Move to the song with the [scroll wheel](fg:yellow,mod:bold)\n")
		sb.WriteString("  2. Push the [blinking button to the right](fg:yellow,mod:bold)\n")
		sb.WriteString("\n")

		if player.QueueFull() {
			sb.WriteString(" [ >>>>>>> The queue is now full. Please wait <<<<<<< ](fg:white,bg:red,mod:bold)\n")
		} else {
			sb.WriteString(fmt.Sprintf(" There can only be [%d](mod:bold) tracks in the queue. One per person please!\n", *maxQueueSize))
		}

		uiUsage.Text = sb.String()
	}
	updateInstructions()

	// triggered when queue is full
	queusIsNowFull := func() {
		// stop the blinking light on the controlleer
		err := cntrl.WriteCommand(controller.CommandLedOff)
		if err != nil {
			log.WithError(err).Errorf("Unable to send command to controller")
		}

		updateInstructions()
	}

	// triggered when queue is open!
	queusIsNowOpen := func() {
		// start the blinking light on the controlleer
		err := cntrl.WriteCommand(controller.CommandLedBlink)
		if err != nil {
			log.WithError(err).Errorf("Unable to send command to controller")
		}

		updateInstructions()
	}

	// update the header text
	headerTextIndex := 0
	updateHeaderText := func() {
		var headerText string

		switch headerTextIndex {
		case 0:
			headerText = "MUSIKMASKINEN"
		case 1:
			headerText = "RICKARD 40"
		case 2:
			{
				track := player.CurrentlyPlaying()
				if track != nil {
					rd := track.Album.ReleaseDate
					if len(rd) >= 4 {
						year := rd[0:4]
						headerText = fmt.Sprintf("PARTY LIKE ITS %s!", year)
					} else {
						headerText = "MUSIKMASKINEN"
					}
				} else {
					headerText = "MUSIKMASKINEN"
				}
			}
		}

		uiHeader.Text = headerText
		headerTextIndex = (headerTextIndex + 1) % 3
	}
	updateHeaderText()

	renderPlaylistTitles := func() {
		log.Debug("rendering titles")
		// format the tracks for UI

		currentlyPlayingTrack := player.CurrentlyPlaying()

		formattedTracks := make([]string, 0, len(curatedPlaylist.Tracks))
		for _, track := range curatedPlaylist.Tracks {
			_, isBlacklisted := curatedPlaylist.IsTrackBlacklisted(track.ID)
			var title string
			if currentlyPlayingTrack != nil && currentlyPlayingTrack.ID == track.ID {
				title = fmt.Sprintf(" [%s](fg:white) - [%s](fg:yellow) [(playing)](fg:white) ", track.Artists[0].Name, track.Name)
			} else if player.IsInQueue(track.ID) {
				title = fmt.Sprintf(" [%s](fg:white) - [%s](fg:yellow) [(in queue)](fg:white) ", track.Artists[0].Name, track.Name)
			} else if isBlacklisted {
				title = fmt.Sprintf(" [%s](fg:white) - [%s](fg:yellow) [(recently played)](fg:white) ", track.Artists[0].Name, track.Name)
			} else {
				title = fmt.Sprintf(" [%s](fg:white,mod:bold) - [%s](fg:yellow,mod:bold) [(%s)](fg:white) ", track.Artists[0].Name, track.Name, formatLength(track.Duration/1000))
			}
			formattedTracks = append(formattedTracks, title)
		}

		uiTrackList.Title = fmt.Sprintf(" %d tracks to choose from ", len(formattedTracks))
		uiTrackList.Rows = formattedTracks
	}

	queueSelectedTrack := func() {
		currentlySelectedTrack := curatedPlaylist.Tracks[uiTrackList.SelectedRow]

		_, isBlacklisted := curatedPlaylist.IsTrackBlacklisted(currentlySelectedTrack.ID)
		if isBlacklisted {
			return
		}

		if player.QueueFull() {
			return
		}

		curatedPlaylist.BlacklistTrack(currentlySelectedTrack.ID, 60*time.Minute)

		player.QueueAdd(currentlySelectedTrack)

		renderPlaylistTitles()
	}

	uiEvents := ui.PollEvents()
	for {
		select {
		case controllerCommand := <-cntrl.CommandEvents:
			{
				switch controllerCommand {
				case controller.EventCmdRotaryEncoderClockwise:
					uiTrackList.ScrollDown()
				case controller.EventCmdRotaryEncoderCounterClockwise:
					uiTrackList.ScrollUp()
				case controller.EventCmdPushButton:
					queueSelectedTrack()
				}
			}
		case controllerErr := <-cntrl.Errs:
			log.WithError(controllerErr).Fatal("Controller failure")
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				spotifyClient.Pause()
				return
			case "d":
				if !player.QueueEmpty() {
					player.QueueRemove()
				}
			case "k", "<Down>":
				uiTrackList.ScrollDown()
			case "j", "<Up>":
				uiTrackList.ScrollUp()
			case "<Enter>":
				queueSelectedTrack()
			case "s":
				player.Skip()
			}
		case <-ticker:
			ui.Render(grid)
		case <-bannerTextTicker:
			updateHeaderText()

		case <-bannerColorTicker:
			/*
				if uiHeader.TextStyle.Fg == 46 {
					uiHeader.TextStyle.Fg = 16
				} else {
					uiHeader.TextStyle.Fg += 1
				}*/
			uiHeader.FadeOffset = (uiHeader.FadeOffset + 1) % 200
			ui.Render(uiHeader)
		case <-player.QueueEvents:
			{
				if player.QueueFull() {
					queusIsNowFull()
				} else {
					queusIsNowOpen()
				}

			}
		case <-queueTicker:
			{

				rows := [][]string{
					[]string{"   ", " Track", " Dur.", " Wait"},
				}

				for i, qs := range player.GetQueue() {
					row := []string{
						fmt.Sprintf(" %d ", i+1),
						fmt.Sprintf(" [%s](fg:white,mod:bold) - [%s](fg:yellow,mod:bold)", qs.Track.Artists[0].Name, qs.Track.Name),
						fmt.Sprintf(" %s ", formatLength(qs.Track.Duration/1000)),
						fmt.Sprintf(" %s ", formatLength(qs.TimeUntilStart)),
					}
					rows = append(rows, row)
				}

				uiQueueTable.Rows = rows
			}
		case trackEvent := <-player.TrackEvents:
			// the periodic (>1 event per second) player update
			{
				//log.Debugf("Got trackEvent %v", trackEvent)
				var currentTrack string
				var gaugeLabel string
				var gaugePercent int

				if trackEvent.Done {
					currentTrack = ""
					gaugeLabel = ""
					gaugePercent = 0
				} else {
					s := trackEvent.Track

					template := `
					 [Artist](fg:blue,mod:bold):   [%s](fg:white,mod:bold)
					 [Title](fg:blue,mod:bold):    [%s](fg:white,mod:bold)
					 [Album](fg:blue,mod:bold):    [%s](fg:white,mod:bold)
					 [Release](fg:blue,mod:bold):  [%s](fg:white,mod:bold)`

					currentTrack = fmt.Sprintf(template, s.Artists[0].Name, s.Name, s.Album.Name, s.Album.ReleaseDate)
					gaugeLabel = formatLength(trackEvent.Remaining)
					gaugePercent = int((float32((s.Duration/1000)-trackEvent.Remaining) / float32(s.Duration/1000)) * 100)
				}

				uiTrackInfo.Text = currentTrack
				uiTrackPlayerGauge.Label = gaugeLabel
				uiTrackPlayerGauge.Percent = gaugePercent
			}
		case <-curatedPlaylist.Changes:
			// nope
			renderPlaylistTitles()
		case <-curatedPlaylistTicker:
			// the curated playlist changed
			renderPlaylistTitles()
		}

	}
}
