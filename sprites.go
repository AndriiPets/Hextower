package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	rv "github.com/solarlune/resolv"
	"github.com/yohamta/ganim8/v2"
)

type Sprite struct {
	Object    *rv.Object
	Layer     LayerID
	Animation *ganim8.Animation
	DrawPos   Vec2
	Drawable  bool
	Behind    bool

	Color color.RGBA
}

func (s *Sprite) Update(g *Game) {
	leftEdge, rightEdge := g.player.Object.Position.X-(96+s.Object.Size.X+10), g.player.Object.Position.X+(96+s.Object.Size.X+10)
	s.DrawPos = Vec2{s.Object.Position.X, s.Object.Position.Y}
	s.Color = color.RGBA{225, 30, 60, 225}

	if s.Animation != nil {
		s.Animation.Update()
	}

	s.Layer = BeforeTower

	if s.Object.Position.X <= leftEdge {
		s.Layer = BehindTower
		s.Behind = true
		s.Color = color.RGBA{30, 225, 60, 225}

		offset := math.Abs(s.Object.Position.X - leftEdge)
		s.DrawPos = Vec2{leftEdge + offset, s.Object.Position.Y}

		if s.Object.Position.X <= leftEdge-TOWER_WIDTH {
			s.Layer = Invisible
			s.Color = color.RGBA{30, 30, 225, 225}
		}
	}

	if s.Object.Right() >= rightEdge {
		s.Layer = BehindTower
		s.Behind = true
		s.Color = color.RGBA{30, 225, 60, 225}

		offset := math.Abs(s.Object.Right() - rightEdge)
		s.DrawPos = Vec2{(rightEdge - offset) - s.Object.Size.X, s.Object.Position.Y}

		if s.Object.Right() >= rightEdge+TOWER_WIDTH {
			s.Layer = Invisible
			s.Color = color.RGBA{30, 30, 225, 225}
		}
	}

}

func (s *Sprite) Draw(screen *ebiten.Image) {
	if s.Animation != nil {

		s.Animation.Draw(screen, ganim8.DrawOpts(s.DrawPos[0], s.DrawPos[1]))
	}
}
