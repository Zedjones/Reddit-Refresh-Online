// Decide if I want to allow submission on enter keypress
// $("input[type='text']").keypress(function(event){
//     if(event.which === 13){
//         cardCreate();
//     }
// });

// Necessary objects and arrays
let searchObj = {subs: []};
let masterArr = [];
let checked = [];
let activeDevics = [];
let activeSub = '';

// Function that takes info from form and creates a card
// that is added to the page
function cardCreate(arr) {
    $('#subAppend').append(
    `<div class="p-2 access">
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

// removes all the cards from the screen
function removeCards() {
    let elements = document.querySelectorAll('.access');
    elements.forEach((element) => {
        element.parentNode.removeChild(element);
    });
}

// Runs when user clicks submit button
// Takes all the search info and creates one big object
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

// checks to see if subname has been added to panel already
function editPanelCheck(subName) {
    let check = false
    checked.forEach((sub) => {
        if(sub === subName){
            check = true;
        }
    });
    return check;
}

// removes active class from edit list items
function unactivateSubs() {
    let arr = document.querySelectorAll('.rand');
    arr.forEach((subName) => {
        subName.classList.remove('active');
    });
}

function updateSubList() {
    let arr = document.querySelectorAll('.rand');
    arr.forEach((subName) => {
        subName.parentNode.removeChild(subName);
    });
}

// Event listener that saves form info and calls cardCreate func
$('#clickClick').on('click', (event) => {
    let arr = $('#subSearch').serializeArray();
    masterArr.push(arr);
    removeCards();
    masterArr.forEach((arr) => {
        cardCreate(arr);
    });
});

// Calls objCreate for each sub search
$('#bigSubmit').on('click', (event) => {
    masterArr.forEach((arr) => {
        objCreate(arr);
    });
});

// Allows user to pick a sub search to edit
$('#editBtn').on('click', (event) => {
    document.querySelector('#deleteSub').disabled = true;
    document.querySelector('#editSub').disabled = true;
    removeCards();
    masterArr.forEach((search) => {
        if(editPanelCheck(search[0].value)) {}
        else {
            $('#editList').append(
                `<a href="#" class="list-group-item list-group-item-action rand">
                ${search[0].value}
                </a>`
            );
            checked.push(search[0].value);
        }
    });
    unactivateSubs();
    let arr = document.querySelectorAll('.rand');
    arr.forEach((subName) => {
        subName.onclick = () => {
            document.querySelector('#deleteSub').disabled = false;
            document.querySelector('#editSub').disabled = false;
            arr.forEach((sub) => {
                sub.classList.remove('active');
            });
            subName.classList.add('active');
            activeSub = subName.textContent.trim();
        };
    });
});

// fills out sub search modal with correct info from original serach
$('#editSub').on('click', (event) => {
    console.log(masterArr);
    let arr = [];
    let index = 0;
    masterArr.forEach((subSearch) => {
        if(subSearch[0].value == activeSub) {
            arr.push(subSearch);
            masterArr.splice(index, 1);
        }
        index++;
    });
    let inputs = document.querySelectorAll('.form-group input');
    for(let i = 0; i < inputs.length; i++) {
        inputs[i].value = arr[0][i].value;
    }
});

// deletes a specific sub from the screen
$('#deleteSub').on('click', (event) => {
    let index = 0;
    masterArr.forEach((subSearch) => {
        if(subSearch[0].value == activeSub) {
            masterArr.splice(index, 1);
            console.log(masterArr);
        }
        index++;
    });
    updateSubList();
    masterArr.forEach((arr) => {
        cardCreate(arr);
    });

});

// clears out values in sub search modal 
$('#searchBtn').on('click', (event) => {
    $('#subSearch')[0].reset();
});

document.querySelectorAll('.pushBtn').forEach((button) => {
    button.onclick = () => {
        button.classList.toggle('btn-primary');
        button.classList.toggle('btn-dark');
    }
});






