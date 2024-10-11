package main

import (
	"math/rand"
	"time"

	rv "github.com/solarlune/resolv"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"github.com/yohamta/ganim8/v2"
)

type Platform struct {
	Object *rv.Object
	Sprite Sprite
	used   bool
	pType  string
	tween  *gween.Sequence
}

type PlatformType int

const (
	PlatformNormal PlatformType = iota
	PlatformMoveHorizontal
)

func NewPlatform(game *Game, pos Vec2, tag string, pType PlatformType, inx int) *Platform {
	var sizeX, sizeY float64
	var tween *gween.Sequence
	var anim *ganim8.Animation

	switch pType {
	case PlatformNormal:
		sizeX, sizeY = 32, 16
		tween = gween.NewSequence()
		tween.Add(
			gween.New(float32(pos[1]), float32(pos[1]), 2, ease.Linear),
		)

		grid := ganim8.NewGrid(32, 16, AtlasW, AtlasH, 192, 16)
		anim = ganim8.New(Atlas, grid.Frames(1, "1-1"), time.Millisecond*30)

	case PlatformMoveHorizontal:
		sizeX, sizeY = 16, 16
		tween = gween.NewSequence()
		tween.Add(
			gween.New(float32(pos[0]), float32(pos[0]+128), 2, ease.Linear),
			gween.New(float32(pos[0]+128), float32(pos[0]), 2, ease.Linear),
		)

		grid := ganim8.NewGrid(16, 16, AtlasW, AtlasH, 192)
		anim = ganim8.New(Atlas, grid.Frames("1-2", 1), time.Second)
	}

	p := &Platform{
		Object: rv.NewObject(pos[0], pos[1], sizeX, sizeY, tag),
		pType:  tag,
	}
	//p.Object.SetShape(rv.NewRectangle(0, 0, p.Object.Size.X, p.Object.Size.Y))
	game.space.Add(p.Object)

	p.Sprite = Sprite{
		Object:    p.Object,
		Layer:     BeforeTower,
		Animation: anim,
	}

	p.tween = tween

	game.sprites[inx] = &p.Sprite
	//game.platforms = append(game.platforms, p)
	return p
}

func (p *Platform) Update(posY float64) {

	x, _, seqDone := p.tween.Update(1.0 / 60.0)
	//p.Object.Position.Y = float64(y)
	if seqDone {
		p.tween.Reset()
	}

	p.Object.Position.Y += GameSpeed
	p.Object.Position.X = float64(x)

	p.Object.Update()
}

type PlatformSpawner struct {
	Game      *Game
	Platforms []*Platform
}

func NewPlatformSpawner(game *Game, size int) *PlatformSpawner {
	ps := &PlatformSpawner{
		Platforms: make([]*Platform, size),
		Game:      game,
	}

	return ps
}

func (ps *PlatformSpawner) Spawn(pos Vec2, pType PlatformType, tags string) {
	for inx, p := range ps.Platforms {
		if p == nil || !p.used {
			platform := NewPlatform(ps.Game, pos, tags, pType, inx)
			platform.used = true
			ps.Platforms[inx] = platform
			return
		}
	}
}

func (ps *PlatformSpawner) Update() {

	var spawnAreaCount int

	for inx, p := range ps.Platforms {
		if p != nil && p.used {
			p.Update(ps.Game.player.Ypos)

			if p.Object.Position.Y < SCREEN_HEIGHT {
				spawnAreaCount++
			}

			if p.Object.Position.Y > ps.Game.player.Object.Bottom()+HALF_HEIGHT {
				ps.Release(inx)
				//fmt.Println("Platform destroyed", inx)
			}
		}
	}

	if spawnAreaCount < 1 {
		ps.Generate(15 + Difficulty)
		//ps.Game.fillPockets(SCREEN_HEIGHT)
	}
}

func (ps PlatformSpawner) Sweep() {
	for inx := range ps.Platforms {
		ps.Release(inx)
	}
}

// main random object generation function
func (ps *PlatformSpawner) Generate(ammount int) {
	boundX, boundY := 38, 30
	taken := make(map[Vec2_i]int)

	for i := range ammount {
		for range 3 { //attempt to find coordinates again if failed

			cx, cy := rand.Intn(boundX), rand.Intn(boundY)
			coord := Vec2_i{cx, cy}

			if checkCoords(taken, coord) {
				taken[coord] = i
				pos := Vec2{float64(TOWER_OFFSET + cx*TILE_SIZE), float64(cy * TILE_SIZE)}
				dice := rand.Intn(10)
				if dice < 3 {
					ps.Spawn(pos, PlatformMoveHorizontal, "platform")
				} else {

					ps.Spawn(pos, PlatformNormal, "platform")
				}
				break
			}
		}
	}
}

func checkCoords(taken map[Vec2_i]int, coord Vec2_i) bool {
	nearCells := []Vec2_i{
		{coord[0], coord[1]},
		{coord[0] + 1, coord[1]},
		{coord[0] - 1, coord[1]},
		{coord[0], coord[1] + 1},
		{coord[0], coord[1] - 1},
		{coord[0] + 1, coord[1] + 1},
		{coord[0] - 1, coord[1] - 1},
		{coord[0] + 1, coord[0] - 1},
		{coord[0] - 1, coord[0] + 1},
	}

	for _, c := range nearCells {
		_, ok := taken[c]
		if ok {
			return false
		}
	}

	return true
}

func (ps *PlatformSpawner) Release(inx int) {
	p := ps.Platforms[inx]
	if p != nil {
		ps.Game.space.Remove(p.Object)
		delete(ps.Game.sprites, inx)
		p.used = false
	}
}
