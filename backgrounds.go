package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/ganim8/v2"
)

type TowerBackground struct {
	anim     *ganim8.Animation
	animBack *ganim8.Animation
	flip     bool
	frame    int
	tower    *ebiten.Image
	viewport viewport
}

type viewport struct {
	X int
	Y int
}

func (v *viewport) move(posX, posY float64, tower *ebiten.Image) {
	s := tower.Bounds().Size()
	maxX16 := s.X * 16
	maxY16 := s.Y * 16

	v.X += int(posX)
	v.Y += int(posY)
	v.X %= maxX16
	v.Y %= maxY16
}

func (v *viewport) position() (int, int) {
	return v.X, v.Y
}

func NewBackground(towerImg *ebiten.Image) *TowerBackground {
	t := TowerBackground{}
	t.tower = towerImg
	t.viewport = viewport{}
	t.SetupAnimation()
	t.frame = 1
	return &t
}

func (t *TowerBackground) Flip(flip bool) {
	if t.flip != flip {

		t.flip = flip

		if !t.flip {
			t.anim.GoToFrame(16 - t.frame + 1)
			//fmt.Println(t.frame, "forward", t.flip, flip)
		} else {
			t.animBack.GoToFrame(16 - t.frame + 1)
			//fmt.Println(t.frame, "back", t.flip, flip)
		}
	}

}

func (t *TowerBackground) Update() {
	if !t.flip {
		//t.anim.GoToFrame(t.frame)
		t.anim.Update()
		t.frame = t.anim.Position()
	} else {
		//t.animBack.GoToFrame(t.frame)
		t.animBack.Update()
		t.frame = t.animBack.Position()
	}
}

func (t *TowerBackground) drawSegment(screen *ebiten.Image, opts *ganim8.DrawOptions) {
	if !t.flip {
		t.anim.Draw(screen, opts)
	} else {
		t.animBack.Draw(screen, opts)
	}
}

var (
	offset1, offset2 = float64(WORLD_HEIGTH - HALF_HEIGHT), float64(WORLD_HEIGTH - (HALF_HEIGHT + SCREEN_HEIGHT))
)

func (t *TowerBackground) Draw(world *ebiten.Image, playerPosX, playerPosY float64) {
	//t.viewport.move(playerPosX, playerPosY, t.tower)

	for i := range 15 {
		offset := 32 * i
		t.drawSegment(t.tower, ganim8.DrawOpts(0, float64(SCREEN_HEIGHT-offset)))
	}

	offset1 += GameSpeed
	offset2 += GameSpeed
	if offset1 >= float64(WORLD_HEIGTH+HALF_HEIGHT) {
		offset1 = float64(WORLD_HEIGTH - (HALF_HEIGHT + SCREEN_HEIGHT))
	}
	if offset2 >= float64(WORLD_HEIGTH+HALF_HEIGHT) {
		offset2 = float64(WORLD_HEIGTH - (HALF_HEIGHT + SCREEN_HEIGHT))
	}

	op1 := &ebiten.DrawImageOptions{}
	op1.GeoM.Translate(playerPosX-96, offset1)

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(playerPosX-96, offset2)
	world.DrawImage(t.tower, op1)
	world.DrawImage(t.tower, op2)

	//px, py := t.viewport.position()
	//offsetX, offsetY := float64(-px)/16, float64(-py)/16

	//draw tower repeatedly
	//const repeat = 3
	//w, h := t.tower.Bounds().Dx(), t.tower.Bounds().Dy()
	//for j := range repeat {
	//	for i := range repeat {
	//		op := &ebiten.DrawImageOptions{}
	//		op.GeoM.Translate(float64(w*i), float64(h*j))
	//		op.GeoM.Translate(offsetX, offsetY)
	//		world.DrawImage(t.tower, op)
	//	}
	//}
}

func (t *TowerBackground) SetupAnimation() {
	grid := ganim8.NewGrid(192, 32, AtlasW, AtlasH)

	t.anim = ganim8.New(Atlas, grid.Frames(1, "16-1"), time.Millisecond*30)
	t.animBack = ganim8.New(Atlas, grid.Frames(1, "1-16"), time.Millisecond*30)
}
