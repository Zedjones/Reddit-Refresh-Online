// Decide if I want to allow submission on enter keypress
// $("input[type='text']").keypress(function(event){
//     if(event.which === 13){
//         cardCreate();
//     }
// });
 
let searchObj = {subs: []};
let masterArr = [];
 
function cardCreate(arr) {
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

function objCreate(arr){
    let sub = arr[0].value
    let searches = []
    arr.slice(1).forEach((search) => {
        if (search.value != ""){
            searches.push(search.value);
        }
    });
    let subObj = {sub: sub, searches: searches}
    searchObj.subs.push(subObj);
}
 
$('#clickClick').on('click', (event) => {
    let arr = $('#subSearch').serializeArray();
    masterArr.push(arr);
    cardCreate(arr);
});

$('#bigSubmit').on('click', (event) => {
    masterArr.forEach((arr) => {
        objCreate(arr);
    });
});

$('#editBtn').on('click', (event) => {
    masterArr.forEach((search) => {
        console.log(search[0].value);
    });
    $('.form-group input')[0].val(masterArr[0][0].value);
});