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

// Regex para capturar las etiquetas de color personalizadas
var colorTagRegex = regexp.MustCompile(`\[COLOR:(\w+)\](.*?)\[/COLOR\]`)

// Función para interpretar y convertir las etiquetas de color personalizadas en ui.Style
func parseColorTags(text string, defaultStyle ui.Style) ([]ui.Cell, error) {
	cells := []ui.Cell{}
	lastIndex := 0

	// Mapa de colores por nombre
	colorMap := map[string]ui.Color{
		"RED":     ui.ColorRed,
		"GREEN":   ui.ColorGreen,
		"YELLOW":  ui.ColorYellow,
		"BLUE":    ui.ColorBlue,
		"MAGENTA": ui.ColorMagenta,
		"CYAN":    ui.ColorCyan,
		"WHITE":   ui.ColorWhite,
	}

	// Encontrar coincidencias de las etiquetas de color
	matches := colorTagRegex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		// Agregar texto previo sin formato de color
		if lastIndex < match[0] {
			cells = append(cells, ui.ParseStyles(text[lastIndex:match[0]], defaultStyle)...)
		}

		// Extraer el color y el texto coloreado de la etiqueta
		colorName := text[match[2]:match[3]]
		coloredText := text[match[4]:match[5]]

		// Obtener el color del mapa
		color, exists := colorMap[colorName]
		if !exists {
			color = defaultStyle.Fg // Usar el color predeterminado si no se encuentra
		}

		// Crear estilo con el color encontrado
		newStyle := ui.NewStyle(color)
		cells = append(cells, ui.ParseStyles(coloredText, newStyle)...)

		// Actualizar el índice para continuar después de la etiqueta de cierre
		lastIndex = match[5]
	}

	// Agregar cualquier texto restante después de la última etiqueta
	if lastIndex < len(text) {
		cells = append(cells, ui.ParseStyles(text[lastIndex:], defaultStyle)...)
	}

	return cells, nil
}

func (t *TextBox) Draw(buffer *ui.Buffer) {

	t.Block.Draw(buffer)

	cells, err := parseColorTags(t.text, t.style)
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
