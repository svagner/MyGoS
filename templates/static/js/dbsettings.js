var websocket;
var reconnect = false;
var initAdd = false;

function WebSocket_connect_with_sleep() {
    window.setTimeout(WebSocket_connect, 1000)
}

//$(function() {
function stepsInit() {
    // there's the stepsList and the selectedSteps
    var $stepsList = $( "#stepsList" ),
      $selectedSteps = $( "#selectedSteps" );
 
    // let the stepsList items be draggable
    $( "li", $stepsList ).draggable({
      cancel: "a.ui-icon", // clicking an icon won't initiate dragging
      revert: "invalid", // when not dropped, the item will revert back to its initial position
      containment: "document",
      helper: "clone",
      cursor: "move"
    });
 
    // let the selectedSteps be droppable, accepting the stepsList items
    $selectedSteps.droppable({
      accept: "#stepsList > li",
      activeClass: "ui-state-highlight",
      drop: function( event, ui ) {
        addStepToSelected( ui.draggable );
      }
    });
 
    // let the stepsList be droppable as well, accepting items from the selectedSteps
    $stepsList.droppable({
      accept: "#selectedSteps li",
      activeClass: "custom-state-active",
      drop: function( event, ui ) {
        deleteStepFromSelected( ui.draggable );
      }
    });
 
    // image deletion function
    var recycle_icon = "<a href='link/to/recycle/script/when/we/have/js/off' title='Recycle this image' class='ui-icon ui-icon-refresh'>Recycle image</a>";
    function addStepToSelected( $item ) {
      $item.fadeOut(function() {
        var $list = $( "ul", $selectedSteps ).length ?
          $( "ul", $selectedSteps ) :
          $( "<ul class='stepsList ui-helper-reset'/>" ).appendTo( $selectedSteps );
 
        $item.find( "a.ui-icon-selectedSteps" ).remove();
        $item.append( recycle_icon ).appendTo( $list ).fadeIn(function() {
          $item
          .animate({ width: "96%" });
        });
      });
    }
 
    // image recycle function
    var selectedSteps_icon = "<a href='link/to/selectedSteps/script/when/we/have/js/off' title='Delete this image' class='ui-icon ui-icon-selectedSteps'>Delete image</a>";
    function deleteStepFromSelected( $item ) {
      $item.fadeOut(function() {
        $item
          .find( "a.ui-icon-refresh" )
            .remove()
          .end()
          .css( "width", "auto")
          .append( selectedSteps_icon )
          .find( "img" )
            .css( "height", "72px" )
          .end()
          .appendTo( $stepsList )
          .fadeIn();
      });
    }

    function viewScriptContent( $item ) {
        addReplicationScript($item);
    }

    function deleteStep( $item ) {
        var name = $item[0].parentNode.getElementsByTagName('h5')[0].innerText;
        var cmd = {
            'Cmd': 'deleteReplicationStepScript',
            'Data': name
        };
        websocket.send(JSON.stringify(cmd));
    }

    function testStep( $item ) {
        var name = $item[0].parentNode.getElementsByTagName('h5')[0].innerText;
        var cmd = {
            'Cmd': 'testReplicationStepScript',
            'Data': name
        };
        websocket.send(JSON.stringify(cmd));
    }

    // resolve the icons behavior with event delegation
    $( "ul.stepsList > li" ).click(function( event ) {
      var $item = $( this ),
        $target = $( event.target );
 
      if ( $target.is( "a.ui-icon-selectedSteps" ) ) {
        addStepToSelected( $item );
      } else if ( $target.is( "a.ui-icon-refresh" ) ) {
        deleteStepFromSelected( $item );
      } else if ( $target.is( "a.ui-icon-zoomin" ) ) {
        viewScriptContent( $target );
      } else if ( $target.is( "a.ui-icon-trash" ) ) {
        deleteStep( $target );
      } else if ( $target.is( "a.ui-icon-play" ) ) {
        testStep( $target );
      }
 
      return false;
    });
  };//);

window.onload = function () {
    WebSocket_connect();
    stepsInit();
};

function saveSetting() {
    var selectedSteps = [];
    var lis = document.getElementById('selectedSteps').getElementsByTagName('li');
    for (var i = 0; i < lis.length; i++) {
        selectedSteps.push(lis[i].getElementsByTagName('h5')[0].innerText)
    }
    console.log(selectedSteps);
    var cmd = {
        'Cmd': 'saveReplicationStepsSelected',
        'Data': JSON.stringify(selectedSteps)
    };
    websocket.send(JSON.stringify(cmd));
}

function WebSocket_connect() {
    console.log("[Websocket debug] ==> Websocket url for connect: ws://" + location.host + "/ws");
    websocket = new WebSocket("ws://" + location.host + "/ws");
    websocket.onopen = function () {
        if (reconnect) {
            showSuccessToast("Websocket: Соединение восстановленно");
            reconnect = false;
        }
        console.log("[Websocket debug] ==> Соединение установлено.");
        var cmd = {
            'Cmd': 'subscribe',
            'Data': 'replicationSteps'
        };
        websocket.send(JSON.stringify(cmd));
    };
    websocket.onclose = function (event) {
        if (event.wasClean) {
            showStickyNoticeToast('Websocket: Cоединение закрыто по таймауту');
            console.log('[Websocket debug] ==> Exit: Соединение закрыто чисто');
        } else {
            reconnect = true;
            showErrorToastFunc('Websocket: Обрыв соединения' +
                '<br>код: ' + event.code, WebSocket_connect_with_sleep);
        }
        console.log('[Websocket debug] ==> Код: ' + event.code + ' причина: ' + event.reason);
    };

    websocket.onerror = function (error) {
        console.log("[Websocket debug] ==> Ошибка " + error.message);
    };

    websocket.onmessage = function (event) {
        var data = $.parseJSON(event.data);
        data = JSON.parse(data);
        var mydiv = document.getElementById("main_data");
        console.log("[Websocket debug] ==> Получены данные: " + data);
        if (data.Channel == "Error") {
            showErrorToast(data.Data);
            return;
        }
        if (data.Channel == "replicationSteps") {
            if (data.Command == "update") {
                initAdd = true;
                for (var ii = 0; ii < data.Data.length; ii++) {
                    var SList = document.getElementById("stepsList");
                    var steps = SList.getElementsByClassName('ui-icon-selectedSteps');
                    for (var i = 1; i < steps.length; i++) {
                        var text = steps[i].parentNode.getElementsByTagName('h5')[0].textContent;
                        console.log(text, data.Data[ii]);
                        if (text == data.Data[ii]) {
                            steps[i].click();
                            break;
                        }
                    }
                    if (ii == data.Data.length - 1) {
                        initAdd = false;
                    }
                }
                return;
            }
            if (data.Command == "add") {
                if (data.Data.Content != '') {
                    stepsList[data.Data.Name] = data.Data.Content;
                }
                if (data.Data.Content != '') {
                    document.getElementById("stepsList").innerHTML += '<li class="ui-widget-content ui-corner-tr"> \
                <h5 class="ui-widget-header">' + data.Data.Name + '</h5> \
                <a href="#" title="View step content" class="ui-icon ui-icon-zoomin">View step content</a> \
                <a href="#" title="Run test" class="ui-icon ui-icon-play">Run test</a> \
                <a href="#" title="Delete step" class="ui-icon ui-icon-trash">Delete step</a> \
                <a href="link/to/selectedSteps/script/when/we/have/js/off" title="Select step" class="ui-icon ui-icon-selectedSteps">Select step</a></li>';
                    console.log(document.getElementById("stepsList").innerHTML);
                } else {
                    document.getElementById("stepsList").innerHTML += '<li class="ui-widget-content ui-corner-tr"> \
                <h5 class="ui-widget-header">' + data.Data.Name + '</h5> \
                <a href="link/to/selectedSteps/script/when/we/have/js/off" title="Select step" class="ui-icon ui-icon-selectedSteps">Select step</a></li>';
                }
                stepsInit();
            }
            if (data.Command == "delete") {
                console.log(data.Data);
                delete stepsList[data.Data];
                var parent = document.getElementById('selectedSteps').getElementsByTagName('ul')[0];
                var s = parent.getElementsByTagName('li');
                for (var i = 0; i < s.length; i++) {
                    var Node = s[i].getElementsByTagName('h5');
                    if (Node[0].innerText == data.Data) {
                        console.log(parent);
                        parent.removeChild(s[i]);
                    }
                }
                var parent = document.getElementById('stepsList');
                var s = parent.getElementsByTagName('li');
                for (var i = 0; i < s.length; i++) {
                    var Node = s[i].getElementsByTagName('h5');
                    if (Node[0].innerText == data.Data) {
                        parent.removeChild(s[i]);
                    }
                }
                stepsInit();
            }
            if (data.Command == "reinit") {
                var parent = document.getElementById('selectedSteps').getElementsByTagName('ul')[0];
                var s = parent.getElementsByTagName('li');
                for (var i = 0; i < s.length; i++) {
                    var text = s[i].getElementsByTagName('h5')[0].innerText;
                    if (data.Data.indexOf(text) < 0) {
                        s[i].getElementsByTagName('a')[0].click();
                    }
                }
                var SList = document.getElementById("stepsList");
                var steps = SList.getElementsByClassName('ui-icon-selectedSteps');
                console.log(steps.length);
                for (var i = 1; i < steps.length; i++) {
                    console.log(i);
                    var text = steps[i].parentNode.getElementsByTagName('h5')[0].innerText;
                    console.log(text, '=', data.Data[ii]);
                    if (data.Data.indexOf(text) >= 0) {
                        steps[i].click();
                    }
                }
                console.log(data.Data);
            }
        }
        if (data.Channel == "testReplicationStepScript") {
            var res = JSON.parse(data.Data);
            if (data.Command == "error") {
                testReplicationScript(res, true);
            } else if (data.Command == "result") {
                testReplicationScript(res, false);
            }
        }
    };
}

function testReplicationScript(data, error) {
    var options = {
        animation: 300,
        buttons: {
            close: {
                text: 'Close',
                className: 'red',
                action: function (e) {
                    Apprise('close');
                }
            }
        },
        input: false
    };
  if (error) {
      var divInput = "<div><h2>Error:</h2><br><textarea readonly style='width: 100%; height: 100%'>"+data.Error+"</textarea></div>";
  } else {
      var divInput = "<div><h2>Result:</h2><br><textarea readonly style='width: 100%; height: 100%'>"+data.Data+"</textarea></div>";
  }

  Apprise(divInput,options);
}

function addReplicationScript(obj) {
    var editor;
    var newScript = true;
    if (obj) {
        newScript = false;
        var name = obj[0].parentNode.getElementsByTagName('h5')[0].innerText;
    }
    var options = {
        animation: 300,
        buttons: {
            close: {
                text: 'Close',
                className: 'red',
                action: function (e) {
                    Apprise('close');
                }
            },
            submit: {
                text: 'Submit',
                className: 'blue',
                id: 'confirm',
                action: function (e) {
                    if (document.getElementById("scriptname").value == '') {
                        document.getElementById("scriptname").style.borderColor = 'red';
                        return
                    } else {
                        document.getElementById("scriptname").style.borderColor = 'green';
                    }

                    if (editor.getValue() == '') {
                        document.getElementsByClassName("CodeMirror")[0].style.background = 'lightpink';
                        return
                    }
                    var data = {
                        name: document.getElementById("scriptname").value,
                        content: editor.getValue()
                    };
                    if (newScript) {
                        var cmd = {
                            'Cmd': 'addReplicationStepScript',
                            'Data': JSON.stringify(data)
                        };
                    } else {
                        var cmd = {
                            'Cmd': 'editReplicationStepScript',
                            'Data': JSON.stringify(data)
                        };
                    }
                    websocket.send(JSON.stringify(cmd));
                    Apprise('close');
                }
            }
        },
        input: false
    };
  var divInput = "";
  if (newScript) {
      divInput = '<div>Script name: <input id="scriptname" name="scriptname" type="text" maxlength="15"></input></div><br> \
<form><textarea id="code" name="code"></textarea></form>';
  } else {
      divInput = '<div>Script name: <input id="scriptname" name="scriptname" type="text" maxlength="15" value="'+name+'"></input></div><br> \
<form><textarea id="code" name="code"></textarea></form>';
  }

  Apprise(divInput,options);
  editor = CodeMirror.fromTextArea(document.getElementById("code"), {
        lineNumbers: true,
        styleActiveLine: true,
        matchBrackets: true,

  });
  editor.on('focus', function () {
            document.getElementsByClassName("CodeMirror")[0].style.background = 'white';
  });

  editor.setOption("theme", "eclipse");
  if (!newScript) {
      editor.replaceRange(stepsList[name], CodeMirror.Pos(editor.lastLine()));
  }
  document.getElementById("scriptname").focus();
}

function showSuccessToast(message) {
    $().toastmessage('showSuccessToast', message);
}

function showStickySuccessToast(message) {
    $().toastmessage('showToast', {
        text     : message,
        sticky   : true,
        position : 'top-right',
        type     : 'success',
        closeText: '',
        close    : function () {
            console.log("toast is closed ...");
        }
    });

}

function showNoticeToast(message) {
        $().toastmessage('showNoticeToast', message);
    }

function showStickyNoticeToast(message) {
        $().toastmessage('showToast', {
             text     : message,
             sticky   : true,
             position : 'top-right',
             type     : 'notice',
             closeText: '',
             close    : function () {console.log("toast is closed ...");}
        });
    }

function showWarningToast(message) {
        $().toastmessage('showWarningToast', message);
    }

function showStickyWarningToast(message) {
        $().toastmessage('showToast', {
            text     : message,
            sticky   : true,
            position : 'top-right',
            type     : 'warning',
            closeText: '',
            close    : function () {
                console.log("toast is closed ...");
            }
        });
    }

function showErrorToast(message) {
        $().toastmessage('showErrorToast', message);
    }

function showStickyErrorToast(message) {
        $().toastmessage('showToast', {
            text     : message,
            sticky   : true,
            position : 'top-right',
            type     : 'error',
            closeText: ''
        });
    }

function showErrorToastFunc(message, func) {
        $().toastmessage('showToast', {
            text     : message,
            sticky   : false,
            position : 'top-right',
            type     : 'error',
            closeText: '',
            close : func()
        });
    }


function htmlEntities(str) {
    return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}