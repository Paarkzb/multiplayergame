/**
 * login will send a login request to the server and then connect websocket
 */
function login() {
    let formData = {
        username: document.getElementById("username").value,
    };
    // Send the request
    fetch("http://localhost:8080/login", {
        method: "POST",
        body: JSON.stringify(formData),
        mode: "cors",
    })
        .then((response) => {
            console.log(response);
            if (response.ok) {
                connectWebsocket();
            } else {
                throw "unauthorized";
            }
        })
        .catch((e) => {
            alert(e);
        });

    return false;
}

/**
 * connectWebsocket will connect to websocket and add listeners
 */
function connectWebsocket() {
    if (window["WebSocket"]) {
        console.log("supports websockets");
        conn = new WebSocket("ws://" + "localhost:8080" + "/ws");

        // Onopen
        conn.onopen = function (evt) {
            sendMessage();
            document.getElementById("connection-header").innerHTML =
                "Connected to Websocket: true";
        };

        conn.onclose = function (evt) {
            // Set disconnected
            document.getElementById("connection-header").innerHTML =
                "Connected to Websocket: false";
        };

        conn.onmessage = function (event) {
            console.log(event);

            // parse websocket message as JSON
            const eventData = JSON.parse(event.data);
            // Assign JSON data to new Event Object
            const evt = Object.assign(new Event(), eventData);

            routeEvent(evt);
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
class Event {
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
 * NewMessageEvent is messages comming from clients
 */
class NewMessageEvent {
    constructor(message, from, sent) {
        this.message = message;
        this.from = from;
        this.sent = sent;
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
        case "new_message":
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
    const event = new Event(eventName, payload);
    // Format as JSON and send
    conn.send(JSON.stringify(event));
}
/**
 * sendMessage will send a new message onto the Websocket
 * */
function sendMessage() {
    let outGoingEvent = new SendMessageEvent(
        "Here's some text that the server is urgently awaiting!",
        "paark"
    );
    sendEvent("send_message", outGoingEvent);
    return false;
}

/**
 * Once the website loads
 */
window.onload = function () {
    // Apply listener functions to the submit event
    document.getElementById("login-form").onsubmit = login;
};
