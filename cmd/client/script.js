class Player {
    constructor(id, name, pos) {
        this.id = id;
        this.name = name;
        this.pos = pos;
    }

    draw() {
        ctx.beginPath();
        ctx.fillStyle = "red";
        ctx.fillRect(this.pos.x, this.pos.y, 50, 50);
        ctx.closePath();
    }
}

let canvas = document.getElementById("game-canvas");
let ctx = canvas.getContext("2d");

class Game {
    constructor(player, canvas) {
        this.player = player;
        this.canvas = canvas;
        this.ctx = this.canvas.getContext("2d");
    }

    start() {
        let form = document.getElementById("form");
        form.remove();
    }

    update(evt) {
        ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        this.player.pos = evt.payload.player.position;

        console.log("Update", this.player);

        let otherPlayers = evt.payload.OtherPlayers;
        console.log("Others", otherPlayers);
        otherPlayers.forEach((otherPlayer) => {
            let oPlayer = new Player(otherPlayer.id, otherPlayer.name, otherPlayer.position)
            oPlayer.draw();
        })

        this.player.draw();
    }
}

let game;

document.addEventListener("keydown", function (event) {
    if (event.code == "KeyA") {
        sendMessage("move", "left");
    }
    if (event.code == "KeyD") {
        sendMessage("move", "right");
    }
    if (event.code == "KeyW") {
        sendMessage("move", "up");
    }
    if (event.code == "KeyS") {
        sendMessage("move", "down");
    }
});

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
            sendMessage("login", username);
            
            // console.log(game.player, "onopen");
            document.getElementById("connection-header").innerHTML =
                "Connected to Websocket: true";
            game.start();
        };

        conn.onmessage = function (event) {
            console.log(event);

            // parse websocket message as JSON
            const eventData = JSON.parse(event.data);
            console.log(eventData);
            // Assign JSON data to new Event Object
            const evt = Object.assign(new EventMsg(), eventData);

            routeEvent(evt);

            // game.update(evt);
        };
    } else {
        alert("Not supporting websockets");
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
 * SendMessageEvent is used to send messages to other players
 */
class SendMessageEvent {
    constructor(message, from) {
        this.message = message;
        this.from = from;
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
    switch (event.type) {
        case "login":
            console.log("new message");
            console.log(event.payload);
        case "move":
            console.log("new message");
            console.log(event.payload);
            break;

        default:
            alert("unsupported message type");
            break;
    }
}

/**
 * sendEvent
 * eventName - the event name to send on
 * payload - the data payload
 * */
function sendEvent(eventName, payload) {
    // Create a event Object with a event named send_message
    const event = new EventMsg(eventName, payload);
    // Format as JSON and send
    conn.send(JSON.stringify(event));
}

/**
 * sendMessage will send a new message onto the Websocket
 * */
function sendMessage(eventName, payload) {
    let msg;
    switch (eventName) {
        case "login":
            msg = payload;
            break;
        case "move":
            msg = { direction: payload };
            break;
    }
    let outGoingEvent = new SendMessageEvent(msg, game.player.id);
    console.log(outGoingEvent);
    sendEvent(eventName, outGoingEvent);
    return false;
}
