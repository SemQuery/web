var App = App || {pages: {}};

$(document).ready(function() {
    if (CURRENT_PAGE === undefined)
        return;

    var namespace = App.pages[CURRENT_PAGE];

    var hooks = namespace.hooks;
    var binds = namespace.binds;

    if (hooks) {
        for (var i = 0; i < hooks.length; i++)
            hooks[i]();
    }
    if (binds) {
        for (var i = 0; i < binds.length; i++) {
            var obj = binds[i];
            $(obj.selector).on(obj.action, obj.callback);
        }
    }
});

function bind(namespace, obj) {
    namespace.binds.push(obj);
}
