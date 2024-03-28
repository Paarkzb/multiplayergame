class Player {
    constructor(id, name, pos, angle, ctx) {
        this.id = id;
        this.name = name;
        this.pos = pos;
        this.width = 50;
        this.height = 50;
        this.angle = angle;
        this.ctx = ctx;
    }

    draw() {
        this.ctx.save();

        this.ctx.beginPath();
        this.ctx.fillStyle = "red";
        console.log("DRAW", this.pos.x, this.pos.y, this.width, this.height, this.angle);
        this.ctx.translate(this.pos.x + this.width/2, this.pos.y + this.height/2 );
        this.ctx.rotate((this.angle * Math.PI) / 180);
        this.ctx.fillRect(-this.width/2, -this.height/2, this.width, this.height);
        this.ctx.closePath();

        // Reset current transformation matrix to the identity matrix
        // this.ctx.setTransform(1, 0, 0, 1, 0, 0);

        this.ctx.restore();
    }
}

class Game {
    
    constructor() {
        this.canvas = document.getElementById("game-canvas");;
        this.ctx = this.canvas.getContext("2d");
        this.firstUpdate = false;
        this.state = "";
    }

    start() {
        let form = document.getElementById("form");
        form.remove();
    }

    update() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        const {player, otherPlayers} = this.getCurrentState();

        if(player) {
            console.log("Update", player);
            console.log("Others", otherPlayers);

            player.draw();
            otherPlayers.forEach((otherPlayer) => {otherPlayer.draw();});
        }
    }

    gameLoop() {
        this.update();

        requestAnimationFrame(this.gameLoop());
    }

    getCurrentState() {
        let player = new Player(this.state.player.id, this.state.player.name, this.state.player.position, this.state.player.angle, this.ctx);
        let otherPlayers = [];
        this.state.otherPlayers?.forEach((otherPlayer) => {
            let oPlayer = new Player(otherPlayer.id, otherPlayer.name, otherPlayer.position, otherPlayer.angle, this.ctx);
            otherPlayers.push(oPlayer);
        })
        console.log("Current State", player);
        return {
            player: player,
            otherPlayers: otherPlayers,
        }

    }

    setCurrentState(evt){
        console.log("setCurrentState", evt);
        this.state = evt;
    }
}

let game = new Game();

/**
 * login will send a login request to the server and then connect websocket
 */
function login() {
    connectWebsocket();

    return false;
}

/**
 * connectWebsocket will connect to websocket and add listeners
 */
function connectWebsocket() {
    if (window["WebSocket"]) {
        console.log("supports websockets");
        conn = new WebSocket("ws://localhost:8080/ws");

        // Onopen
        conn.onopen = function (evt) {
            let username = document.getElementById("username").value;
            sendEvent("login", username);
            
            game.start();

            document.addEventListener("keydown", function (event) {
                if (event.code == "KeyA") {
                    sendEvent("keydown", "left");
                }
                if (event.code == "KeyD") {
                    sendEvent("keydown", "right");
                }
                if (event.code == "KeyW") {
                    sendEvent("keydown", "forward");
                }
                if (event.code == "KeyS") {
                    sendEvent("keydown", "back");
                }
            });
            document.addEventListener("keyup", function (event) {
                if (event.code == "KeyA") {
                    sendEvent("keyup", "left");
                }
                if (event.code == "KeyD") {
                    sendEvent("keyup", "right");
                }
                if (event.code == "KeyW") {
                    sendEvent("keyup", "forward");
                }
                if (event.code == "KeyS") {
                    sendEvent("keyup", "back");
                }
            });

        };

        conn.onmessage = function (event) {
            console.log(event);

            // parse websocket message as JSON
            const eventData = JSON.parse(event.data);
            console.log(eventData);
            routeEvent(eventData);

        };
    } else {
        alert("Not supporting websockets");
    }
}

/**
 * routeEvent is a proxy function that routes
 * events into their correct Handler
 * based on the type field
 * */
function routeEvent(event) {
    if (event.type === undefined) {
        alert("no 'type' field in event");
    }
    console.log("EVENT", event.type);
    switch (event.type) {
        case "update":
            game.setCurrentState(event);

            if(!game.firstUpdate) {
                game.update();
            }
            // game.update();
            break;
        default:
            alert("unsupported message type");
            break;
    }
}

/**
 * Event is used to wrap all messages Send and Receive
 * on the Websocket
 * The type is used as a RPC
 **/
class EventMsg {
    // Each event needs a type
    // The playload is not required
    constructor(type, payload) {
        this.type = type;
        this.payload = payload;
    }
}

/**
 * sendEvent
 * eventName - the event name to send on
 * payload - the data payload
 * */
function sendEvent(eventName, payload) {
    // let msg;
    // switch (eventName) {
    //     case "login":
    //         msg = payload;
    //         break;
    //     case "move":
    //         msg = payload;
    //         break;
    // }

    // Create a event Object with a event named send_message
    const event = new EventMsg(eventName, payload);
    console.log("SENDEVENT", event);
    conn.send(JSON.stringify(event));
}

