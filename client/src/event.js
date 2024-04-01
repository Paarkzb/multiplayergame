/**
 * Event is used to wrap all messages Send and Receive
 * on the Websocket
 * The type is used as a RPC
 **/
export class EventMsg {
    constructor(type, payload) {
        this.type = type;
        this.payload = payload;
    }
}

/**
 * routeEvent is a proxy function that routes
 * events into their correct Handler
 * based on the type field
 * */
export function routeEvent(event, game) {
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
 * sendEvent
 * eventName - the event name to send on
 * payload - the data payload
 * */
export function sendEvent(eventName, payload, conn) {
    // Create a event Object with a event named send_message
    const event = new EventMsg(eventName, payload);
    console.log("SENDEVENT", event);
    conn.send(JSON.stringify(event));
}
