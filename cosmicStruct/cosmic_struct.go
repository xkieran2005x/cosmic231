package cosmicStruct

import (
	"github.com/ByteArena/box2d"
	"errors"
	"cosmicio/settings"
)

type PlayerShip struct
{
	Id int
	Transform *box2d.B2Body
	Health int
	Username string
	Score int
	SockId string
	Alive bool
	Cooldown float64
	SkinId int
	Movement Movement
	DustPop chan int
	SyncDust chan bool
	AddParticle chan ClientParticle
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

type Particle struct {
	Transform *box2d.B2Body
	Size int
	Type int
	Lifetime float64
	Owner *PlayerShip
}

type ClientParticle struct {
	X float64
	Y float64
	VX float64 //X Velocity
	VY float64 //Y Velocity
	Size int
	Type int
	Lifetime float64
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

///Returns an ship index in array by transform
func FindShipByTransform(ships *[]PlayerShip,transform *box2d.B2Body) *int {
	for r,ship :=range *ships{
		if ship.Transform == transform{
			return &r
		}
	}
	return nil
}

///Returns an dust index in array by transform
func FindDustByTransform(dusts *[]Dust,transform *box2d.B2Body) *int {
	for i,dust :=range *dusts{
		if dust.Transform == transform{
			return &i
		}
	}
	return nil
}

func (ship *PlayerShip) CleanTurn() {
	ship.Score = 0 //Reset score
	ship.Transform.SetTransform(box2d.MakeB2Vec2(2,4),0) //Set position to 2,4
	ship.Health = settings.STARTING_HP //Reset HP
}

func (particle *Particle) ToClientParticle() ClientParticle {
	return ClientParticle{
		X: particle.Transform.GetPosition().X,
		Y: particle.Transform.GetPosition().Y,
		VX: particle.Transform.M_linearVelocity.X,
		VY: particle.Transform.M_linearVelocity.Y,
		Size: particle.Size,
		Lifetime: particle.Lifetime,
		Type: particle.Type,
	}
}