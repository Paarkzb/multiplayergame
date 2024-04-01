import Game from "./game";
import { routeEvent, sendEvent } from "./event";

let game = new Game();

/**
 * login will send a login request to the server and then connect websocket
 */
export function login() {
    connectWebsocket();

    return false;
}

document.getElementById("login-btn").addEventListener('click', login);

/**
 * connectWebsocket will connect to websocket and add listeners
 */
function connectWebsocket() {
    if (window["WebSocket"]) {
        console.log("supports websockets");
        game.conn = new WebSocket("ws://localhost:8080/ws");

        // Onopen
        game.conn.onopen = function (evt) {
            let username = document.getElementById("username").value;
            sendEvent("login", username);
            
            game.start();
            game.animationFrame = requestAnimationFrame(() => game.update());
            
            game.addListenters();
        };

        game.conn.onmessage = function (event) {
            // parse websocket message as JSON
            const eventData = JSON.parse(event.data);
            console.log(eventData);
            routeEvent(eventData, game);

        };

        game.conn.onclose = function (event) {
            cancelAnimationFrame(game.animationFrame);
        }
    } else {
        alert("Not supporting websockets");
    }
}