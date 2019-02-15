let modal = null;
let modalContent = null;
let searchCount=0;
let valid = false;
// Get Modal instance
document.addEventListener('DOMContentLoaded', function() {
    let modalElement = document.querySelectorAll('.modal');
    var instances = M.Modal.init(modalElement);
    modalContent = document.querySelector('.modal-content');
    modal = instances[0];
});

// dealing with floating button
document.addEventListener('DOMContentLoaded', function() {
    var elems = document.querySelectorAll('.fixed-action-btn');
    var instances = M.FloatingActionButton.init(elems);
    elems[0].onclick = () => {
        document.querySelector('#delete-search').classList.add('hide');
        newSearch();
        modal.open();
    }
});
// Initialize materialize js components
M.AutoInit();

document.querySelector('#add-search').onclick = () => {
    addSeachInput();
};

document.querySelector('#remove-search').onclick = () => {
    removeSearchInput();
};

function getCookie(name){
    var pattern = RegExp(name + "=.[^;]*")
    matched = document.cookie.match(pattern)
    if(matched){
        var cookie = matched[0].split('=')
        return cookie[1]
    }
    return false
}

// newSearch pulls up the modal with the content for a new search
function newSearch() {
    searchCount = 0;
    modalContent.innerHTML = `<h4>Subreddit Search</h4>
        <div class="row" id="subSearchField">
            <form class="col s12">
                <div class="row">
                    <div class="input-field col s12">
                    <input id="subname" type="text" onchange="if(this.value.trim() !== '') validateSubName(this.value.trim().toLowerCase())">
                        <label for="subname">Subreddit</label>
                        <span class="helper-text" id="valid-message"></span>
                    </div>
                </div>
            </form>
        </div>`;
    document.querySelector('#confirm-search').onclick = () => {
        confirmSearch();
    };
}

// addSearchInput adds a text field below the sub name
function addSeachInput() {
    id = `search${searchCount}`;
    document.querySelector('#subSearchField').insertAdjacentHTML('beforeend',`
        <div class="input-field col s12" id="${id}">
            <input id=${id} type="text" class="search">
            <label for=${id}>Search</label>
        </div>`);
    searchCount++;
}

// removeSearchInput removes the last search field from the modal
function removeSearchInput() {
    // Get element and then remove it
    let removedInput = document.querySelector(`div#search${searchCount - 1}`);
    removedInput.remove();
    searchCount--;
}

// confirmSearch validates the search, sends it to the server, and adds it to the page
function confirmSearch() {
    // Check is sub is valid
    if (!valid) {
        M.toast({html: 'Please provide a valid subreddit name.'});
        return;
    }
    // Get subname and a list of all the searches
    const subname = document.getElementById('subname').value.trim().toLowerCase();
    let inputList = document.getElementsByClassName('search');
    let searchList=[];
    for (let i = 0; i < inputList.length; i++) {
        let searchValue = inputList[i].value.trim();
        if (searchValue !== "") {
            searchList.push(inputList[i].value.trim());
        }
    }
    // basic validation
    if (subname.length === 0) {
        M.toast({html: 'Please provide a subreddit name.'});
        return;
    }
    if (searchList.length === 0) {
        M.toast({html: 'Please provide at least one valid search term.'});
        return;
    }
    let obj = {
        subreddit: subname,
        searches: searchList,
    };
    // Add the contents provided to the page
    addSearchToPage(subname, searchList);
    csrfToken = getCookie("_csrf");
    var req = new XMLHttpRequest();
    postUrl = "/addSearch";
    req.open("POST", postUrl, true);
    req.setRequestHeader('X-CSRF-Token', csrfToken);
    req.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');
    req.send(JSON.stringify(obj));
    console.log(JSON.stringify(obj));
    modal.close();
}

// addSearchToPage formats the search data and adds it to the page
function addSearchToPage(subname, searchList) {
    // build output
    let output = `
        <li>
            <div class="collapsible-header">
                 ${subname}
                <i class="large material-icons right-align edit-icon">create</i>
            </div>
            <div class="collapsible-body">
            <span>
                <table>
                    <tbody>`;
    searchList.forEach((search) => {
        output += `
            <tr>
                <td>${search}</td>
            </tr>`
    });
    output += `
                </tbody>
            </table>
          </span>
        </div>
    </li>
    `;
    // Insert adjacent so that html does get rebuilt
    document.querySelector('.collapsible').insertAdjacentHTML('beforeend', output);
    // Add event listener to newly created edit icon
    const icons = document.querySelectorAll('.edit-icon');
    addEditEventListeners(subname, searchList, icons[icons.length - 1]);
    searchCount = 0;
}

// addEditEventListeners adds event listeners to all of the edit buttons
function addEditEventListeners(subname, searchList, icon) {
    searchCount = 0;
    // add event listen to edit button passed in
    icon.addEventListener('click', (ev) => {
        // Make it so that click on button does not count as click on parent
        ev.stopPropagation();
        // Build output
        let out = `
                <h4>Edit Search</h4>
                <div class="row" id="subSearchField">
                    <form class="col s12">
                        <div class="row">
                            <div class="input-field col s12">
                                <input disabled id="subname" type="text" value="${subname}">
                                <label for="subname" class="active">Subreddit</label>
                            </div>
                        </div>
                    </form>`;
        searchList.forEach((search) => {
            id = `search${searchCount}`;
            out += `
                <div class="input-field col s12" id="${id}">
                    <input id=${id} type="text" class="search" value="${search}">
                    <label class="active" for=${id}>Search</label>
                </div>`;
            searchCount++;
        });
        out += `</div>`;
        // Make sure that the delete button is visible
        document.querySelector('#delete-search').classList.remove('hide');
        // Make the delete button remove the
        document.querySelector('#delete-search').onclick = () => {
            if(confirm('Are you sure you want to delete this search?')) {
                let obj = {
                    subreddit: subname
                };
                csrfToken = getCookie("_csrf");
                var req = new XMLHttpRequest();
                postUrl = "/deleteSub";
                req.open("POST", postUrl, true);
                req.setRequestHeader('X-CSRF-Token', csrfToken);
                req.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');
                req.send(JSON.stringify(obj));
                icon.parentElement.parentElement.remove();
                modal.close();
            }
        };
        // Change logic for the confirm button
        document.querySelector('#confirm-search').onclick = () => {
            updateSearch(icon);
        };
        // Add html to modal and open it
        modalContent.innerHTML = out;
        modal.open();
    });
}

// updateSearch updates specific search on the page
function updateSearch(editButton) {
    // Get the div that actually holds the content
    let contentDiv = editButton.parentElement.nextSibling.nextSibling;
    if (contentDiv === undefined || contentDiv === null) {
        contentDiv = editButton.parentElement.nextSibling;
    }
    // Get subname, only for function call, and the list of searches
    const subname = document.getElementById('subname').value.trim().toLowerCase();
    let inputList = document.getElementsByClassName('search');
    let searchList=[];
    for (let i = 0; i < inputList.length; i++) {
        let searchValue = inputList[i].value.trim();
        if (searchValue !== "") {
            searchList.push(inputList[i].value.trim());
        }
    }
    // Simple validation on subname and searches
    if (subname.length === 0) {
        M.toast({html: 'Please provide a subreddit name.'});
        return;
    }
    if (searchList.length === 0) {
        M.toast({html: 'Please provide at least one valid search term.'});
        return;
    }

    // Build output for page
    let output = `
            <span>
                <table>
                    <tbody>`;
    searchList.forEach((search) => {
        output += `
            <tr>
                <td>${search}</td>
            </tr>`
    });
    output += `
                </tbody>
            </table>
          </span>
    `;
    // Put output on page
    contentDiv.innerHTML = output;
    // re-add eventlistener to update button to reflect changes
    addEditEventListeners(subname, searchList, editButton);
    let obj = {
        subreddit: subname,
        searches: searchList,
    };
    csrfToken = getCookie("_csrf");
    var req = new XMLHttpRequest();
    postUrl = "/addSearch";
    req.open("POST", postUrl, true);
    req.setRequestHeader('X-CSRF-Token', csrfToken);
    req.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');
    req.send(JSON.stringify(obj));
    console.log(JSON.stringify(obj));
    searchCount = 0;
    modal.close();
}

function validateSubName(subname) {
    const subJ = JSON.stringify({
        'subreddit': subname,
    });
    let request = new XMLHttpRequest();
    let csrfToken = getCookie("_csrf");
    request.open('POST', '/validateSubreddit', true);
    // Set JSON and CSRF headers
    request.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');
    request.setRequestHeader('X-CSRF-Token', csrfToken);
    request.onload = () => {
        if (request.status >= 200 && request.status < 400) {
            // Success!
            let data = JSON.parse(request.responseText);
            // Change text box color and label text/color based on validity
            if (data.valid) {
                document.getElementById('subname').classList.add('valid');
                const validMessage = document.getElementById('valid-message');
                validMessage.innerText = 'Valid';
                validMessage.style.color = '#4CAF50';
                valid = true;
                return valid;
            } else {
                document.getElementById('subname').classList.add('invalid');
                const validMessage = document.getElementById('valid-message');
                validMessage.innerText = 'Subreddit is invalid.';
                validMessage.style.color = '#F44336';
                valid = false;
                return valid;
            }
        }
    };
    request.onerror = () => {
        // There was a connection error of some sort
        console.log("There was an error of some type, please try again")
    };
    request.send(subJ);
}
