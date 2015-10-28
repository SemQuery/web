window.addEventListener("load", function() {
    if (typeof WS_ID === "undefined") return;
    var url = 'ws://localhost:3000/socket';
    var socket = new WebSocket(url);

    socket.onopen = function() {
        socket.send("" + WS_ID);
    };

    var results = document.getElementById("query-results");

    socket.onmessage = function(event) {
        var packet = JSON.parse(event.data);
        if (packet.action == "results") {
            populateCode(JSON.parse(packet.payload));
            return;
        } else if (packet.action == "files") {
            searchResCount(packet.payload.count);
            return;
        } else if (packet.action == "warning") {
            results.innerHTML = packet.payload.message;
            return;
        }
        var parts = event.data.split(",");

        var percent = parts[0], files = parts[1], lines = parts[2];

        results.innerHTML = "Indexed <b>" + lines + "</b> lines in <b>" + files + "</b> files (<b>" + percent + "</b>%)";

        document.getElementById("bar").setAttribute("style", "width: " + percent + "%");
        document.getElementById("progress-label").innerText = percent + "%";

        if (percent == "100") {
            var pBar = document.getElementById("progress-bar");
            var cssClasses = pBar.getAttribute("class");
            pBar.setAttribute("class", cssClasses + " success");
        }
    }
});

function populateCode(json) {
    console.log(json);
    var table = document.createElement("TABLE");
    table.setAttribute("class", "codeview");
    
    var res = document.getElementById("search-results");
    var divNode = document.createElement("DIV");
    var iconNode = document.createElement("I");
    iconNode.setAttribute("class", "file text outline icon");
    divNode.setAttribute("class", "filename");
    divNode.appendChild(iconNode);

    var fn = json["file"];
    var parts = fn.split("/");
    for (var i = 3; i < parts.length; i++) {
        var segment = document.createElement("B");
        segment.setAttribute("class", "pathsegment");
        segment.innerText = parts[i];
        if (i < parts.length - 1)
            segment.innerText += "/";

        divNode.appendChild(segment);
    }

    res.appendChild(divNode);

    var relStart = json["relative_start"];
    var relEnd = json["relative_end"];

    var count = 0;
    var total = Object.keys(json["lines"]);
    Object.keys(json["lines"]).forEach(function(k) {
        var line = json["lines"][k];

        var tr = document.createElement("TR");
        var td = document.createElement("TD");

        td.setAttribute("class", "linenum");
        td.setAttribute("data-line-number", k);
        tr.appendChild(td);
        var td = document.createElement("TD");
        td.setAttribute("class", "codeline");
        var pre = document.createElement("pre");

        var highlight = document.createElement("DIV");
        highlight.setAttribute("class", "highlight");
        if (count == 0) {
            // first line
            
            var normalData = line.substring(0, relStart);
            var normal = document.createTextNode(normalData);
            if (total == 1) {
                highlight.innerText = line.substring(relStart, relEnd);
            } else {
                highlight.innerText = line.substring(relStart);
            }
            pre.appendChild(normal);
            pre.appendChild(highlight);
            if (total == 1) {
                normalData = line.substring(relEnd);
                normal = document.createTextNode(normalData);
                pre.appendChild(normal);
            }
        } else if (count == total - 1 && total > 1) {
            // last 
            highlight.innerText = line.substring(0, relEnd);
            pre.append(highlight);
            var normalData = line.substring(relEnd);
            var normal = document.createTextNode(normalData);
            pre.appendChild(normal);
        } else {
            highlight.innerText = line;
            pre.appendChild(highlight);
        }
        // pre.innerText = line;
        td.appendChild(pre);
        tr.appendChild(td);

        table.appendChild(tr);

        count++;
    });

    document.getElementById("search-results").appendChild(table);
    var br = document.createElement("BR");
    document.getElementById("search-results").appendChild(br);
}

function searchResCount(c) {
    $('#search-res-count').text("Found results in " + c + " files");
}
