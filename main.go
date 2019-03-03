package main

import (
	"fmt"
	"log"
	"os"
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
	libraryPath      = kingpin.Flag("library", "Where's the music library?").Required().String()
	libraryIndexPath = kingpin.Flag("library-index", "Where's the music library index?").Default(".mm-index").String()
)

func formatLength(l int) string {
	mins := int(l / 60.0)
	secs := int(l) % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}

func titles(songs []*Song) []string {
	titles := make([]string, 0, len(songs))
	for _, song := range songs {
		title := fmt.Sprintf(" %s - %s (%s) ", song.Artist, song.Title, formatLength(song.Length))
		titles = append(titles, title)
	}
	return titles
}

func PlayAndExit() {

	statusChan := make(chan *SongStatus)
	abortChan := make(chan bool)
	go PlaySong("testdata/David_Szesztay_-_Cheese.mp3", statusChan, abortChan)

	for {
		s := <-statusChan
		if s.Err != nil {
			panic(s.Err)
		}

		fmt.Printf("Remaining: %d of %d Done? %v\n", s.Remaining, s.Length, s.Done)

		if s.Remaining == 28 {
			go func() {
				abortChan <- true
			}()
		}

		if s.Done {
			break
		}
	}

	os.Exit(1)
}

func main() {
	kingpin.Parse()

	//PlayAndExit()

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

	fmt.Print("Press 'Enter' to continue...")
	fmt.Scanln()

	songs := library.Songs
	titles := titles(songs)

	player := NewPlayer(7)

	uiHeader := widgets.NewParagraph()
	uiHeader.Text = header[1:]
	uiHeader.WrapText = false
	uiHeader.TextStyle = ui.NewStyle(ui.ColorRed)
	uiHeader.Border = false
	uiHeader.SetRect(0, 0, 110, 8)
	ui.Render(uiHeader)

	uiSongList := widgets.NewList()
	uiSongList.Title = "Songs"
	uiSongList.Rows = titles
	uiSongList.TextStyle = ui.NewStyle(ui.ColorYellow)
	uiSongList.SelectedRowStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow, ui.ModifierBold)
	uiSongList.WrapText = false

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

	uiSongInfo := widgets.NewParagraph()
	uiSongInfo.Title = "Current Song"
	uiSongInfo.Text = "\n [Artist](fg:blue,mod:bold): [Shout out Louds](fg:white,mod:bold)\n [Title](fg:blue,mod:bold):  [Sound is the Word](fg:white,mod:bold)\n [Album](fg:blue,mod:bold):  [Le Album](fg:white,mod:bold)"
	uiSongInfo.WrapText = false

	uiSongPlayerGauge := widgets.NewGauge()
	uiSongPlayerGauge.Title = "Playing"
	uiSongPlayerGauge.Percent = 30
	uiSongPlayerGauge.LabelStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack)
	uiSongPlayerGauge.Label = "Hello!"
	uiSongPlayerGauge.BarColor = ui.ColorBlue

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(1, 8, termWidth-1, termHeight-1)

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(0.4, uiSongList),
			ui.NewCol(0.6,
				ui.NewRow(0.2, uiSongInfo),
				ui.NewRow(0.1, uiSongPlayerGauge),
				ui.NewRow(0.7, uiQueueTable),
			),
		),
	)

	//go Run()

	ticker := time.NewTicker(time.Second / 30).C

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
				uiSongList.ScrollDown()
			case "j", "<Up>":
				uiSongList.ScrollUp()
			case "<Enter>":
				if !player.QueueFull() {
					currentlySelectedSong := songs[uiSongList.SelectedRow]
					player.QueueAdd(currentlySelectedSong)
				}
			}
		case <-ticker:
			ui.Render(grid)
		case qe := <-player.QueueEvents:
			{
				rows := [][]string{
					[]string{"   ", " Artist", " Title", " Dur.", " Wait"},
				}

				for i, qs := range qe.Queue {
					row := []string{
						fmt.Sprintf(" %d ", i+1),
						fmt.Sprintf(" %s ", qs.Song.Artist),
						fmt.Sprintf(" %s ", qs.Song.Title),
						fmt.Sprintf(" %s ", formatLength(qs.Song.Length)),
						fmt.Sprintf(" %s ", formatLength(qs.TimeUntilStart)),
					}
					rows = append(rows, row)
				}

				uiQueueTable.Rows = rows
			}
		}

	}
}
