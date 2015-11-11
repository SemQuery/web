$("#indexaction").submit(function(e) {
    e.preventDefault();

    $.post("index_source", { search: location.search });

    var url = 'ws://localhost:3000/socket?' + location.search;
    var socket = new WebSocket(url);

    var results = $("#query-results");

    socket.onmessage = function(event) {
        console.log(event.data);
        var packet = JSON.parse(event.data);
        switch (packet.action) {
            case "results":
                $('#search-res-count').text("Found results in " + packet.payload.files + " files");
                for (var i = 0; i != packet.payload.found.length; i++) {
                    populateCode(packet.payload.found[i]);
                }
                break;
            case "warning":
                results.text(packet.payload.message);
                break;
            case "indexing":
                results.html("Indexed <b>" + packet.payload.lines + "</b> lines in <b>" + packet.payload.files + "</b> files (<b>" + packet.payload.percent + "</b>%)"); 
                $("#bar").attr("style", "width: " + packet.payload.percent + "%");
                $("#progress-label").text(packet.payload.percent + "%");
                if (packet.payload.message == "100") {
                    var pBar = $("#progress-bar");
                    var cssClasses = pBar.attr("class");
                    pBar.attr("class", cssClasses + " success");
                }
                break;
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
