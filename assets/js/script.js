var App = App || {pages: {}};

$(document).ready(function() {
    var CURRENT_PAGE = "home";
    var namespace    = App.pages[CURRENT_PAGE];

    var hooks = namespace.hooks;
    var binds = namespace.binds;

    if (hooks) {
        for (var i = 0; i < hooks.length; i++)
            hooks[i]();
    }
    if (binds) {
        for (var i = 0; i < binds.length; i++) {
            var arr = binds[i];
            $(arr[0]).on(arr[1], arr[2]);
        }
    }
});

function bind(namespace, selector, action, callback) {
    namespace.binds.push([selector, action, callback]);
}
