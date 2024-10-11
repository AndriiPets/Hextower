package main

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	rv "github.com/solarlune/resolv"
	"github.com/yohamta/ganim8/v2"
)

type ControlMode int

const (
	Jumping ControlMode = iota
	Flying
)

type Player struct {
	Object         *rv.Object
	Ypos           float64
	Speed          rv.Vector
	OnGround       *rv.Object
	IgnorePlatform *rv.Object
	Sprite         Sprite
	FacingRight    bool
	controls       ControlMode
	stuck          bool
	dead           bool
}

func (p *Player) PlayerUpdate() {

	if !p.dead {
		if p.controls == Jumping {
			p.Speed.Y += GRAVITY
		} else if p.controls == Flying {
			p.Speed.Y -= GameSpeed
		}

		p.stuck = false

		if ebiten.IsKeyPressed(ebiten.KeyRight) {
			p.Speed.X += PLAYER_ACCEL
			p.FacingRight = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyLeft) {
			p.Speed.X -= PLAYER_ACCEL
			p.FacingRight = false
		}

		if ebiten.IsKeyPressed(ebiten.KeyUp) && p.controls == Flying {
			if p.Object.Position.Y > (WORLD_HEIGTH-HALF_HEIGHT)+64 {
				p.Object.Position.Y -= GameSpeed
			}
		}

		if ebiten.IsKeyPressed(ebiten.KeyDown) && p.controls == Flying {
			if p.Object.Bottom() < WORLD_HEIGTH+100 {
				p.Object.Position.Y += GameSpeed
			}
		}

		//Apply friction and clamp speed
		Clamp(&p.Speed.X)
		if p.controls == Flying {
			Clamp(&p.Speed.Y)
		}

		//Check for jumping
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) && p.controls == Jumping {

			if ebiten.IsKeyPressed(ebiten.KeyDown) && p.OnGround != nil && p.OnGround.HasTags("platform") {

				p.IgnorePlatform = p.OnGround

			} else {

				if p.OnGround != nil {
					p.Speed.Y = -JMP_SPEED
				}

			}

		}

		//check for collisions

		//Horizontal
		dx := p.Speed.X

		if check := p.Object.Check(p.Speed.X, 0, "solid"); check != nil {

			dx = check.ContactWithCell(check.Cells[0]).X
			p.Speed.X = 0

		}

		p.Object.Position.X += dx

		//Vertical
		p.OnGround = nil

		dy := p.Speed.Y

		dy = math.Max(math.Min(dy, 16), -16)

		checkDistance := dy
		if dy >= 0 {
			checkDistance++
		}

		if check := p.Object.Check(0, checkDistance, "solid", "platform", "ramp"); check != nil {

			//check if we can slide horizontaly aka coyote time
			slide, slideOK := check.SlideAgainstCell(check.Cells[0], "solid")

			if dy < 0 && check.Cells[0].ContainsTags("solid") && slideOK && math.Abs(slide.X) <= 8 {
				p.Object.Position.X += slide.X
			} else {

				//Check platforms (possibly moving)
				if platforms := check.ObjectsByTags("platform"); len(platforms) > 0 {

					platform := platforms[0]

					if p.Object.Right() < platform.Position.X || p.Object.Position.X > platform.Right() {
						return
					}

					if p.Object.Position.Y-p.Object.Size.Y < platform.Position.Y {
						dy = check.ContactWithObject(platform).Y
						p.OnGround = platform
						//p.Speed.Y = 0

						p.dead = true
					}
				}

				//Check solid ground
				if solids := check.ObjectsByTags("solid"); len(solids) > 0 && (p.OnGround == nil || p.OnGround.Position.Y >= solids[0].Position.Y) {
					dy = check.ContactWithObject(solids[0]).Y
					p.Speed.Y = 0

					if solids[0].Position.Y > p.Object.Position.Y {
						p.OnGround = solids[0]
					}
				}

				if p.OnGround != nil {
					p.IgnorePlatform = nil
				}
			}
		}

		if p.controls == Jumping {
			p.Object.Position.Y += dy
		}
		p.Ypos = dy

		if p.Object.Position.X > TOWER_BOUNDS+TOWER_OFFSET {
			p.Object.Position.X = TOWER_OFFSET + TOWER_BOUNDS
			p.stuck = true
			//p.Object.Position.X = TOWER_OFFSET
			//fmt.Println("Teleport to left")
		}
		if p.Object.Position.X < TOWER_OFFSET {
			p.Object.Position.X = TOWER_OFFSET
			p.stuck = true
			//p.Object.Position.X = (TOWER_BOUNDS - p.Object.Size.X) + TOWER_OFFSET
			//fmt.Println("Teleport right")
		}
		p.Object.Update()

	}

}

func Clamp(speed *float64) {
	if *speed > PLAYER_FRICTION {
		*speed -= PLAYER_FRICTION
	} else if *speed < -PLAYER_FRICTION {
		*speed += PLAYER_FRICTION
	} else {
		*speed = 0
	}

	if *speed > MAX_SPEED {
		*speed = MAX_SPEED
	} else if *speed < -MAX_SPEED {
		*speed = -MAX_SPEED
	}
}

func NewPlayer(game *Game, pos Vec2) *Player {

	p := &Player{
		Object:      rv.NewObject(pos[0], pos[1], 16, 16),
		FacingRight: true,
		controls:    Flying,
	}

	p.Object.SetShape(rv.NewRectangle(0, 0, p.Object.Size.X, p.Object.Size.Y))
	game.space.Add(p.Object)

	grid := ganim8.NewGrid(16, 16, AtlasW, AtlasH, 192, 32)
	anim := ganim8.New(Atlas, grid.Frames("1-3", 1), time.Millisecond*60)

	p.Sprite = Sprite{
		Object:    p.Object,
		Layer:     BeforeTower,
		Animation: anim,
	}

	game.sprites[99] = &p.Sprite

	return p

}
