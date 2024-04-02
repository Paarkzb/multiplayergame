import Player from "./player";

export default class Game {
    constructor() {
        this.canvas = document.getElementById("game-canvas");
        this.canvas.width = 1920;
        this.canvas.height = 1080;
        this.ctx = this.canvas.getContext("2d");

        this.firstServerTimestamp = 0;
        this.gameStart = 0;
        this.serverDelay = 100;
        this.states = [];

        this.animationFrame;

        this.conn;
    }

    start() {
        let form = document.getElementById("form");
        form.remove();
    }

    update() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        this.ctx.beginPath();
        this.ctx.fillStyle = "gray";
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
        this.ctx.closePath();

        
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
                player: this.lerpObject(currentState.player, nextState.player, ratio),
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
        return start * (1 - t) + end * t;
    }

    lerpObject(start, end, t) { 
        if(!end) {
            return start;
        }

        start.position.x = this.lerp(start.position.x, end.position.x, t);
        start.position.y = this.lerp(start.position.y, end.position.y, t);

        start.angle = this.lerpAngle(start.angle, end.angle, t);

        return start;
    }

    lerpArray(startArray, endArray, t) {
        return startArray.map(elem => this.lerpObject(elem, endArray.find(elem2 => elem.id === elem2.id), t));
        // return startArray.map((elem, index) => this.lerp(elem, endArray[index], t));
    }

    repeat(t, m) {
        return Math.min(m, Math.max(0, t - Math.floor(t / m) * m));
    }

    lerpAngle(startAngle, endAngle, t) {
        const dt = this.repeat(endAngle - startAngle, 2 * Math.PI);
        return this.lerp(startAngle, startAngle + (dt > Math.PI ? dt - 2 * Math.PI : dt), t);
    }
}