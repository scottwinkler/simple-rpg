package player

import (
	"math/rand"
	"strconv"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/rs/xid"
	"github.com/scottwinkler/pixel-experiment/animation"
	"github.com/scottwinkler/pixel-experiment/sound"
	"github.com/scottwinkler/pixel-experiment/world"
)

type Player struct {
	id               string
	sprite           *pixel.Sprite
	speed            float64
	v                pixel.Vec
	r                float64 //used for collider calculations
	matrix           pixel.Matrix
	animationManager *animation.AnimationManager
	soundManager     *sound.SoundManager
	world            *world.World
	direction        int
}

func NewPlayer(v pixel.Vec, r float64, animations []*animation.Animation, sounds []*sound.Sound, w *world.World) *Player {
	animationManager := animation.NewAnimationManager(animations)
	soundManager := sound.NewSoundManager(sounds)
	animationManager.Select("Idle")
	sprite := animationManager.Selected.Next()
	matrix := pixel.Matrix(pixel.IM.Moved(v))
	//radius := float64(world.Tilemap.TileHeight / 3) //not a great solution for rectangles
	//fmt.Printf("r: %f", radius)
	id := xid.New().String()
	player := &Player{
		id:               id,
		sprite:           sprite,
		speed:            3, //default
		v:                v,
		r:                r,
		matrix:           matrix,
		animationManager: animationManager,
		soundManager:     soundManager,
		world:            w,
		direction:        world.DOWN,
	}
	return player
}

func (p *Player) Id() string {
	return p.id
}

func (p *Player) V() pixel.Vec {
	return p.v
}

func (p *Player) R() float64 {
	return p.r
}

func (p *Player) Speed() float64 {
	return p.speed
}

func (p *Player) Direction() int {
	return p.direction
}

func (p *Player) Material() string {
	return world.MATERIAL_FLESH
}

func (p *Player) Move(direction int) {
	nextPos := pixel.V(p.v.X, p.v.Y)
	p.direction = direction
	switch direction {
	case world.LEFT:
		nextPos.X -= p.speed
		p.animationManager.Select("MoveLeft")
	case world.RIGHT:
		nextPos.X += p.speed
		p.animationManager.Select("MoveRight")
	case world.DOWN:
		nextPos.Y -= p.speed
		p.animationManager.Select("MoveDown")
	case world.UP:
		nextPos.Y += p.speed
		p.animationManager.Select("MoveUp")
	}
	if !p.world.Collides(p.Id(), nextPos, p.r) {
		p.v = nextPos
	}
	//update matrix and collision circle
	matrix := pixel.IM.Moved(p.v)
	p.matrix = matrix
}

func (p *Player) HandleHit(s world.GameObject, cb world.Fn_Callback) bool {
	//am i near enoguh to be affected?
	//draw a slightly bigger circle than the collision circle
	//so that the hit box is reasonable

	if world.CircleCollision(p.v, p.r, s.V(), s.R()+s.Speed()) {
		//where am i relative to the source?
		relativePos := p.v.Sub(s.V())
		top := relativePos.Y >= relativePos.X    //above line y=x?
		right := relativePos.Y >= -relativePos.X //above line y=-x?

		var relativeDir int
		if top && right {
			relativeDir = world.UP
		} else if top {
			relativeDir = world.LEFT
		} else if right {
			relativeDir = world.RIGHT
		} else {
			relativeDir = world.DOWN
		}
		//is the source facing the right direction?
		if relativeDir == s.Direction() {
			switch p.direction {
			case world.LEFT:
				p.animationManager.Select("HitLeft")
				cb(p)
				return true
			case world.RIGHT:
				p.animationManager.Select("HitRight")
				cb(p)
				return true
			case world.DOWN:
				p.animationManager.Select("HitDown")
				cb(p)
				return true
			case world.UP:
				p.animationManager.Select("HitUp")
				cb(p)
				return true
			}
		}
	}
	return false
}

func (p *Player) Attack(direction int) {
	p.direction = direction
	//p.SoundManager.Play("attack1")
	switch direction {
	case world.LEFT:
		p.animationManager.Select("AttackLeft")
	case world.RIGHT:
		p.animationManager.Select("AttackRight")
	case world.DOWN:
		p.animationManager.Select("AttackDown")
	case world.UP:
		p.animationManager.Select("AttackUp")
	}

	//create callback for playing the appropriate sound effect
	cb := func(obj interface{}) {
		num := strconv.Itoa(rand.Intn(2) + 1) //for playing a random attack sound
		if obj != nil {
			gameObject := obj.(world.GameObject)
			material := gameObject.Material()

			switch material {
			case world.MATERIAL_FLESH:
				p.soundManager.Play("humanattacking"+num, "swordhitflesh")
			case world.MATERIAL_METAL:
				p.soundManager.Play("humanattacking"+num, "swordhitmetal")
			case world.MATERIAL_WOOD:
				p.soundManager.Play("swordhitwood")
			}
		} else {
			//	fmt.Println("playing swordswing")
			p.soundManager.Play("humanattacking"+num, "swordswing")
		}
	}
	p.world.HitEvent(p, cb)
}

func (p *Player) Update(tick int) {
	var (
		camPos = p.v //position camera centered on player
		//camSpeed = 1.0
		camZoom = 1.0
	)
	//fmt.Println("-----------------")
	win := p.world.Window
	cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
	win.SetMatrix(cam)

	//fmt.Println("deciding to select one...")
	if p.animationManager.Selected.Skippable() || p.animationManager.Selected.Done() { //only listen to new events if the current animation is skippable or done playing
		//fmt.Printf("current animation done: %t", p.AnimationManager.Selected.Done())
		if win.Pressed(pixelgl.KeyLeft) || win.Pressed(pixelgl.KeyA) {
			p.Move(world.LEFT)
		} else if win.Pressed(pixelgl.KeyRight) || win.Pressed(pixelgl.KeyD) {
			p.Move(world.RIGHT)
		} else if win.Pressed(pixelgl.KeyDown) || win.Pressed(pixelgl.KeyS) {
			p.Move(world.DOWN)
		} else if win.Pressed(pixelgl.KeyUp) || win.Pressed(pixelgl.KeyW) {
			p.Move(world.UP)
		} else if win.Pressed(pixelgl.MouseButtonLeft) {
			//determine what quadrant relative to the player the mouse click happened
			mouse := cam.Unproject(win.MousePosition())
			dir := world.RelativeDirection(mouse, p.v)
			p.Attack(dir)

		} else {
			p.animationManager.Select("Idle")
		}

		//fmt.Println("drawing sprite")
		//fmt.Println("-----------------")
	}
	if tick == 0 {
		p.sprite = p.animationManager.Selected.Next()
	}
	p.Draw()
}
func (p *Player) Draw() {
	animation := p.animationManager.Selected
	//chained methods so that we first scale by spritesheet size, then by reflection, then by position
	matrix := animation.Matrix().Chained(p.matrix)
	p.sprite.Draw(p.world.Window, matrix)
}
