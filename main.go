package main

import (
	"github.com/googollee/go-socket.io"
	"log"
	"cosmicio/cosmicStruct"
	"github.com/ByteArena/box2d"
	"cosmicio/settings"
	"net/http"
	"math"
	"cosmicio/jsexec"
	"math/rand"
	"fmt"
)

//Variables
var currentPlayers=0;
var ships = make([]cosmicStruct.PlayerShip,0)
var world = box2d.MakeB2World(box2d.MakeB2Vec2(0,0))
var lobby = true
var time = settings.LOBBY_TIME
var dust = make([]cosmicStruct.Dust,0)
var clientDust = make([]cosmicStruct.ClientDust,0)

func main() {
	game()
}

func game() {
	world.SetContactListener(CollisionListener{})
	log.Println("Loading game server")
	sockets,err := socketio.NewServer(nil)
	if err != nil{
		log.Fatal(err)
	}

	//Connecting
	sockets.On("connection",func(sock socketio.Socket) {
		log.Println("Player connected:" + sock.Id())
		currentPlayers++

		//Creating body definition
		bodydef := box2d.MakeB2BodyDef()
		bodydef.Type = 2
		bodydef.Position.Set(2, 4)
		bodydef.Angle = 0
		bodydef.AngularDamping = settings.PHYSICS_ANGULAR_DUMPING
		bodydef.LinearDamping = settings.PHYSICS_LINEAR_DUMPING

		//Creating collider
		collider := box2d.NewB2PolygonShape()
		collider.SetAsBox(40,120)

		//Creating player ship
		playerShipTmp := cosmicStruct.PlayerShip{
			Id:        currentPlayers,
			Transform: world.CreateBody(&bodydef),
			Heading:   0,
			Username:  "",
			Health:    settings.STARTING_HP,
			SockId:    sock.Id(),
			Alive:     true,
		}
		playerShipTmp.Transform.CreateFixture(collider,1.0)

		//Getting player references
		ships = append(ships, playerShipTmp)
		playerShipInt,err := cosmicStruct.FindShipBySocketId(&ships,sock.Id())
		if err != nil {panic(err)}
		playerShip := &ships[*playerShipInt]

		//Events
		sock.On("movement",func(data cosmicStruct.Movement){
			playerShip.Movement = data
		})

		sock.On("username",func(data string){
			log.Println(fmt.Sprintf("Player %s changed username to %s",playerShip.Username,data))
			playerShip.Username = data
		})

		sock.On("skin",func(data int){
			playerShip.SkinId = data
		})

		//Sync functions
		syncUI := func(){
			sock.Emit("ui", cosmicStruct.UIData{
				Title: "Cosmic",
				Lobby: lobby,
				Time:  math.Floor(time),
			})
		}

		syncShips := func(){
			sock.Emit("ships",cosmicStruct.ConvertToClientShips(&ships))
		}

		syncDust := func(){
			sock.Emit("cosmicDust",clientDust)
		}

		//Sync timers
		jsexec.SetInterval(func(){syncUI()},settings.SYNC_UI,true)
		jsexec.SetInterval(func(){syncShips()},settings.SYNC_SHIPS ,true)
		jsexec.SetInterval(func(){syncDust()},settings.SYNC_DUST,true)
	})

	//Disconnecting
	sockets.On("disconnection",func(sock socketio.Socket) {
		log.Println("Player disconnected:"+sock.Id())
		//Cleanup array
		i, err := cosmicStruct.FindShipBySocketId(&ships,sock.Id())
		if err!=nil{
			panic(err) //Ship not found - something must went really wrong
		}
		ships[*i] = ships[len(ships)-1]
		ships = ships[:len(ships)-1]
	})

	//Server loop
	jsexec.SetInterval(func(){update(float64(settings.SERVER_BEAT)/1000)},settings.SERVER_BEAT,false)

	http.Handle("/socket.io/",sockets)
	http.Handle("/",http.FileServer(http.Dir("./local")))
	log.Println("Server ready")
	log.Fatal(http.ListenAndServe(":3000",nil))
}

func update(deltaTime float64) {
	if !lobby{ //Game-only logic
		updatePosition(deltaTime)
	}
	updateTime(deltaTime)
}

func updateTime(deltaTime float64){
	time -= deltaTime
	if time < 0{
		if lobby{
			//Pre-game operations
			time =settings.GAME_TIME
			lobby = !lobby
			generateDust()
			log.Println("Game started")
		} else {
			//Post-game cleanup
			time = settings.LOBBY_TIME
			lobby = !lobby
			//Dust cleanup
			for _,dust := range dust{
				world.DestroyBody(dust.Transform)
			}
			dust= dust[:0]
			log.Println("Game ended")
		}
	}
}

func updatePosition(deltaTime float64){
	for _,value := range ships {

		forceDirection := value.Transform.GetWorldVector(box2d.MakeB2Vec2(1,0)) //Forward vector
		force := box2d.B2Vec2CrossScalarVector(settings.PHYSICS_FORCE,forceDirection) //Forward force

		//Input movement handling
		if value.Movement.Up {value.Transform.SetLinearVelocity(force)}
		//if value.Movement.Down {value.Transform.ApplyForce(nforce,point, true)}
		if value.Movement.Left {value.Transform.SetAngularVelocity(settings.PHYSICS_ROTATION_FORCE*-1)}
		if value.Movement.Right {value.Transform.SetAngularVelocity(settings.PHYSICS_ROTATION_FORCE)}
	}
	world.Step(deltaTime,8,3) //Physics update
}

func generateDust(){
	for i := 0; i < settings.AMOUNT_OF_DUST; i++ {
		//Position generation
		x := rand.Float64() * (500 * settings.MAP_SIZE - -500 * settings.MAP_SIZE) + -500 * settings.MAP_SIZE
		y := rand.Float64() * (500 * settings.MAP_SIZE - -500 * settings.MAP_SIZE) + -500 * settings.MAP_SIZE
		//Dust physics body definition
		bodydef := box2d.MakeB2BodyDef()
		bodydef.Type = 0
		bodydef.Position.Set(x, y)
		bodydef.Angle = 0
		//Dust collider
		shape := box2d.MakeB2CircleShape()
		shape.SetRadius(5)
		dust = append(dust,cosmicStruct.Dust{
			Transform: world.CreateBody(&bodydef),
		})
		fixture := dust[i].Transform.CreateFixture(shape,0.0)
		fixture.SetSensor(true)
	}
	updateClientDust()
}

func updateClientDust(){
	clientDust = cosmicStruct.GenerateClientDust(&dust)
}

//Contact listener
type CollisionListener struct{}

func (CollisionListener) BeginContact(contact box2d.B2ContactInterface){
	//Get colliding bodies
	bodyA := contact.GetFixtureA().GetBody() //Dynamic body
	bodyB := contact.GetFixtureB().GetBody() //Static body

	res1 := cosmicStruct.FindShipByTransform(&ships,bodyA) //Get ship reference
	if res1 != nil { //Check if it isn't null pointer
		i := cosmicStruct.FindDustByTransform(&dust,bodyB) //Find dust index reference
		if i != nil {
			//Remove dust from array by index
			dust[*i] = dust[len(dust)-1]
			dust = dust[:len(dust)-1]

			ships[*res1].Score++
			updateClientDust()
		}
	} else {
		res2 := cosmicStruct.FindShipByTransform(&ships,bodyB)
		i := cosmicStruct.FindDustByTransform(&dust,bodyA)
		dust[*i] = dust[len(dust)-1]
		dust = dust[:len(dust)-1]
		ships[*res2].Score++
	}

}
func (CollisionListener) EndContact(contact box2d.B2ContactInterface){

}

func (CollisionListener) PreSolve(contact box2d.B2ContactInterface,oldManifold box2d.B2Manifold){

}
func (CollisionListener) PostSolve(contact box2d.B2ContactInterface,impulse *box2d.B2ContactImpulse){

}
