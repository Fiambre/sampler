package component

import (
	"github.com/sqshq/sampler/config"
	ui "github.com/sqshq/termui"
)

type Component struct {
	Type          config.ComponentType
	Drawable      ui.Drawable
	Title         string
	Position      config.Position
	Size          config.Size
	RefreshRateMs int
}

func (c *Component) Move(x, y int) {
	c.Position.X += x
	c.Position.Y += y
	c.normalize()
}

func (c *Component) Resize(x, y int) {
	c.Size.X += x
	c.Size.Y += y
	c.normalize()
}

func (c *Component) normalize() {
	if c.Size.X < 0 {
		c.Size.X = 0
	}
	if c.Size.Y < 0 {
		c.Size.Y = 0
	}
	if c.Position.X < 0 {
		c.Position.X = 0
	}
	if c.Position.Y < 0 {
		c.Position.Y = 0
	}
}