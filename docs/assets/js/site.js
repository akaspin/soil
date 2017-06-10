$(document).ready(function(){
    // $('.content').on('mouseenter', 'h1[id], h2[id], h3[id], h4[id], h5[id], h6[id]', function(e) {
    //     $(e.target).append($('<a />').addClass('header-anchor').attr('href', '#' + e.target.id).html('<i class="fa fa-link " aria-hidden="true"></i>'));
    // });
    //
    // $('.content').on('mouseleave', 'h1[id], h2[id], h3[id], h4[id], h5[id], h6[id]', function(e) {
    //     $(e.target).parent().find('.header-anchor').remove();
    // });

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
    console.info(hljs.getLanguage('javascript'));
    hljs.registerLanguage('hcl', function (hljs) {
        lang = jQuery.extend({}, hljs.getLanguage('javascript'));
        return lang;
    });
    $('pre code').each(function(i, block) {
        hljs.highlightBlock(block);
    });
});