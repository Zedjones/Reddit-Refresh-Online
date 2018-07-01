// Decide if I want to allow submission on enter keypress
// $("input[type='text']").keypress(function(event){
//     if(event.which === 13){
//         cardCreate();
//     }
// });

let formArray=[];

function cardCreate() {
    let arr = $('#subSearch').serializeArray();
    formArray.push(arr);
    console.log(JSON.stringify(formArray));
    $('#subAppend').append(
    `<div class="p-2">
        <div class='card' style='width: 18rem;'>
            <div class='card-header'>
                ${arr[0].value}
            </div>
            <ul class='list-group list-group-flush card-list'>
            </ul>
        </div>
    </div>`);
    for(let i = 1; i < arr.length; i++){
        if(arr[i].value === ""){}
        else {
            $('.card-list').append(`<li class="list-group-item">${arr[i].value}</li>`);
        }
    }
    $('.card-list').removeClass('card-list');
    $('#subSearch')[0].reset();
}

$('#clickClick').on('click', (event) => {
    cardCreate();
});