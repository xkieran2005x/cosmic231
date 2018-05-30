package cosmicStruct

import (
	"github.com/ByteArena/box2d"
	"errors"
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
	Left bool
	Right bool
	Up bool
	Down bool
	Shoot bool
}

type Dust struct {
	Transform *box2d.B2Body
}

type ClientDust struct {
	X float64
	Y float64
	Size int
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

func FindShipBySocketId(ships *[]PlayerShip, socketId string) (*int,error) {
	for iteration, ship := range *ships {
		if ship.SockId==socketId{
			return &iteration,nil
		}
	}
	return nil,errors.New("failed to find ship")
}

func GenerateClientDust(dust *[]Dust) []ClientDust {
	clientDust := make([]ClientDust,0,len(*dust))
	for _,dust := range *dust {
		clientDust = append(clientDust,ClientDust{
			X: dust.Transform.GetPosition().X,
			Y: dust.Transform.GetPosition().Y,
			Size: 5,
		})
	}
	return clientDust
}