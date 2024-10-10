var selectedChat = "general";

class Event {
    constructor(type, payload) {
        this.type = type
        this.payload = payload
    }
}

class SendMessageEvent {
    constructor(message, from) {
        this.message = message
        this.from = from
    }
}
class NewMessageEvent {
    constructor(message, from, sent) {
        this.message = message
        this.from = from
        this.sent = sent
    }
}

function routeEvent(event) {
    if (event.type === undefined) {
        alert('no type field in the event')
    }
    switch (event.type) {
        case "new_message":
            const messageEvent = Object.assign(new NewMessageEvent(), event.payload)
            appendChatMessage(messageEvent)
            break
        default:
            alert("unsupported message type")
            break
    }
}

function appendChatMessage(messageEvent) {
    var date = new Date(messageEvent.sent)
    const formattedMsg = `${date.toLocaleString()} : ${messageEvent.message}`

    textarea = document.getElementById("chatmessages")
    textarea.innerHTML = textarea.innerHTML + "\n" + formattedMsg
    textarea.scrollTop = textarea.scrollHeight 
}

function sendEvent(eventName, payload) {
    const event = new Event(eventName, payload)

    conn.send(JSON.stringify(event))
}

function changeChatRoom() {
    var newChat = document.getElementById("chatroom");

    if (newChat != null && newchat.value != selectedChat) {
        console.log(newChat)
    }
    return false
}

function sendMessage() {
    var newMessage = document.getElementById("message");

    if (newMessage != null) {
        let outgoingEvent = new SendMessageEvent(newMessage.value, "user")
        sendEvent("send_message", outgoingEvent)
    }
    return false
}

function login() {
    let formData = {
        "username": document.getElementById("username").value,
        "password": document.getElementById("password").value,
    }
    fetch("login", {
        method: 'post',
        body: JSON.stringify(formData),
        mode: 'cors'
    }).then((response) => {
        if (response.ok) {
            return response.json()
        } else {
            throw 'unauthorized'
        }
    }).then((data) => {
        connectWebsocket(data.otp)
    }).catch((e) => { alert(e) })
    return false
}

function connectWebsocket(otp) {
    if (window["WebSocket"]) {
        console.log("supports websockets")
        conn = new WebSocket("ws://" + document.location.host + "/ws?otp=" + otp)

        conn.onopen = function (event) {
            document.getElementById("connection-header").innerHTML = "Connected to Websocket : true"
        }

        conn.onclose = function (event) {
            document.getElementById("connection-header").innerHTML = "Connected to Websocket : true"

        }

        console.log(document.location.host)

        conn.onmessage = function (event) {
            const eventData = JSON.parse(event.data)

            const newevent = Object.assign(new Event, eventData)

            routeEvent(newevent)
        }
    } else alert("not support websocket")
}

window.onload = function () {
    document.getElementById("chatroom-selection").onsubmit = changeChatRoom
    document.getElementById("chatroom-message").onsubmit = sendMessage
    document.getElementById("login-form").onsubmit = login
}  