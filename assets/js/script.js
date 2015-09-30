window.onload = function() {
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

    $('#login, #register').click(function() {
        $('#form-usr').val($('#username').val());
        $('#form-pwd').val($('#password').val());

        $('#form').submit();
    }); 

    $('#glogin').click(function() {
        window.location = "/githubauth"
    });

}
                    

function animateBG(duration) {
    var freq  = 100; // 0.1 sec
    var ticks = duration * 1000 / freq;

    var children = document.getElementById("Pattern").children;

    var i = 0;
    var handle = setInterval(function() {
        var idx     = 1 + Math.floor(Math.random() * (children.length - 1));
        var opacity = 0.2 - (Math.random() / 4);
        children[idx].setAttribute("fill-opacity", opacity);

        i++;
        if (i > ticks) {
            clearInterval(handle);
        }
    }, freq);        
}
