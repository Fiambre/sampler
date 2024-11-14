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
	cells := []ui.Cell{}

	// Solo como ejemplo: mapeo ANSI a colores básicos
	// Puedes agregar otros colores según necesites
	colorMap := map[string]ui.Color{
		"31": ui.ColorRed,
		"32": ui.ColorGreen,
		"33": ui.ColorYellow,
		"34": ui.ColorBlue,
		"35": ui.ColorMagenta,
		"36": ui.ColorCyan,
	}

	lastIndex := 0
	matches := ansiColorRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		if lastIndex < match[0] {
			cells = append(cells, ui.ParseStyles(text[lastIndex:match[0]], defaultStyle)...)
		}

		colorCode := ansiColorRegex.FindStringSubmatch(text[match[0]:match[1]])[1]
		color, exists := colorMap[colorCode]
		if !exists {
			color = defaultStyle.Fg
		}

		// Aplicamos el nuevo estilo de color encontrado
		newStyle := ui.NewStyle(color)
		end := match[1]
		nextTextStart := match[1]
		for nextTextStart < len(text) && text[nextTextStart] != '\033' {
			nextTextStart++
		}
		cells = append(cells, ui.ParseStyles(text[end:nextTextStart], newStyle)...)
		lastIndex = nextTextStart
	}

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
