package textbox

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/sqshq/sampler/component"
	"github.com/sqshq/sampler/config"
	"github.com/sqshq/sampler/console"
	"github.com/sqshq/sampler/data"
	"image"
	"regexp"
	"strconv"
)

type TextBox struct {
	*ui.Block
	*data.Consumer
	alert  *data.Alert
	text   string
	border bool
	style  ui.Style
}

func NewTextBox(c config.TextBoxConfig, palette console.Palette) *TextBox {

	color := c.Color
	if color == nil {
		color = &palette.BaseColor
	}

	box := TextBox{
		Block:    component.NewBlock(c.Title, *c.Border, palette),
		Consumer: data.NewConsumer(),
		style:    ui.NewStyle(*color),
	}

	go func() {
		for {
			select {
			case sample := <-box.SampleChannel:
				box.text = sample.Value
			case alert := <-box.AlertChannel:
				box.alert = alert
			}
		}
	}()

	return &box
}

// Regex para capturar secuencias de escape ANSI
var ansiColorRegex = regexp.MustCompile(`\033\[(\d+);(\d+);(\d+)m`)

// Función para interpretar y convertir el código de color ANSI en ui.Style
func parseANSIColorCode(text string, defaultStyle ui.Style) ([]ui.Cell, error) {
	matches := ansiColorRegex.FindAllStringIndex(text, -1)
	cells := []ui.Cell{}

	lastIndex := 0
	for _, match := range matches {
		if lastIndex < match[0] {
			cells = append(cells, ui.ParseStyles(text[lastIndex:match[0]], defaultStyle)...)
		}

		// Parseamos los valores RGB de ANSI
		params := ansiColorRegex.FindStringSubmatch(text[match[0]:match[1]])
		r, _ := strconv.Atoi(params[1])
		g, _ := strconv.Atoi(params[2])
		b, _ := strconv.Atoi(params[3])

		// Configuramos el nuevo estilo
		newStyle := ui.NewStyle(ui.ColorRGB(uint8(r), uint8(g), uint8(b)))

		// El texto tras el código de color
		end := match[1]
		nextTextStart := match[1]
		for nextTextStart < len(text) && text[nextTextStart] != '\033' {
			nextTextStart++
		}
		cells = append(cells, ui.ParseStyles(text[end:nextTextStart], newStyle)...)
		lastIndex = nextTextStart
	}

	// Agrega el texto restante
	if lastIndex < len(text) {
		cells = append(cells, ui.ParseStyles(text[lastIndex:], defaultStyle)...)
	}

	return cells, nil
}

func (t *TextBox) Draw(buffer *ui.Buffer) {

	t.Block.Draw(buffer)

	cells, err := parseANSIColorCode(t.text, t.style)
	if err != nil {
		fmt.Println("Error parsing colors:", err)
		return
	}

	cells = ui.WrapCells(cells, uint(t.Inner.Dx()-2))
	rows := ui.SplitCells(cells, '\n')

	for y, row := range rows {
		if y+t.Inner.Min.Y >= t.Inner.Max.Y-1 {
			break
		}
		row = ui.TrimCells(row, t.Inner.Dx()-2)
		for _, cx := range ui.BuildCellWithXArray(row) {
			x, cell := cx.X, cx.Cell
			buffer.SetCell(cell, image.Pt(x+1, y+1).Add(t.Inner.Min))
		}
	}

	component.RenderAlert(t.alert, t.Rectangle, buffer)
}
