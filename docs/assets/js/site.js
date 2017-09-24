$(document).ready(function(){
    return $(".content h2").each(function(i, el) {
        var $el, icon, id;
        $el = $(el);
        id = $el.attr('id');
        if (id) {
            return $el.append($("<a />").addClass("header-link").attr("href", "#" + id).html("#"));
        }
    });
});

$(document).ready(function(){
    return $(".content dl>dt>code:first-child").each(function(i, el) {
        var $el, id, prev;
        $el = $(el);
        prev = $el.closest("dl").prevAll("h2, h1");
        if (prev.length > 0) {
            id = prev[0].id;
        }
        if (id) {
            return $el.wrap($("<a />").addClass("term-link").attr("href", "#term-" + id + "-" +  $el.text().replace(/[^a-z0-9+]+/gi, '-')));
        } else {
            return $el.wrap($("<a />").addClass("term-link").attr("href", "#term-" +  $el.text().replace(/[^a-z0-9+]+/gi, '-')));
        }
    });
});

$(document).ready(function() {
    $('pre code').each(function(i, block) {
        hljs.highlightBlock(block);
    });
});