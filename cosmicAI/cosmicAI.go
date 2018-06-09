package cosmicAI

import (
"cosmicio/cosmicStruct"
"github.com/ByteArena/box2d"
	"math"
)

//Variables
var ships *[]cosmicStruct.PlayerShip
var dusts *[]cosmicStruct.Dust

func Load(shipsRef *[]cosmicStruct.PlayerShip,dustRef *[]cosmicStruct.Dust){
	ships=shipsRef
	dusts=dustRef
}

func pickRandomDustLocation() box2d.B2Vec2 {
	max := len(*dusts)
	prePos := *dusts
	return prePos[randomRange(0,max)].Transform.GetPosition()
}

func getAngleToVec(this,vec box2d.B2Vec2) float64 {
	return math.Atan2(vec.Y,vec.X) - math.Atan2(this.Y,this.X)
}

