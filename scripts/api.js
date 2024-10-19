/*
	2024 Kasyanov Nikolay Alexeyevich (Unbewohnte)
*/

async function post(url, json) {
    return fetch(url, {
        method: "POST",
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(json)
    })
}


async function postNewTodo(newTodo) {
    return post("/api/todo/create", newTodo)
}

async function postNewGroup(newGroup) {
    return post("/api/group/create", newGroup)
}

async function postNewUser(newUser) {
    return post("/api/user/create", newUser)
}

async function get(url) {
    return fetch(url, {
        method: "GET",
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
        },
    })
}

async function getUser() {
    return get("/api/user/get");
}

async function getTodos() {
    return get("/api/todo/get");
}

async function getGroup() {
    return get("/api/group/get");
}

async function getAllGroups() {
    return get("/api/user/get");
}

async function del(url) {
    return fetch(url, {
        method: "DELETE",
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
        },
    })
}

async function deleteTodo(id) {
    return del("/api/todo/delete/"+id);
}

async function update(url, json) {
    return post(url, json);
}

async function updateTodo(id, updatedTodo) {
    return update("/api/todo/update/"+id, updatedTodo);
}

async function updateGroup(id, updatedGroup) {
    return update("/api/group/update/"+id, updatedGroup);
}

async function updateUser(updatedUser) {
    return update("/api/user/update", updatedUser);
}