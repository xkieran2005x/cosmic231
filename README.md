# Cosmic-Server
Open-source Cosmic game server, written in Go  
[Link for game client, written in JavaScript](https://github.com/yknomeh/CosmicIO---Client)
## Compilation
`go build cosmicio` Note: Package should be cloned into `GOPATH%\src\cosmicio` directory
## Usage
`./cosmicio` - Start a basic server with filserver from `local` directory
`./cosmicio <2 SSL Key files>` - Start a SSL server with filserver from `local` directory
## Security
Physics is fully server-side, only keyboard input is send from the client, so it's impossible to cheat

