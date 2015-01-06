var websocket;
var databases = {};
var reconnect = false;

window.onload = function () {
    window.onload = tablecloth;
    WebSocket_connect();
};

function WebSocket_connect_with_sleep() {
    window.setTimeout(WebSocket_connect, 1000)
}

function WebSocket_connect() {
    console.log("[Websocket debug] ==> Websocket url for connect: ws://" + location.host + "/ws");
    websocket = new WebSocket("ws://" + location.host + "/ws");
    websocket.onopen = function () {
        if (reconnect) {
            showSuccessToast("Websocket: Соединение восстановленно");
            reconnect = false;
            var mydiv = document.getElementById("main_data");
            mydiv.innerHTML = "";
        }
        console.log("[Websocket debug] ==> Соединение установлено.");
        var cmd = {
            'Cmd': 'subscribe',
            'Data': 'replicationGroups'
        };
        websocket.send(JSON.stringify(cmd));
        console.log("[Websocket debug] ==> Отправлены данные: "+JSON.stringify(cmd));
        var cmd = {
            'Cmd': 'subscribe',
            'Data': 'MySQLHost'
        };
        websocket.send(JSON.stringify(cmd));
        console.log("[Websocket debug] ==> Отправлены данные: "+JSON.stringify(cmd));
        var cmd = {
            'Cmd': 'subscribe',
            'Data': 'MySQLData'
        };
        websocket.send(JSON.stringify(cmd));
        console.log("[Websocket debug] ==> Отправлены данные: "+JSON.stringify(cmd));
        var cmd = {
            'Cmd': 'subscribe',
            'Data': 'databaseHosts'
        };
        websocket.send(JSON.stringify(cmd));
        console.log("[Websocket debug] ==> Отправлены данные: "+JSON.stringify(cmd));
        var cmd = {
            'Cmd': 'getDatabasesData',
            'Data': 'get'
        };
        websocket.send(JSON.stringify(cmd));
        console.log("[Websocket debug] ==> Отправлены данные: "+JSON.stringify(cmd));
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
        if (data.Command == "getDatabasesData") {
            databases = data.Data;
            InitPage(data.Data);
            return;
        }
        if (data.Command == "getHostData") {
            fillEditWindow(data.Data);
            return;
        }
        if (data.Channel == "MySQLHost") {
            if (data.Command == "add") {
                InsertDbHosts(data.Data);
                return;
            }
            if (data.Command == "delete") {
                console.log(data.Data);
                document.getElementById(data.Data).remove({opacity:0.3},800);
                // FIXME: remove from databases var
                return
            }
            if (data.Command == "update") {
                EditDbHosts(data.Data);
                return;
            }
        }
        if (data.Channel == "replicationGroups") {
            if (data.Command == "add") {
                databases[data.Data] = [];
                mydiv.innerHTML += CreateHeaderGroup(data.Data);
                mydiv.innerHTML += CreateTablePrototipe(data.Data);
                return
            }
            if (data.Command == "update") {
                var input = $.parseJSON(data.Data);
                var header = document.getElementById("header_" + input.From);
                header.id = "header_" + input["To"];
                header.innerHTML = CreateHeaderContent(input["To"]);
                delete databases[input["From"]];
                databases[input["To"]] = [];
                return
            }
            if (data.Command == "delete") {
                delete databases[data.Data];
                document.getElementById("header_"+data.Data).remove();
                document.getElementById(data.Data).remove();
                return
            }
        }
        if (data.Channel == "MySQLData" && data.Command == "update") {
            console.log(data.Data);
            for (var key in data.Data) {
                var table = document.getElementById(key).cells;
                table[2].innerHTML = 'Uptime: '+data.Data[key].Uptime;
                table[4].innerHTML = data.Data[key].Master.File+': '+data.Data[key].Master.Position;
            }

        }
    };
}

function CreateColunmDB(groupName) {
    return new("<div class='DBGHeader' id=\"header_" + groupName + "\"><br>" + groupName + "&nbsp;<a style='cursor: crosshair;'><cpan class='iconic pencil' style='font-size: small;' onclick='addReplicationGroup(\"" + groupName + "\")'><cpan></a>&nbsp;<a style='cursor: crosshair;'><cpan class='iconic trash' style='font-size: small;' onclick='deleteReplicationGroup(\"" + groupName + "\")'><cpan></a><br></div>")
}

function CreateHeaderGroup(groupName) {
    var res = "<div class='DBGHeader' id=\"header_" + groupName + "\">";
    res += CreateHeaderContent(groupName)+"</div>";
    return res

}

function CreateHeaderContent(groupName) {
    return "<br>" + groupName + "&nbsp;<a style='cursor: crosshair;'><cpan class='iconic pencil' style='font-size: small;' onclick='addReplicationGroup(\"" + groupName + "\")'><cpan></a>&nbsp;<a style='cursor: crosshair;'><cpan class='iconic trash' style='font-size: small;' onclick='deleteReplicationGroup(\"" + groupName + "\")'><cpan></a><br>"
}

function CreateTablePrototipe(dbgroup) {
    return "<table class='Databases' id='" + dbgroup + "'><thead><tr><th>Host</th><th>Port</th><th>Mysql Status</th><th>Slave thread</th><th>Master Status</th><th>Slave Status</th><th>Service</th></tr></thead><tbody></tbody></table>"
}

function InsertDbHosts(data) {
    var tableRef = document.getElementById(data["Group"]).getElementsByTagName('tbody')[0];
    var newRow = tableRef.insertRow(tableRef.rows.length);
    newRow.id = data["Ip"]+":"+data["Port"];
    var newCell = newRow.insertCell(0);
    var newText = document.createTextNode(data["Ip"]);
    newCell.appendChild(newText);
    newCell = newRow.insertCell(1);
    newText = document.createTextNode(data["Port"]);
    newCell.appendChild(newText);
    newCell = newRow.insertCell(2);
    newCell = newRow.insertCell(3);
    newCell = newRow.insertCell(4);
    newCell = newRow.insertCell(5);
    newCell = newRow.insertCell(6);
    newCell.innerHTML = "<a style='cursor: crosshair;'><cpan class='iconic settings' style='font-size: small;' title='Edit host' onclick='editMySQLHost(this)'><cpan></a>&nbsp;";
    newCell.innerHTML += "<a style='cursor: crosshair;'><cpan class='iconic trash' style='font-size: small;' onclick='DeleteDbHost(\""+data["Ip"]+":"+data["Port"]+"\")'><cpan></a>&nbsp;";
    newCell.innerHTML += "<a style='cursor: crosshair;'><cpan class='iconic check' style='font-size: small;' title='test' onclick='TestSqlInfo(\""+data["Group"]+"\", \""+data["Ip"]+"\", \""+data["Port"]+"\")'><cpan></a>&nbsp;";
}

function InitPage(data) {
    var mydiv = document.getElementById("main_data");
    mydiv.innerHTML = "";
    for (var key in data) {
        mydiv.innerHTML += CreateHeaderGroup(key);
        mydiv.innerHTML += CreateTablePrototipe(key);
        for (var i = 0; i < data[key].length; i++) {
            InsertDbHosts(data[key][i]);
        }
    }
}

function DeleteDbHost(host) {
    var cmd = {
        'Cmd': 'MySQLHostDelete',
        'Data': host
    };
    websocket.send(JSON.stringify(cmd));
}

function TestSqlInfo(group, host, port) {
    var hostInfo = {
        Host: host,
        Port: port
    };
    var cmd = {
         'Cmd': 'GetSlaveInfo',
         'Data': JSON.stringify(hostInfo)
    };
    websocket.send(JSON.stringify(cmd));
    console.log("[Websocket debug] ==> Отправлены данные: "+JSON.stringify(cmd));
    showStickyNoticeToast('<p style="font-size: small">test</p>');
}

function deleteReplicationGroup(name) {
    var options = {
    animation: 300,
    buttons: {
        close: {
            text: 'No',
            className: 'red',
            action: function(e) {
                Apprise('close');
            }
        },
        confirm: {
            text: 'Yes',
            className: 'blue',
            id: 'confirm',
            action: function(e) {
                var cmd = {
                    'Cmd': 'replicationGroupsDelete',
                    'Data': name
                };
                websocket.send(JSON.stringify(cmd));
                console.log(e);
                Apprise('close');
            }
        }
    }
  };
  Apprise('Are you really want to delete group '+name+' with all included hosts?',options);
}

function addReplicationGroup(edit) {
  var options = {
    animation: 300,
    buttons: {
        close: {
            text: 'Close',
            className: 'red',
            action: function(e) {
                Apprise('close');
            }
        },
        confirm: {
            text: 'Ok',
            className: 'blue',
            id: 'confirm',
            action: function(e) {
                if (edit) {
                    var data = {
                        From: edit,
                        To: e.input
                    }
                    var cmd = {
                        'Cmd': 'replicationGroupsEdit',
                        'Data': JSON.stringify(data)
                    };
                } else {
                    var cmd = {
                        'Cmd': 'replicationGroups',
                        'Data': e.input
                    };
                }
                websocket.send(JSON.stringify(cmd));
                console.log(e);
                Apprise('close');
            }
        },
    },
    input: true,
  };
  Apprise('Input new replication group name:',options);
}

function fillEditWindow(data) {
    console.log(data);
    var form = document.getElementById("MySQLUser");
    form.value = data.User;
    form = document.getElementById("MySQLSG");
    form.value = data.Group;
}

function editMySQLHost(obj) {
    var newHost = true;
    if (obj) {
        newHost = false;
        var TableRow = obj.parentNode.parentNode.parentNode.cells;
    }
    var options = {
    animation: 300,
    buttons: {
        close: {
            text: 'Close',
            className: 'red',
            action: function(e) {
                Apprise('close');
            }
        },
        confirm: {
            text: 'Ok',
            className: 'blue',
            id: 'confirm',
            action: function(e) {
                var data = {
                    ip: $('#hostIp').val(),
                    group: $("#MySQLSG option:selected").text(),
                    port: $('#MySQLPort').val(),
                    user: $('#MySQLUser').val(),
                    password: $('#MySQLPassword').val()
                };
                if (newHost) {
                    var cmd = {
                        'Cmd': 'MySQLHost',
                        'Data': JSON.stringify(data)
                    };
                } else {
                    var cmd = {
                        'Cmd': 'MySQLHostEdit',
                        'Data': JSON.stringify(data)
                    };
                }
                websocket.send(JSON.stringify(cmd));
                Apprise('close');
            }
        },
    },
    input: false,
  };
  var divInput = "";
  if (newHost) {
      divInput = "New host configuration:<br><br>" +
          "<div align='center'>" +
          '<table class="AddHost">' +
          '<tr><td>Host IP:</td><td><input placeholder="ip" id="hostIp" type="text" value="127.0.0.1"></input></td></tr>' +
          '<tr><td>MySQL Port:</td><td><input placeholder="port" id="MySQLPort" type="text" value="3306"></input></td></tr>' +
          '<tr><td>MySQL User:</td><td><input placeholder="user" id="MySQLUser" type="text" value="root"></input></td></tr>' +
          '<tr><td>MySQL Password:</td><td><input placeholder="password" id="MySQLPassword" type="password"></input></td></tr>' +
          '<tr><td>Replication group:</td><td><div class="dropdown"><select class="SelectRGroup" id="MySQLSG" title="Select Replication group" id="MySQLReplicaGroup"><option>Select Replication group</option>';
      for (key in databases) {
          divInput += "<option>" + key + "</option>";
      }
  } else {
      var cmd = {
            'Cmd': 'getHostData',
            'Data': TableRow[0].innerText+":"+TableRow[1].innerText
      };
      websocket.send(JSON.stringify(cmd));
      divInput = "Edit host's configuration:<br><br>" +
          "<div align='center'>" +
          '<table class="AddHost">' +
          '<tr><td>Host IP:</td><td><input placeholder="ip" id="hostIp" type="text" value="'+TableRow[0].innerText+'" readonly></input></td></tr>' +
          '<tr><td>MySQL Port:</td><td><input placeholder="port" id="MySQLPort" type="text" value="'+TableRow[1].innerText+'" readonly></input></td></tr>' +
          '<tr><td>MySQL User:</td><td><input placeholder="user" id="MySQLUser" type="text" value="root"></input></td></tr>' +
          '<tr><td>MySQL Password:</td><td><input placeholder="password" id="MySQLPassword" type="password"></input></td></tr>' +
          '<tr><td>Replication group:</td><td><div class="dropdown"><select class="SelectRGroup" id="MySQLSG" title="Select Replication group" id="MySQLReplicaGroup"><option>Select Replication group</option>';
      for (key in databases) {
          divInput += "<option>" + key + "</option>";
      }
  }
  divInput += '</select></div></td></tr></table></div>';
  Apprise(divInput,options);
  document.getElementById("hostIp").focus();
}

this.tablecloth = function(){

	// CONFIG

	// if set to true then mouseover a table cell will highlight entire column (except sibling headings)
	var highlightCols = true;

	// if set to true then mouseover a table cell will highlight entire row	(except sibling headings)
	var highlightRows = true;

	// if set to true then click on a table sell will select row or column based on config
	var selectable = true;

	// this function is called when
	// add your own code if you want to add action
	// function receives object that has been clicked
	this.clickAction = function(obj){
		//alert(obj.innerHTML);

	};

	// END CONFIG (do not edit below this line)

	this.cpanel = function(){
		var form = document.forms[0];
		form.onsubmit = function(){
			highlightCols = form.hc[0].checked;
			highlightRows = form.hr[0].checked;
			selectable = form.s[0].checked;
			unselectAll();
			return false;
		};
	};
	cpanel();


	var tableover = false;
	this.start = function(){
		var tables = document.getElementsByTagName("table");
		for (var i=0;i<tables.length;i++){
			tables[i].onmouseover = function(){tableover = true};
			tables[i].onmouseout = function(){tableover = false};
			rows(tables[i]);
		};
	};

	this.rows = function(table){
		var css = "";
		var tr = table.getElementsByTagName("tr");
		for (var i=0;i<tr.length;i++){
			css = (css == "odd") ? "even" : "odd";
			tr[i].className = css;
			var arr = new Array();
			for(var j=0;j<tr[i].childNodes.length;j++){
				if(tr[i].childNodes[j].nodeType == 1) arr.push(tr[i].childNodes[j]);
			};
			for (var j=0;j<arr.length;j++){
				arr[j].row = i;
				arr[j].col = j;
				if(arr[j].innerHTML == "&nbsp;" || arr[j].innerHTML == "") arr[j].className += " empty";
				arr[j].css = arr[j].className;
				arr[j].onmouseover = function(){
					over(table,this,this.row,this.col);
				};
				arr[j].onmouseout = function(){
					out(table,this,this.row,this.col);
				};
				arr[j].onmousedown = function(){
					down(table,this,this.row,this.col);
				};
				arr[j].onmouseup = function(){
					up(table,this,this.row,this.col);
				};
				arr[j].onclick = function(){
					click(table,this,this.row,this.col);
				};
			};
		};
	};

	// appyling mouseover state for objects (th or td)
	this.over = function(table,obj,row,col){
		if (!highlightCols && !highlightRows) obj.className = obj.css + " over";
		if(check1(obj,col)){
			if(highlightCols) highlightCol(table,obj,col);
			if(highlightRows) highlightRow(table,obj,row);
		};
	};
	// appyling mouseout state for objects (th or td)
	this.out = function(table,obj,row,col){
		if (!highlightCols && !highlightRows) obj.className = obj.css;
		unhighlightCol(table,col);
		unhighlightRow(table,row);
	};
	// appyling mousedown state for objects (th or td)
	this.down = function(table,obj,row,col){
		obj.className = obj.css + " down";
	};
	// appyling mouseup state for objects (th or td)
	this.up = function(table,obj,row,col){
		obj.className = obj.css + " over";
	};
	// onclick event for objects (th or td)
	this.click = function(table,obj,row,col){
		if(check1){
			if(selectable) {
				unselect(table);
				if(highlightCols) highlightCol(table,obj,col,true);
				if(highlightRows) highlightRow(table,obj,row,true);
				document.onclick = unselectAll;
			}
		};
		clickAction(obj);
	};

	this.highlightCol = function(table,active,col,sel){
		var css = (typeof(sel) != "undefined") ? "selected" : "over";
		var tr = table.getElementsByTagName("tr");
		for (var i=0;i<tr.length;i++){
			var arr = new Array();
			for(j=0;j<tr[i].childNodes.length;j++){
				if(tr[i].childNodes[j].nodeType == 1) arr.push(tr[i].childNodes[j]);
			};
			var obj = arr[col];
			if (check2(active,obj) && check3(obj)) obj.className = obj.css + " " + css;
		};
	};
	this.unhighlightCol = function(table,col){
		var tr = table.getElementsByTagName("tr");
		for (var i=0;i<tr.length;i++){
			var arr = new Array();
			for(j=0;j<tr[i].childNodes.length;j++){
				if(tr[i].childNodes[j].nodeType == 1) arr.push(tr[i].childNodes[j])
			};
			var obj = arr[col];
			if(check3(obj)) obj.className = obj.css;
		};
	};
	this.highlightRow = function(table,active,row,sel){
		var css = (typeof(sel) != "undefined") ? "selected" : "over";
		var tr = table.getElementsByTagName("tr")[row];
		for (var i=0;i<tr.childNodes.length;i++){
			var obj = tr.childNodes[i];
			if (check2(active,obj) && check3(obj)) obj.className = obj.css + " " + css;
		};
	};
	this.unhighlightRow = function(table,row){
		var tr = table.getElementsByTagName("tr")[row];
		for (var i=0;i<tr.childNodes.length;i++){
			var obj = tr.childNodes[i];
			if(check3(obj)) obj.className = obj.css;
		};
	};
	this.unselect = function(table){
		tr = table.getElementsByTagName("tr")
		for (var i=0;i<tr.length;i++){
			for (var j=0;j<tr[i].childNodes.length;j++){
				var obj = tr[i].childNodes[j];
				if(obj.className) obj.className = obj.className.replace("selected","");
			};
		};
	};
	this.unselectAll = function(){
		if(!tableover){
			tables = document.getElementsByTagName("table");
			for (var i=0;i<tables.length;i++){
				unselect(tables[i])
			};
		};
	};
	this.check1 = function(obj,col){
		return (!(col == 0 && obj.className.indexOf("empty") != -1));
	}
	this.check2 = function(active,obj){
		return (!(active.tagName == "TH" && obj.tagName == "TH"));
	};
	this.check3 = function(obj){
		return (obj.className) ? (obj.className.indexOf("selected") == -1) : true;
	};

	start();

};

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

