/*
	2023 Kasyanov Nikolay Alexeyevich (Unbewohnte)
*/


async function post(url, login, password, json) {
    return fetch(url, {
        method: "POST",
        headers: {
            "Authorization": "Basic " + btoa(login + ":" + password),
            "Content-Type": "application/json",
        },
        body: JSON.stringify(json)
    })
}

async function postNewTodo(login, password, newTodo) {
    return post("/api/todo/create", login, password, newTodo)
}

async function postNewGroup(login, password, newGroup) {
    return post("/api/group/create", login, password, newGroup)
}

async function postNewUser(newUser) {
    return post("/api/user/create", "", "", newUser)
}

async function get(url, login, password) {
    return fetch(url, {
        method: "GET",
        headers: {
            "Authorization": "Basic " + btoa(login + ":" + password),
            "Content-Type": "application/json",
        },
    })
}

async function getUser(login, password) {
    return get("/api/user/get", login, password);
}

async function getTodos(login, password) {
    return get("/api/todo/get", login, password);
}

async function getGroup(login, password) {
    return get("/api/group/get", login, password);
}

async function getAllGroups(login, password) {
    return get("/api/user/get", login, password);
}

async function del(url, login, password) {
    return fetch(url, {
        method: "DELETE",
        headers: {
            "Authorization": "Basic " + btoa(login + ":" + password),
            "Content-Type": "application/json",
        },
    })
}

async function deleteTodo(login, password, id) {
    return del("/api/todo/delete/"+id, login, password);
}

async function update(url, login, password, json) {
    return post(url, login, password, json);
}

async function updateTodo(login, password, id, updatedTodo) {
    return update("/api/todo/update/"+id, login, password, updatedTodo);
}

async function updateGroup(login, password, id, updatedGroup) {
    return update("/api/group/update/"+id, login, password, updateGroup);
}

async function updateUser(login, password, updatedUser) {
    return update("/api/group/update/"+login, login, password, updatedUser);
}