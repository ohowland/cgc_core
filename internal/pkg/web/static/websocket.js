const socket = new WebSocket('ws://localhost:8000/ws');
console.log('opened websocket');

socket.addEventListener('message', function (event) {
    console.log('Message from server ', event.data);
    var msg = JSON.parse(event.data)
    var name = "." + msg.asset
    $(name + "-name").text(msg.asset)
    $(name + "-kw").text(msg.kw)
    $(name + "-kvar").text(msg.kvar)

});