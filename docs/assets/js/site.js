$(document).ready(function(){

        return $(".content h2").each(function(i, el) {
            var $el, icon, id;
            $el = $(el);
            id = $el.attr('id');
            icon = '<i class="fa fa-link"></i>';
            if (id) {
                return $el.append($("<a />").addClass("header-link").attr("href", "#" + id).html("#"));
            }
        });
});

$(document).ready(function() {
    $('pre code').each(function(i, block) {
        hljs.highlightBlock(block);
    });
});