package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	rv "github.com/solarlune/resolv"
	"golang.org/x/image/font"
)

type Game struct {
	space           *rv.Space
	background      *TowerBackground
	player          *Player
	camera          Camera
	world           *ebiten.Image
	tower           *ebiten.Image
	sprites         map[int]*Sprite
	platformSpawner *PlatformSpawner
	debug           bool
	score           float64
	font            font.Face
	title           bool
}

var GameSpeed = 2.0

type LayerID int

const (
	Background LayerID = iota
	BehindTower
	Tower
	BeforeTower
	UI
	Invisible
)

//go:embed assets/*
var assets embed.FS

//go:embed assets/excel.ttf
var exelFont []byte

func readImage(file string) image.Image {
	b, err := assets.ReadFile(file)
	if err != nil {
		panic(fmt.Sprintf("Cannot find a file %s", file))
	}
	return bytes2Image(&b)
}

func bytes2Image(raw *[]byte) image.Image {
	img, format, err := image.Decode(bytes.NewReader(*raw))
	if err != nil {
		log.Fatal("Byte2Image Failed:", format, err)
	}

	return img
}

var (
	startPos   = Vec2{400, WORLD_HEIGTH - 32}
	Atlas      *ebiten.Image
	AtlasW     = 384
	AtlasH     = 512
	Difficulty = 0
	Font       font.Face
	FontBig    font.Face
)

func init() {
	Atlas = ebiten.NewImageFromImage(readImage("assets/tile_atlas.png"))
	fontData, _ := truetype.Parse(exelFont)
	opts := &truetype.Options{
		Size:    18,
		DPI:     72,
		Hinting: font.HintingFull,
	}
	optsBig := &truetype.Options{
		Size:    36,
		DPI:     72,
		Hinting: font.HintingFull,
	}
	Font = truetype.NewFace(fontData, opts)
	FontBig = truetype.NewFace(fontData, optsBig)
}

func NewGame() *Game {
	g := &Game{}
	g.space = rv.NewSpace(WORLD_WIDTH, WORLD_HEIGTH+HALF_HEIGHT, 16, 16)
	g.sprites = make(map[int]*Sprite)
	g.platformSpawner = NewPlatformSpawner(g, 100)
	g.score = 0

	g.title = true

	player := NewPlayer(g, startPos)
	g.player = player

	g.world = ebiten.NewImage(WORLD_WIDTH, WORLD_HEIGTH+HALF_HEIGHT) //circumference of the tower
	g.tower = ebiten.NewImage(TOWER_WIDTH, SCREEN_HEIGHT+HALF_HEIGHT)
	g.camera = Camera{
		ViewPort:   Vec2{SCREEN_WIDTH, SCREEN_HEIGHT},
		ZoomFactor: 48,
		Position:   startPos,
	}

	//platforms
	//g.platformSpawner.Spawn(Vec2{400, 400}, "platform")
	//g.platformSpawner.Spawn(Vec2{336, 350}, "platform")
	//g.platformSpawner.Spawn(Vec2{100, 400}, "platform")
	//g.space.Add(
	//	rv.NewObject(0, WORLD_HEIGTH-16, WORLD_WIDTH, 16),
	//)

	g.fillPockets(WORLD_HEIGTH)
	g.background = NewBackground(g.tower)

	return g

}

// fill sides with objects for smoth rotation
func (g *Game) fillPockets(yOffset float64) {
	pocketLeft := Vec2{TOWER_OFFSET, TOWER_OFFSET * 2}
	pocketRight := Vec2{(TOWER_BOUNDS + TOWER_OFFSET) - TOWER_OFFSET, TOWER_BOUNDS + TOWER_OFFSET}
	fmt.Println(pocketLeft, pocketRight)
	for _, p := range g.platformSpawner.Platforms {

		if p != nil {

			if p.Object.Position.X >= pocketLeft[0] && p.Object.Position.X <= pocketLeft[1] && p.Object.Position.Y < yOffset {
				if p.Object.Position.Y < yOffset {
					fmt.Println("Object in upper left poket")
				}
				offset := math.Abs(TOWER_OFFSET - p.Object.Position.X)
				g.platformSpawner.Spawn(Vec2{TOWER_OFFSET + TOWER_BOUNDS + offset, p.Object.Position.Y}, PlatformNormal, "platform")
			}

			if p.Object.Position.X >= pocketRight[0] && p.Object.Position.X <= pocketRight[1] && p.Object.Position.Y < yOffset {
				fmt.Println("Right pocket has a platform", pocketRight)
				offset := math.Abs((TOWER_OFFSET + TOWER_BOUNDS) - p.Object.Position.X)
				g.platformSpawner.Spawn(Vec2{TOWER_OFFSET - offset, p.Object.Position.Y}, PlatformNormal, "platform")
			}
		}
	}
}

func (g *Game) Restart() {
	GameSpeed = 2.0
	g.score = 0.0
	g.player.dead = false
	g.platformSpawner.Sweep()
	g.player.Object.Position.Y = startPos[1]
}

func (g *Game) RaiseDiff() {
	if int(g.score)%20 == 0 && int(g.score) > 0 {
		GameSpeed += 0.3
		Difficulty++
		g.score++
		return
	}
}
func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.debug = !g.debug
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if g.player.dead || g.title {
			g.Restart()
			g.title = false
		}
	}

	if g.player.Speed.X != 0 && !g.player.stuck && !g.player.dead {

		if !g.player.FacingRight {
			g.background.Flip(true)
			g.player.Sprite.Animation.Sprite().SetFlipH(true)
		} else {
			g.background.Flip(false)
			g.player.Sprite.Animation.Sprite().SetFlipH(false)
		}

		g.background.Update()
	}

	if !g.player.dead && !g.title {

		g.platformSpawner.Update()

		g.score += GameSpeed / 60
		//fmt.Println(int(g.score))

		g.RaiseDiff()
	}

	g.player.PlayerUpdate()
	if g.player.dead {
		GameSpeed = 0.0
	}

	for _, s := range g.sprites {
		s.Update(g)
	}

	playerPos := Vec2{g.player.Object.Position.X, g.player.Object.Position.Y}
	g.camera.Update(playerPos)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Clear()
	g.world.Clear()
	g.tower.Clear()

	//ebitenutil.DebugPrint(screen, strconv.Itoa(g.background.frame))
	for _, s := range g.sprites {
		if s.Object != nil {
			if s.Layer == BehindTower {
				//obj := s.Object
				//fmt.Println(s.Color, s.DrawPos)
				//vector.DrawFilledRect(g.world, float32(s.DrawPos[0]), float32(s.DrawPos[1]), float32(obj.Size.X), float32(obj.Size.Y), s.Color, false)
				s.Draw(g.world)

			}
		}
	}

	g.background.Draw(g.world, g.player.Object.Position.X, g.player.Object.Position.Y)

	for _, s := range g.sprites {
		if s.Object != nil {
			if s.Layer == BeforeTower {
				//obj := s.Object
				//fmt.Println(s.Color, s.DrawPos)
				//vector.DrawFilledRect(g.world, float32(s.DrawPos[0]), float32(s.DrawPos[1]), float32(obj.Size.X), float32(obj.Size.Y), s.Color, false)
				if !g.title {
					s.Draw(g.world)
				}

			}
		}
	}

	//worldX, worldY := g.camera.ScreenToWorld(g.player.Object.CellPosition())
	//ebitenutil.DebugPrint(
	//	screen,
	//	strconv.FormatFloat(ebiten.ActualFPS(), 'f', 1, 64),
	//)

	//ebitenutil.DebugPrintAt(
	//	screen,
	//	fmt.Sprintf("%s\nCursor World Pos: %.2f, %.2f", g.camera.String(), worldX, worldY),
	//	0, SCREEN_HEIGHT-32,
	//)

	if g.debug {
		g.DebugDraw(g.world)
	}

	g.camera.Render(g.world, screen)

	//UI

	if !g.player.dead && !g.title {
		g.DrawText(screen, 16, 16, Font, "Score: ", fmt.Sprintf("%d", int(g.score)))
	}

	if g.player.dead {
		g.DrawText(screen,
			170,
			HALF_HEIGHT-128,
			FontBig,
			"+++YOU DIED!+++", "", fmt.Sprintf("++Final Score: %d++", int(g.score)), "", "press R to restart")
	}

	if g.title {
		g.DrawText(
			screen,
			170,
			HALF_HEIGHT-128,
			FontBig,
			"+++HEXTOWER+++", "+++HEXTOWER+++", "+++HEXTOWER+++", "+++HEXTOWER+++", "+++HEXTOWER+++",
		)
	}

}

func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	return SCREEN_WIDTH, SCREEN_HEIGHT
}

func main() {
	ebiten.SetVsyncEnabled(true)
	ebiten.SetWindowSize(SCREEN_WIDTH, SCREEN_HEIGHT)
	ebiten.SetWindowTitle("HEXTOWER")
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}

func (g *Game) DrawText(screen *ebiten.Image, x, y int, fnt font.Face, textLines ...string) {
	rectHeight := 38
	for _, txt := range textLines {
		w := float64(font.MeasureString(fnt, txt).Round())
		ebitenutil.DrawRect(screen, float64(x), float64(y-rectHeight+10), w, float64(rectHeight), color.RGBA{0, 0, 0, 255})

		text.Draw(screen, txt, fnt, x+1, y+1, color.RGBA{255, 255, 255, 255})
		text.Draw(screen, txt, fnt, x, y, color.RGBA{0, 0, 0, 0})
		y += rectHeight
	}
}

func (g *Game) DebugDraw(screen *ebiten.Image) {

	space := g.space

	for y := 0; y < space.Height(); y++ {

		for x := 0; x < space.Width(); x++ {

			cell := space.Cell(x, y)

			cw := float32(space.CellWidth)
			ch := float32(space.CellHeight)
			cx := float32(cell.X) * cw
			cy := float32(cell.Y) * ch

			drawColor := color.RGBA{20, 20, 20, 255}

			if cell.Occupied() {
				drawColor = color.RGBA{255, 255, 0, 255}
			}

			vector.StrokeLine(screen, cx, cy, cx+cw, cy, 1, drawColor, false)

			vector.StrokeLine(screen, cx+cw, cy, cx+cw, cy+ch, 1, drawColor, false)

			vector.StrokeLine(screen, cx+cw, cy+ch, cx, cy+ch, 1, drawColor, false)

			vector.StrokeLine(screen, cx, cy+ch, cx, cy, 1, drawColor, false)
		}

	}

}
