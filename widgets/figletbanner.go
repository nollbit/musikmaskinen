package widget

import (
	"image"

	termui "github.com/gizak/termui/v3"
	"github.com/lukesampson/figlet/figletlib"
)

// FigletBanner is an animated header that uses figlet fonts to render the color
// faded header text. Call Tick() to animate it.
type FigletBanner struct {
	termui.Block
	Text       string
	TextStyle  termui.Style
	FigletFont *figletlib.Font
	fadeOffset int
}

func NewFigletBanner() *FigletBanner {
	return &FigletBanner{
		Block:     *termui.NewBlock(),
		TextStyle: termui.Theme.Paragraph.Text,
	}
}

var (
	fadeColors []termui.Color
)

func init() {
	fadeColors = make([]termui.Color, 0)

	for i := 16; i < 51; i++ {
		fadeColors = append(fadeColors, termui.Color(i))
	}
	for i := 195; i > 161; i-- {
		fadeColors = append(fadeColors, termui.Color(i))
	}
	for i := 50; i < 160; i++ {
		fadeColors = append(fadeColors, termui.Color(i))
	}
	for i := 50; i > 15; i-- {
		fadeColors = append(fadeColors, termui.Color(i))
	}

}

func (f *FigletBanner) Tick() {
	f.fadeOffset = (f.fadeOffset + 1) % len(fadeColors)
}

func (f *FigletBanner) Draw(buf *termui.Buffer) {
	f.Block.Draw(buf)

	settings := f.FigletFont.Settings()
	renderedText := figletlib.SprintMsg(f.Text, f.FigletFont, f.Inner.Max.X, settings, "center")

	rows := f.cyclicHoriFade(renderedText)

	for y, row := range rows {
		if y+f.Inner.Min.Y >= f.Inner.Max.Y {
			break
		}
		row = termui.TrimCells(row, f.Inner.Dx())
		for _, cx := range termui.BuildCellWithXArray(row) {
			x, cell := cx.X, cx.Cell
			buf.SetCell(cell, image.Pt(x, y).Add(f.Inner.Min))
		}
	}
}

func (f *FigletBanner) cyclicHoriFade(s string) [][]termui.Cell {
	runes := []rune(s)

	fadeOffset := f.fadeOffset

	rows := make([][]termui.Cell, 0)
	row := make([]termui.Cell, 0, 100)
	colorIndex := fadeOffset
	for _, rune := range runes {
		if rune == '\n' {
			colorIndex = fadeOffset + len(rows)
			rows = append(rows, row)
			row = make([]termui.Cell, 0, 100)
		} else {
			colorIndex++
			cell := termui.Cell{
				Rune:  rune,
				Style: termui.NewStyle(fadeColors[colorIndex%len(fadeColors)]),
			}
			row = append(row, cell)
		}

	}

	if len(row) > 0 {
		rows = append(rows, row)
	}

	return rows
}
