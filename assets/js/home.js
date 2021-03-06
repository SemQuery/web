var App = App || {pages: {}};

(function() {
    var HOME_PAGE_ID        = "home";
    App.Home                = {hooks: [], binds: []};
    App.pages[HOME_PAGE_ID] = App.Home;

    // Handles search bar transition
    function searchTransition() {
        var expanded = false;
        var input = $('#repo-input');
        input.keydown(function(event) {
            if (event.keyCode == 13 && !expanded) {
                var val = input[0].value;
                if (val.split("/").length != 2) {
                    event.preventDefault();
                    return;
                }
                console.log("here");
                var ele = $('#repo-search .github.icon');
                ele.animate({
                    top: "-=50px"
                });
                var ele2 = $('#repo-search .code.icon');
                ele2.animate({
                    top: "-=50px"
                });
                $('#repo-form-value').val(input[0].value);
                input[0].value = "";
                input.attr('placeholder', 'enter semantic query');
                expanded = true;

                input.attr('style', 'font-family: Consolas, monospace !important; font-size: 25px;');
                event.preventDefault();

                animateBG(0.7);
            }
            $('#query-form-value').val(input[0].value);
        });    
    }

    App.Home.hooks.push(searchTransition);

    // scrolls to "Learn More" section
    function learnMore(e) {
        if (e)
            e.preventDefault();

        $('body').animate({
            scrollTop: $('.about-segment').offset().top
        }, 500);
    }

    App.Home.learnMore = learnMore;
    bind(App.Home, {
        selector: '#learn-more',
        action: 'click',
        callback: learnMore
    });

    var Sources = {
        GITHUB: "github-source-input",
        LINK: "link-source-input"
    }
    var currentSourceSelection = Sources.GITHUB;
    function selectSource(src) {
        if (src != currentSourceSelection) {
            currentSourceSelection = src;
            disableHiddenInputs();
            Object.keys(Sources).forEach(function (k) {
                var inputId = Sources[k];
                if (inputId != src) {
                    $(document.getElementById(inputId)).attr('style', 'display: none');
                }
            });
            $(document.getElementById(src)).attr('style', '');
        }
    }
    function selectGithubSource(e) {
        if (e)
            e.preventDefault();

        selectSource(Sources.GITHUB);
    }
    function selectLinkSource(e) {
        if (e)
            e.preventDefault();

        selectSource(Sources.LINK);
    }

    bind(App.Home, {
        selector: '#source-picker-github',
        action: 'click',
        callback: selectGithubSource
    });
    bind(App.Home, {
        selector: '#source-picker-link',
        action: 'click',
        callback: selectLinkSource
    });

    function disableHiddenInputs() {
        $("#" + currentSourceSelection + " input").removeAttr('disabled');
        Object.keys(Sources).forEach(function (k) {
            var id = Sources[k];
            if (id != currentSourceSelection) {
                $("#" + id + " input").attr('disabled', 'disabled');
            }
        });
    }
    bind(App.Home, {
        selector: 'form#source-form',
        action: 'submit',
        callback: disableHiddenInputs
    });
})();
