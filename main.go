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
var socks = make([]*socketio.Socket,0)
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
	log.Println("Loading game server")
	sockets,err := socketio.NewServer(nil)
	if err != nil{
		log.Fatal(err)
	}

	//Connecting
	sockets.On("connection",func(sock socketio.Socket) {
		log.Println("Player connected:" + sock.Id())
		socks = append(socks)
		currentPlayers++
		bodydef := box2d.MakeB2BodyDef()
		bodydef.Type = 0
		bodydef.Position.Set(2, 4)
		bodydef.Angle = 0

		playerShip := cosmicStruct.PlayerShip{
			Id:        currentPlayers,
			Transform: world.CreateBody(&bodydef),
			Heading:   0,
			Username:  "",
			Health:    settings.STARTING_HP,
			SockId:    sock.Id(),
			Alive:     true,
		}
		ships = append(ships, playerShip)
		socks = append(socks, &sock)

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
		jsexec.SetInterval(func(){syncShips()},settings.SYNC_UI,true)
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
	if !lobby{
		updatePosition(deltaTime)
	}
	updateTime(deltaTime)
}

func updateTime(deltaTime float64){
	time -= deltaTime
	if time < 0{
		if lobby{
			time =settings.GAME_TIME
			lobby = !lobby
			generateDust()
			log.Println("Game started")
		} else {
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
	point := box2d.MakeB2Vec2(0, 0)
	for _,value := range ships {
		force := box2d.MakeB2Vec2(0, settings.PHYSICS_FORCE*deltaTime)
		nforce := box2d.MakeB2Vec2(0, settings.PHYSICS_FORCE*deltaTime*-1)
		if value.Movement.Up {value.Transform.ApplyForce(force,point, true)}
		if value.Movement.Down {value.Transform.ApplyForce(nforce,point, true)}
		if value.Movement.Left {value.Transform.SetAngularVelocity(settings.PHYSICS_ROTATION_FORCE*-1)}
		if value.Movement.Left {value.Transform.SetAngularVelocity(settings.PHYSICS_ROTATION_FORCE)}
	}
}

func generateDust(){
	for i := 0; i < settings.AMOUNT_OF_DUST; i++ {
		//Position generation
		x := rand.Float64() * (500 * settings.MAP_SIZE - -500 * settings.MAP_SIZE) + -500 * settings.MAP_SIZE
		y := rand.Float64() * (500 * settings.MAP_SIZE - -500 * settings.MAP_SIZE) + -500 * settings.MAP_SIZE
		bodydef := box2d.MakeB2BodyDef()
		bodydef.Type = 0
		bodydef.Position.Set(x, y)
		bodydef.Angle = 0
		dust = append(dust,cosmicStruct.Dust{
			Transform: world.CreateBody(&bodydef),
		})
	}
	clientDust = cosmicStruct.GenerateClientDust(&dust)
}

