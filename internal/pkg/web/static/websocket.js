const socket = new WebSocket('ws://localhost:8000/ws');
console.log('opened websocket');

socket.addEventListener('message', function (event) {
    console.log('Message from server ', event.data);
    var msg = JSON.parse(event.data);
    var name = msg.asset
    $( "#" + name + "_name").text(msg.asset)
    $( "#" + name + "_kw").text(msg.kw + " kW")
    $( "#" + name + "_kvar").text(msg.kvar + " kVAR")

});

function updateRunCommand(id, asset_name){
    var run=$( '#' + id ).val();
    if (run == "Start") {
        $( '#' + id ).val("Stop")
    } else {
        $( '#' + id ).val("Start")
    };
    sendFormData(asset_name)
}

function sendFormData(name){
    var kw=$("#" + name + "_kw_in").val();
    var kvar=$("#" + name + "_kvar_in").val();
    var run=$("#" + name + "_run_in").val();
    msg =  JSON.stringify({
        Asset: name,
        Run: run == "Stop",
        KW: kw,
        KVAR: kvar
    });
    console.log(msg);
    socket.send(msg);
}