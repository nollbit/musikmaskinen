package widget

import (
	"image"

	termui "github.com/gizak/termui/v3"
	"github.com/lukesampson/figlet/figletlib"
)

type FigletBanner struct {
	termui.Block
	Text       string
	TextStyle  termui.Style
	FigletFont *figletlib.Font
	FadeOffset int
}

func NewFigletBanner() *FigletBanner {
	return &FigletBanner{
		Block:     *termui.NewBlock(),
		TextStyle: termui.Theme.Paragraph.Text,
	}
}

var (
	fadeColors = []termui.Color{
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45,
		46, 45, 44, 43, 42, 41, 40, 39, 38, 37, 36, 35, 34, 33, 32,
		31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17,
	}
)

func (f *FigletBanner) Draw(buf *termui.Buffer) {
	f.Block.Draw(buf)

	settings := f.FigletFont.Settings()
	renderedText := figletlib.SprintMsg(f.Text, f.FigletFont, f.Inner.Max.X, settings, "center")

	//cells := termui.ParseStyles(renderedText, self.TextStyle)

	//rows := termui.SplitCells(cells, '\n')

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

	rows := make([][]termui.Cell, 0)
	row := make([]termui.Cell, 0, 100)
	colorIndex := f.FadeOffset
	for _, rune := range runes {
		if rune == '\n' {
			colorIndex = f.FadeOffset
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
