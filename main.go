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
	"sync"
	"os"
	"golang.org/x/crypto/acme/autocert"
)

//Variables
var currentPlayers=0
var ships = make([]cosmicStruct.PlayerShip,0)
var world = box2d.MakeB2World(box2d.MakeB2Vec2(0,0))
var worldLock = &sync.Mutex{}
var lobby = true
var time = settings.LOBBY_TIME
var dust = make([]cosmicStruct.Dust,0)
var clientDust = make([]cosmicStruct.ClientDust,0)
//Channels
var dustPopChannel = make([]chan int,0)

func main() {
	game()
}

func game() {
	world.SetContactListener(CollisionListener{})
	log.Println("Loading game server")

	//Server config
	sockets,err := socketio.NewServer(nil)
	//sockets.SetPingInterval(time2.Millisecond*250)
	//sockets.SetPingTimeout(time2.Second*1)
	sockets.SetAllowUpgrades(true)
	if err != nil {
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

		//Creating player
		worldLock.Lock()
		playerShipTmp := cosmicStruct.PlayerShip{
			Id:        currentPlayers,
			Transform: world.CreateBody(&bodydef),
			Heading:   0,
			Username:  "",
			Health:    settings.STARTING_HP,
			SockId:    sock.Id(),
			Alive:     true,
			DustPop:   make(chan int),
			SyncDust:  make(chan bool),
		}
		playerShipTmp.Transform.CreateFixture(collider,1.0).SetRestitution(0.0) //Setting up collider
		worldLock.Unlock()

		//Getting player references
		ships = append(ships, playerShipTmp)

		//Events
		sock.On("movement",func(data cosmicStruct.Movement){
			ship,err := cosmicStruct.FindShipBySocketId(&ships,sock.Id())
			if err == nil {
				ships[*ship].Movement = data
			}
		})

		sock.On("username",func(data string){
			ship,err := cosmicStruct.FindShipBySocketId(&ships,sock.Id())
			if err == nil {
				log.Println(fmt.Sprintf("Player %s changed username to %s", ships[*ship].Username, data))
				ships[*ship].Username = data
			}
		})

		sock.On("skin",func(data int){
			ship,err := cosmicStruct.FindShipBySocketId(&ships,sock.Id())
			if err == nil {
				ships[*ship].SkinId = data
			}
		})

		//Sync functions
		syncUI := func(){
			var title string //Page title
			if lobby {
				title = "Cosmic - Lobby"
			} else {
				title = "Cosmic"
			}

			var alert cosmicStruct.Alert //Alert in the game

			sock.Emit("ui", cosmicStruct.UIData{
				Title: title,
				Lobby: lobby,
				Time:  math.Floor(time),
				Alert: alert,
			})
		}


		syncShips := func(){
			sock.Emit("ships",cosmicStruct.ConvertToClientShips(&ships))
		}

		syncDust := func(){
			ship,err := cosmicStruct.FindShipBySocketId(&ships,sock.Id())
			if err == nil {
				for {
					select {
					case <-ships[*ship].SyncDust:
						sock.Emit("cosmicDust",clientDust)
						log.Println("Full dust sync performed")
					}
				}
			}
		}

		dustPop := func(){
			ship,err := cosmicStruct.FindShipBySocketId(&ships,sock.Id())
			if err == nil {
				sock.Emit("cosmicDust",clientDust) //Full sync at go-routine creation time
				for {
					select {
					case dustToPop := <-ships[*ship].DustPop:
						sock.Emit("dustRemove", dustToPop)
						log.Println("Dust removed:",dustToPop)
					}
				}
			}
		}

		//Sync timers
		jsexec.SetInterval(func(){syncUI()},settings.SYNC_UI,true)
		jsexec.SetInterval(func(){syncShips()},settings.SYNC_SHIPS ,true)

		//Async go-routines
		go syncDust()
		go dustPop()

		//Full game state sync
		syncUI()
		syncShips()

		//Disconnect
		sock.On("disconnection",func(sock socketio.Socket) {
			log.Println("Player disconnected:"+sock.Id())
			//Cleanup array
			i, err := cosmicStruct.FindShipBySocketId(&ships,sock.Id())
			if err!=nil{
				panic(err) //Ship not found - something must went really wrong
			}
			ships[*i] = ships[len(ships)-1]
			ships = ships[:len(ships)-1]
		})
	})

	//Server loop
	jsexec.SetInterval(func(){update(float64(settings.SERVER_BEAT)/1000)},settings.SERVER_BEAT,false)

	http.Handle("/socket.io/",sockets)
	http.Handle("/",http.FileServer(http.Dir("./local")))
	if len(os.Args)>1 {
		log.Println("Server ready [SSL MODE] Domain:",os.Args[1])
		log.Fatal(http.Serve(autocert.NewListener(os.Args[1]),nil))
	} else {
		log.Println("Server ready")
		log.Fatal(http.ListenAndServe(":3000", nil))
	}
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
	worldLock.Lock()
	world.Step(deltaTime,10,10) //Physics update
	worldLock.Unlock()
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
		//Add to world
		worldLock.Lock()
		dust = append(dust,cosmicStruct.Dust{
			Transform: world.CreateBody(&bodydef),
		})
		worldLock.Unlock()
		fixture := dust[i].Transform.CreateFixture(shape,0.0)
		fixture.SetSensor(true)
	}
	fullsyncClientDust()
}

func updateClientDust(){
	clientDust = cosmicStruct.GenerateClientDust(&dust)
}

func fullsyncClientDust(){
	updateClientDust()
	for i,_ :=range ships {
		ships[i].SyncDust <- true
	}
}
func popClientDust(dustId int){
	updateClientDust()
	for i,_ :=range ships {
		ships[i].DustPop <- dustId
	}
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
		if i != nil { //Check if it isn't null pointer to player ship index
			dust = append(dust[:*i], dust[*i+1:]...) //Remove dust from array by index
			ships[*res1].Score++ //Add score to ship
			popClientDust(*i) //Async remove dust from client
		}
	} else {
		res2 := cosmicStruct.FindShipByTransform(&ships,bodyB)
		i := cosmicStruct.FindDustByTransform(&dust,bodyA)
		if i != nil {
			dust[*i] = dust[len(dust)-1]
			dust = dust[:len(dust)-1]
			ships[*res2].Score++
		}
	}

}
func (CollisionListener) EndContact(contact box2d.B2ContactInterface){

}

func (CollisionListener) PreSolve(contact box2d.B2ContactInterface,oldManifold box2d.B2Manifold){

}
func (CollisionListener) PostSolve(contact box2d.B2ContactInterface,impulse *box2d.B2ContactImpulse){

}
