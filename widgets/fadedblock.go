package widget

import (
	"image"

	termui "github.com/gizak/termui/v3"
)

type FadedBlock struct {
	termui.Block
}

// Faded block is a block filled with faded blocks. Try it and you'll see :)
// I only used it render something to put on the controller
func NewFadedBlock() *FadedBlock {
	return &FadedBlock{
		Block: *termui.NewBlock(),
	}
}

func (f *FadedBlock) Draw(buf *termui.Buffer) {
	f.Block.Draw(buf)

	rows := f.cyclicHoriFade()

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

func (f *FadedBlock) cyclicHoriFade() [][]termui.Cell {
	fadeColors := make([]termui.Color, 0)

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

	w := f.Inner.Max.X
	h := f.Inner.Max.Y

	block := '▉'

	rows := make([][]termui.Cell, h)
	for y := 0; y < h; y++ {
		row := make([]termui.Cell, w+1)
		for x := 0; x < w; x++ {
			row[x] = termui.Cell{
				Rune:  block,
				Style: termui.NewStyle(fadeColors[(x+y)%len(fadeColors)]),
			}
		}

		row[w] = termui.Cell{
			Rune: '\n',
		}
		rows[y] = row
	}

	return rows
}
