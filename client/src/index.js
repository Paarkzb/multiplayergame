class Player {
    animationFrame;

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
        // console.log("DRAW", this.pos.x, this.pos.y, this.width, this.height, this.angle);
        this.ctx.translate(this.pos.x + this.width/2, this.pos.y + this.height/2 );
        this.ctx.rotate(this.angle);
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
        this.firstServerTimestamp = 0;
        this.gameStart = 0;
        this.serverDelay = 100;
        this.states = [];
        this.animationFrame;
    }

    start() {
        let form = document.getElementById("form");
        form.remove();
    }

    update() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        
        const {player, otherPlayers} = this.getCurrentState();
        console.log("UPDATE", player);
        if(player) {
            console.log("Update", player);
            console.log("Others", otherPlayers);

            let playerObject = new Player(player.id, player.name, player.position, player.angle, this.ctx);
            playerObject.draw();
            otherPlayers?.forEach((otherPlayer) => {
                let oPlayer = new Player(otherPlayer.id, otherPlayer.name, otherPlayer.position, otherPlayer.angle, this.ctx);
                oPlayer.draw();
            })

        }

        this.animationFrame = requestAnimationFrame(() => this.update());
    }

    getCurrentState() {
        if(!this.firstServerTimestamp) {
            return {};
        }

        const stateIndex = this.getStateIndex();
        const serverTime = this.getServerTime();
        console.log("STATE INDEX", stateIndex);
        console.log("GET STATE", this.states);
        if(stateIndex < 0 || stateIndex === this.states.length - 1) {
            return {
                player: this.states[this.states.length - 1].player,
                otherPlayers: this.states[this.states.length - 1].otherPlayers,
            };
        } else {
            const currentState = this.states[stateIndex];
            const nextState = this.states[stateIndex + 1];
            const ratio = (serverTime - currentState.timestamp) / (nextState.timestamp - currentState.timestamp);
            return {
                player: this.lerp(currentState.player, nextState.player, ratio),
                otherPlayers: this.lerpArray(currentState.otherPlayers, nextState.otherPlayers, ratio),
            }
        }
    }

    setCurrentState(evt){
        if(!this.firstServerTimestamp) {
            this.firstServerTimestamp = evt.timestamp;
            this.gameStart = Date.now();
        }
        console.log("setCurrentState", evt);
        this.states.push(evt);
        console.log("SERVER TIME", this.getServerTime());
        console.log("STATES", this.states);

        const stateIndex = this.getStateIndex();
        if(stateIndex > 0) {
            this.states.splice(0, stateIndex);
        }
    }

    getServerTime(){
        return this.firstServerTimestamp + (Date.now() - this.gameStart) - this.serverDelay;
    }

    getStateIndex() {
        const serverTime = this.getServerTime();
        for(let i = this.states.length - 1; i >= 0; i--) {
            if(this.states[i].timestamp <= serverTime) {
                return i;
            }
        }
        return -1;
    }

    lerp(start, end, t) {
        
        if(!end) {
            return start;
        }

        start.position.x = start.position.x * (1 - t) + end.position.x * t;
        start.position.y = start.position.y * (1 - t) + end.position.y * t;

        start.angle = this.lerpAngle(start.angle, end.angle, t);
        // start.angle = start.angle * (1 - t) + end.angle * t;;

        return start;

        // const interpolated = {};
        // Object.keys(start).forEach(key => {
        //     if(key === "position") {
        //         interpolated[key] = {};
        //         interpolated[key].x = start[key].x * (1 - t) + end[key].x * t;
        //         interpolated[key].y = start[key].y * (1 - t) + end[key].y * t;
        //     } else {
        //         interpolated[key] = start[key];
        //     }
        // });

        // return interpolated;
    }

    lerpArray(startArray, endArray, t) {
        return startArray.map(elem => this.lerp(elem, endArray.find(elem2 => elem.id === elem2.id), t));
        // return startArray.map((elem, index) => this.lerp(elem, endArray[index], t));
    }

    lerpAngle(startAngle, endAngle, t) {
        // const absD = Math.abs(startAngle - endAngle);
        // if(absD >= Math.PI) {
        //     if(startAngle > endAngle) {
        //         return startAngle + (endAngle + 2 * Math.PI - startAngle) * t
        //     } else {
        //         return startAngle - (endAngle - 2 * Math.PI - startAngle) * t
        //     }
        // } else {
        //     return startAngle + (endAngle - startAngle) * t;
        // }

        // const absD = Math.abs(startAngle - endAngle);
        // if(absD >= Math.PI) {
        //     if(startAngle > endAngle) {
        //         return startAngle * (1 - t) + (endAngle + 2 * Math.PI) * t
        //     } else {
        //         return startAngle * (1 - t) - (endAngle - 2 * Math.PI) * t
        //     }
        // } else {
        //     return startAngle * (1 - t) + endAngle * t;
        // }

        const max = Math.PI * 2;
        const da = (endAngle - startAngle) % max;
        return startAngle + (2 * da % max - da) * t;
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
            game.animationFrame = requestAnimationFrame(() => game.update());
            
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

            // parse websocket message as JSON
            const eventData = JSON.parse(event.data);
            console.log(eventData);
            routeEvent(eventData);

        };

        conn.onclose = function (event) {
            cancelAnimationFrame(game.animationFrame);
        }
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
    console.log("EVENT FROM SERVER", event);
    switch (event.type) {
        case "update":
            game.setCurrentState(event);
            // if(!game.firstUpdate) {
            //     // setInterval(game.update(), 1000 / 60);
            //     game.update();
            //     game.firstUpdate = true;
            // }
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

