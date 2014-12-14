var websocket;

window.onload = function () {
    var outDiv = document.getElementById("main_data");
    outDiv.innerHTML = "[Websocket debug] ==> Websocket url for connect: ws://" + location.host + "/ws";
    websocket = new WebSocket("ws://" + location.host + "/ws");
    websocket.onopen = function () {
        outDiv.innerHTML += "<br>[Websocket debug] ==> Соединение установлено.";
        var cmd = {
            'Cmd': 'test',
            'Data': 'message from client'
        };
        websocket.send(JSON.stringify(cmd));
        outDiv.innerHTML += "<br>[Websocket debug] ==> Отправлены данные: ping ["+JSON.stringify(cmd)+"]"
        InterfaceState(true);
    };
    websocket.onclose = function (event) {
        if (event.wasClean) {
            outDiv.innerHTML += '<br>[Websocket debug] ==> Exit: Соединение закрыто чисто';
        } else {
            outDiv.innerHTML += '<br>[Websocket debug] ==> Exit: Обрыв соединения'; // например, "убит" процесс сервера
        }
            outDiv.innerHTML += '<br>[Websocket debug] ==> Код: ' + event.code + ' причина: ' + event.reason;
        };

        websocket.onerror = function (error) {
            outDiv.innerHTML += "<br>[Websocket debug] ==> Ошибка " + error.message;
        };

        websocket.onmessage = function (event) {
            outDiv.innerHTML += "<br>[Websocket debug] ==> Получены данные: " + event.data;
            websocket_msg = event.data;
        };

};

window.onbeforeunload = function() {
        websocket.onclose = function () {}; // disable onclose handler first
        websocket.close();
        console.log("Socket close")
};

function doNothing() {}

jQuery(function ($, undefined) {
        /*      $('#term').terminal(function (command, term) {
         if (command !== '') {
         if (socket.readyState == socket.CLOSED) {
         term.error("Socket to server was closed. Trying to reconnect...");
         socket = WSConnect();
         if (socket.readyState != 1) {

         return;
         };
         };
         socket.send(command);
         socket.onmessage = function(message) {
         console.log("The server says the answer is: " + message);
         if (message == '') {
         term.error("Data wasn't recieve")
         } else {
         term.echo(new String(message.data));
         }
         }
         }
         }, {
         greetings: 'RLTMonMaker Interpreter',
         name: 'js_term',
         height: 400,
         prompt: 'rlt> '});*/
});

$('nav').click(function(){
    var that = $(this);
    console.log(that);
      that.closest('a').find('.selected').removeClass('selected');
        that.addClass('selected');
});

$(function() {
          $("#nav a").click(function() {
              var selected = document.getElementById("nav").getElementsByClassName("selected")[0];
              $(this).addClass('selected');
              selected.removeAttribute('class');
              });
});

function InterfaceSettings() {
    $("#main_data").html("");
    var cmd = {
                'Cmd': 'unsubscribe',
                'Data': 'connectlist'
    };
    websocket.send(JSON.stringify(cmd));
}

function InterfaceState(first) {
    outDiv = document.getElementById("main_data");
    if (first == false) {
        $("#main_data").html("");
    }
    var cmd = {
        'Cmd': 'subscribe',
        'Data': 'connectlist'
    };
    websocket.send(JSON.stringify(cmd));
    outDiv.innerHTML += "<br>[Event subscribe] ==> Trying to subscribe event ["+JSON.stringify(cmd)+"]"
}

