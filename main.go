package main

import (
	"fmt"
	"log"
	"time"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"gopkg.in/alecthomas/kingpin.v2"
)

const header = `
 ███╗   ███╗██╗   ██╗███████╗██╗██╗  ██╗███╗   ███╗ █████╗ ███████╗██╗  ██╗██╗███╗   ██╗███████╗███╗   ██╗
 ████╗ ████║██║   ██║██╔════╝██║██║ ██╔╝████╗ ████║██╔══██╗██╔════╝██║ ██╔╝██║████╗  ██║██╔════╝████╗  ██║
 ██╔████╔██║██║   ██║███████╗██║█████╔╝ ██╔████╔██║███████║███████╗█████╔╝ ██║██╔██╗ ██║█████╗  ██╔██╗ ██║
 ██║╚██╔╝██║██║   ██║╚════██║██║██╔═██╗ ██║╚██╔╝██║██╔══██║╚════██║██╔═██╗ ██║██║╚██╗██║██╔══╝  ██║╚██╗██║
 ██║ ╚═╝ ██║╚██████╔╝███████║██║██║  ██╗██║ ╚═╝ ██║██║  ██║███████║██║  ██╗██║██║ ╚████║███████╗██║ ╚████║
 ╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═══╝`

var (
	pauseAfterLibraryScan = kingpin.Flag("pause", "Pause after library scan").Default("false").Bool()
	libraryPath           = kingpin.Flag("library", "Where's the music library?").Required().String()
	libraryIndexPath      = kingpin.Flag("library-index", "Where's the music library index?").Default(".mm-index").String()
	maxQueueSize          = kingpin.Flag("max-queue-size", "How many tracks can be enqueued?").Default("5").Int()
)

func formatLength(l int) string {
	mins := int(l / 60.0)
	secs := int(l) % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}

func titles(tracks []*Track) []string {
	titles := make([]string, 0, len(tracks))
	for _, track := range tracks {
		title := fmt.Sprintf(" %s - %s (%s) ", track.Artist, track.Title, formatLength(track.Length))
		titles = append(titles, title)
	}
	return titles
}

func main() {
	kingpin.Parse()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	library, err := NewLibrary(*libraryIndexPath)
	if err != nil {
		panic(err)
	}
	err = library.Add(*libraryPath)
	if err != nil {
		panic(err)
	}

	library.Sort()
	err = library.WriteIndex()
	if err != nil {
		panic(err)
	}

	if *pauseAfterLibraryScan {
		fmt.Print("Press 'Enter' to continue...")
		fmt.Scanln()
	}

	tracks := library.Tracks
	titles := titles(tracks)

	player, err := NewPlayer(*maxQueueSize)
	defer player.Close()

	if err != nil {
		panic(err)
	}

	uiHeader := widgets.NewParagraph()
	uiHeader.Text = header[1:]
	uiHeader.WrapText = false
	uiHeader.TextStyle = ui.NewStyle(40)
	uiHeader.Border = false
	uiHeader.SetRect(0, 0, 110, 8)
	ui.Render(uiHeader)

	uiUsage := widgets.NewParagraph()
	uiUsage.Title = "Instruction"
	uiUsage.Text = fmt.Sprintf(" Select a track using the rotary wheel before you!\n Press it to queue the selected track \n Only %d tracks can be queued at a time", *maxQueueSize)
	uiHeader.TextStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)

	uiTrackList := widgets.NewList()
	uiTrackList.Title = "Tracks"
	uiTrackList.Rows = titles
	uiTrackList.TextStyle = ui.NewStyle(ui.ColorYellow)
	uiTrackList.SelectedRowStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow, ui.ModifierBold)
	uiTrackList.WrapText = false

	uiQueueTable := widgets.NewTable()
	uiQueueTable.Rows = [][]string{
		[]string{"   ", " Artist", " Title", " Dur.", " Wait"},
	}
	uiQueueTable.TextStyle = ui.NewStyle(ui.ColorWhite)
	uiQueueTable.RowSeparator = true
	uiQueueTable.FillRow = true

	uiQueueTable.ColumnResizer = func() {
		widthLeft := uiQueueTable.Inner.Dx() - 20
		columnWidth := widthLeft / 2

		uiQueueTable.ColumnWidths = []int{3, columnWidth, columnWidth, 6, 7}
	}

	uiTrackInfo := widgets.NewParagraph()
	uiTrackInfo.Title = "Current Track"
	uiTrackInfo.Text = ""
	uiTrackInfo.WrapText = false

	uiTrackPlayerGauge := widgets.NewGauge()
	uiTrackPlayerGauge.Title = "Playing"
	uiTrackPlayerGauge.Percent = 30
	uiTrackPlayerGauge.LabelStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack)
	uiTrackPlayerGauge.Label = "Hello!"
	uiTrackPlayerGauge.BarColor = ui.ColorBlue

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(1, 8, termWidth-1, termHeight-1)

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(0.4,
				ui.NewRow(0.2, uiUsage),
				ui.NewRow(0.8, uiTrackList),
			),
			ui.NewCol(0.6,
				ui.NewRow(0.2, uiTrackInfo),
				ui.NewRow(0.1, uiTrackPlayerGauge),
				ui.NewRow(0.7, uiQueueTable),
			),
		),
	)

	//go Run()

	ticker := time.NewTicker(time.Second / 30).C
	queueRefresh := time.NewTicker(time.Second / 10).C
	bannerColorTicker := time.NewTicker(time.Second / 5).C

	ui.Render(grid)

	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
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
				if !player.QueueFull() {
					currentlySelectedTrack := tracks[uiTrackList.SelectedRow]
					player.QueueAdd(currentlySelectedTrack)
				}
			case "s":
				player.Skip()
			}
		case <-ticker:
			ui.Render(grid)
		case <-bannerColorTicker:
			if uiHeader.TextStyle.Fg == 46 {
				uiHeader.TextStyle.Fg = 16
			} else {
				uiHeader.TextStyle.Fg += 1
			}
			ui.Render(uiHeader)
		case <-queueRefresh:
			{

				rows := [][]string{
					[]string{"   ", " Artist", " Title", " Dur.", " Wait"},
				}

				for i, qs := range player.GetQueue() {
					row := []string{
						fmt.Sprintf(" %d ", i+1),
						fmt.Sprintf(" %s ", qs.Track.Artist),
						fmt.Sprintf(" %s ", qs.Track.Title),
						fmt.Sprintf(" %s ", formatLength(qs.Track.Length)),
						fmt.Sprintf(" %s ", formatLength(qs.TimeUntilStart)),
					}
					rows = append(rows, row)
				}

				uiQueueTable.Rows = rows
			}
		case trackEvent := <-player.TrackEvents:
			{
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
					 [Artist](fg:blue,mod:bold): [%s](fg:white,mod:bold)
					 [Title](fg:blue,mod:bold):  [%s](fg:white,mod:bold)
					 [Album](fg:blue,mod:bold):  [%s](fg:white,mod:bold)
					 [Year](fg:blue,mod:bold):   [%s](fg:white,mod:bold)`

					currentTrack = fmt.Sprintf(template, s.Artist, s.Title, s.Album, s.Year)
					gaugeLabel = formatLength(trackEvent.Remaining)
					gaugePercent = int((float32(s.Length-trackEvent.Remaining) / float32(s.Length)) * 100)
				}

				uiTrackInfo.Text = currentTrack
				uiTrackPlayerGauge.Label = gaugeLabel
				uiTrackPlayerGauge.Percent = gaugePercent
			}
		}

	}
}
