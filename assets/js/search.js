var App = App || {pages: {}};

(function() {
    var SEARCH_PAGE_ID        = "search";
    App.Search                = {hooks: [], binds: []};
    App.pages[SEARCH_PAGE_ID] = App.Search;

    function handleSocketUpdate(event) {
        var update = JSON.parse(event.data);
        switch (update.action) {
        case "warning":
            $("#mesage").text(update.payload.message);
            break;
        case "indexing":
            updateIndexingProgress(update.payload);
            break;
        }
    }

    function updateIndexingProgress(payload) {
        var lines   = payload.lines;
        var files   = payload.files;
        var percent = payload.percent;
        
        var queryResHTML = "Indexed <b>" + lines + "</b>" +
            " lines in <b>" + files + "</b> files" + 
            " (<b>" + percent + "</b>%)";
        $("query-results").html(queryResHTML);

        $("#bar").attr("style", "width:" + percent + "%");
        $("#progress-label").text(percent + "%");
        if (percent == "100") {
            $("$progress-bar").addClass("success");
        }
    }

    function initiateIndexing() {
        var INDEX_SOURCE_PATH = "/index_source";
        // TODO make URL configurable
        var WS_URL = "ws://localhost:3000/socket" + location.search;

        $.post(INDEX_SOURCE_PATH, {
            search: location.search
        }).done(function(data) {
            var res = JSON.parse(data);
            switch (res.action) {
            case "warning":
                $("#message").text(res.payload.message);
                break;
            case "success":
                var socket = new WebSocket(WS_URL);
                socket.onmessage = handleSocketUpdate;
                break;
            }
        });
    }
    bind(App.Search, {
        selector: '#index-source-button',
        action: 'click',
        callback: initiateIndexing
    });
})();
