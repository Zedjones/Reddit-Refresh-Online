//Click on X to delete todo
$(".search").on("click", "span", function(event){
    $(this).parent().fadeOut(500, function() {
        $(this).remove();
    });
    event.stopPropagation();
});

//Add new Todo
$("input[type='text']").keypress(function(event){
    if(event.which === 13){
        var todoText = $(this).val();
        $(this).val("");
        $(".search").append("<li><span><i class='fa fa-trash'></i></span> " + todoText + "</li>");
    }
});

$(".done").click(function(){
    $("input[type='text']").fadeOut();
});

$(".edit").click(function(){
    $("input[type='text']").fadeIn();
});