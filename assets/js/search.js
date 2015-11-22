var App = App || {pages: {}};

(function() {
    var SEARCH_PAGE_ID        = "search";
    App.Search                = {hooks: [], binds: []};
    App.pages[SEARCH_PAGE_ID] = App.Search;

    function handleSocketUpdate(event) {
        var update = JSON.parse(event.data);
        console.log(update);
        switch (update.action) {
        case "warning":
            $("#mesage").text(update.payload.message);
            break;
        case "indexing":
            updateIndexingProgress(update.payload);
            break;
        case "cloning":
            if (update.payload.status == "started") {
                $('.indexing-phase-queued').addClass('done');
                $('.indexing-phase-cloning').slideToggle();
            } else if (update.payload.status == "finished") {
                $('.indexing-phase-cloning').addClass('done');
                $('.indexing-phase-indexing').slideToggle();
            }
        }
    }

    function updateIndexingProgress(payload) {
        var lines   = parseInt(payload.lines).toLocaleString();
        var files   = parseInt(payload.files).toLocaleString();
        var percent = payload.percent;

        $('#progress-lines-counter').text(lines);
        $('#progress-files-counter').text(files);
        $('#progress-percent-counter').text(percent + "%");
        
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

        $('#index-source-button .loader').addClass('active');

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
                $('#progress-segment').slideToggle();
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
