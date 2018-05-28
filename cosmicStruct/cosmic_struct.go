package cosmicStruct

import (
	"github.com/ByteArena/box2d"
)

type PlayerShip struct
{
	Id int
	Transform *box2d.B2Body
	Heading float64
	Health int
	Username string
	Score int
	SockId string
	Alive bool
	Cooldown float64
	SkinId int
	Movement Movement

}

type ClientShip struct
{
	X float64
	Y float64
	Heading float64
	Health int
	Username string
	Score int
	SockId string
	SkinId int
}

type UIData struct {
	Title string
	Lobby bool
	Time float64
	Alert Alert
}

type Alert struct {
	Message string
	Duration float64
}

type Movement struct {
	left bool
	right bool
	up bool
	down bool
	shoot bool
}

func ConvertToClientShip(ship *PlayerShip) ClientShip{
	return ClientShip{
		X: ship.Transform.GetPosition().X,
		Y: ship.Transform.GetPosition().Y,
		Heading: ship.Transform.GetAngle(),
		Health: ship.Health,
		Username: ship.Username,
		Score: ship.Score,
		SockId: ship.SockId,
		SkinId: ship.SkinId,
	}
}

func ConvertToClientShips(ships *[]PlayerShip) []ClientShip{
	clientShips := make([]ClientShip,0)
	for _,ship :=range *ships{
		clientShips = append(clientShips,ClientShip{
			X: ship.Transform.GetPosition().X,
			Y: ship.Transform.GetPosition().Y,
			Heading: ship.Transform.GetAngle(),
			Health: ship.Health,
			Username: ship.Username,
			Score: ship.Score,
			SockId: ship.SockId,
			SkinId: ship.SkinId,
		})
	}
	return clientShips
}