var xbtnClick;

window.onload = function () {
    var conn;
    var msg = document.getElementById("msg");
    var log = document.getElementById("log");
    var bottomBar = document.getElementById("bottom-bar");
    var buttons = document.getElementById("buttons-wrapper");
    var logout = document.getElementById("logout");
    var showButtons = document.getElementById("show-buttons");
    
    var noSleep = new NoSleep();

    function appendLog(item) {
        var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        log.appendChild(item);
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    };

    function sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    };

    autoscroll = function() {
        log.scrollTop = log.scrollHeight;
        log.focus();
        log.blur();
    }

    function startWs() {
        connected = false;
        if (location.protocol == 'https:') {
            conn = new WebSocket("wss://" + document.location.host + "/ws");
        } else {
            conn = new WebSocket("ws://" + document.location.host + "/ws");
        }

        conn.onopen = function (evt) {
            var item = document.createElement("div");
            item.innerHTML = "<b>Connected.</b>";
            appendLog(item);
            connected = true;
        };

        conn.onclose = function (evt) {
            if (connected) {
                var item = document.createElement("div");
                item.innerHTML = "<b>Disconnected</b>";
                appendLog(item);
            }
            setTimeout(function(){startWs()}, 10000);
            connected = false;
        };
        
        conn.onmessage = function (evt) {
            var messages = evt.data.split('\n');
            for (var i = 0; i < messages.length; i++) {
                var item = document.createElement("div");
                item.innerText = messages[i];
                appendLog(item);
            }
            autoscroll();
        }

        return conn;
    }

    logout.onclick = function() {
        document.cookie = "session=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
        window.location.replace("/logout");
    };

    showButtons.onclick = function() {
        if (buttons.style.display === "none") {
            buttons.style.display = "grid";
            showButtons.innerHTML = "Hide"
            log.style.bottom = "280px";
            log.scrollTop = log.scrollHeight;
        } else {
            buttons.style.display = "none";
            showButtons.innerHTML = "Show"
            log.style.bottom = "48px";
        }
    };

    bottomBar.onsubmit = function () {
        if (!conn) {
            return false;
        }
        if (!msg.value) {
            return false;
        }
        conn.send(msg.value);
        msg.value = "";
        autoscroll();
        return false;
    };

    xbtnClick = function (e) {
        if (!conn) {
            return false;
        }

        noSleep.enable();

        conn.send(e.text);
        autoscroll();
        return false;
    };

    if (window["WebSocket"]) {
        startWs();
    } else {
        var item = document.createElement("div");
        item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
        appendLog(item);
    }
};