package main

import (
	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/scottwinkler/pixel-experiment/animation"
	"github.com/scottwinkler/pixel-experiment/entity"
	"github.com/scottwinkler/pixel-experiment/player"
	"github.com/scottwinkler/pixel-experiment/sound"
	"github.com/scottwinkler/pixel-experiment/tilemap"
	"github.com/scottwinkler/pixel-experiment/utility"
	"github.com/scottwinkler/pixel-experiment/world"
)

func run() {
	player_sound_mapping := utility.LoadJSON("./_configuration/player/sounds.json")
	player_sounds := sound.MappingToSounds(player_sound_mapping)

	tm, _ := tilemap.ParseTiledJSON("_assets/tmx/world1.json")
	animation_mapping := utility.LoadJSON("./_configuration/player/animations.json")

	player_spritesheet := utility.LoadSpritesheet("_assets/spritesheets/chara_hero.png", pixel.R(0, 0, 48, 48), 2.0)
	player_animations := animation.MappingToAnimations(player_spritesheet, animation_mapping)
	slime_spritesheet := utility.LoadSpritesheet("_assets/spritesheets/chara_slime.png", pixel.R(0, 0, 48, 48), 2.0)
	slime_animations := animation.MappingToAnimations(slime_spritesheet, animation_mapping)

	world := world.NewWorld(pixel.R(0, 0, 400, 400), tm)
	slime := entity.NewEntity(pixel.V(150, 300), 16, slime_animations, world)
	player := player.NewPlayer(pixel.V(150, 200), 16, player_animations, player_sounds, world)

	world.AddGameObject("player", player)
	world.AddGameObject("monster", slime)
	world.Start(60.0, 15.0)

}

func main() {
	pixelgl.Run(run)
}
