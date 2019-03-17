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
}

func NewFigletBanner() *FigletBanner {
	return &FigletBanner{
		Block:     *termui.NewBlock(),
		TextStyle: termui.Theme.Paragraph.Text,
	}
}

func (self *FigletBanner) Draw(buf *termui.Buffer) {
	self.Block.Draw(buf)

	settings := self.FigletFont.Settings()
	renderedText := figletlib.SprintMsg(self.Text, self.FigletFont, self.Inner.Max.X, settings, "center")

	cells := termui.ParseStyles(renderedText, self.TextStyle)

	rows := termui.SplitCells(cells, '\n')

	for y, row := range rows {
		if y+self.Inner.Min.Y >= self.Inner.Max.Y {
			break
		}
		row = termui.TrimCells(row, self.Inner.Dx())
		for _, cx := range termui.BuildCellWithXArray(row) {
			x, cell := cx.X, cx.Cell
			buf.SetCell(cell, image.Pt(x, y).Add(self.Inner.Min))
		}
	}
}
