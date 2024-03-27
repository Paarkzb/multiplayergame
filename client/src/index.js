class Player {
    constructor(id, name, pos, angle) {
        this.id = id;
        this.name = name;
        this.pos = pos;
        this.width = 50;
        this.height = 50;
        this.angle = angle;
    }

    draw() {
        ctx.save();

        ctx.beginPath();
        ctx.fillStyle = "red";
        console.log("DRAW", this.pos.x, this.pos.y, this.width, this.height, this.angle);
        ctx.translate(this.pos.x + this.width/2, this.pos.y + this.height/2 );
        ctx.rotate((this.angle * Math.PI) / 180);
        ctx.fillRect(-this.width/2, -this.height/2, this.width, this.height);
        ctx.closePath();

        // Reset current transformation matrix to the identity matrix
        // ctx.setTransform(1, 0, 0, 1, 0, 0);

        ctx.restore();
    }
}

let canvas = document.getElementById("game-canvas");
let ctx = canvas.getContext("2d");

class Game {
    constructor(canvas) {
        this.canvas = canvas;
        this.ctx = this.canvas.getContext("2d");
    }

    start() {
        let form = document.getElementById("form");
        form.remove();
    }

    update(evt) {
        ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        const {player, otherPlayers} = this.getCurrentState(evt);

        console.log("Update", player);
        console.log("Others", otherPlayers);

        player.draw();
        otherPlayers.forEach((otherPlayer) => {otherPlayer.draw();
        })
    }

    getCurrentState(evt) {
        console.log("getCurrentState", evt);
        let player = new Player(evt.player.id, evt.player.name, evt.player.position, evt.player.angle);
        let otherPlayers = [];
        evt.otherPlayers?.forEach((otherPlayer) => {
            let oPlayer = new Player(otherPlayer.id, otherPlayer.name, otherPlayer.position, otherPlayer.angle);
            otherPlayers.push(oPlayer);
        })
        console.log("Current State", player);
        return {
            player: player,
            otherPlayers: otherPlayers,
        }

    }
}

let game;


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
            

            game = new Game(canvas);
            game.start();

            document.addEventListener("keydown", function (event) {
                if (event.code == "KeyA") {
                    sendEvent("move", "left");
                }
                else if (event.code == "KeyD") {
                    sendEvent("move", "right");
                }
                else if (event.code == "KeyW") {
                    sendEvent("move", "forward");
                }
                else if (event.code == "KeyS") {
                    sendEvent("move", "back");
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
            game.update(event);
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
    let msg;
    switch (eventName) {
        case "login":
            msg = payload;
            break;
        case "move":
            msg = payload;
            break;
    }

    // Create a event Object with a event named send_message
    const event = new EventMsg(eventName, msg);
    conn.send(JSON.stringify(event));
}

