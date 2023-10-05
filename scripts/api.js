/*
	2023 Kasyanov Nikolay Alexeyevich (Unbewohnte)
*/


async function postNewTodo(username, password, new_todo) {
    return fetch("/api/todo", {
        method: "POST",
        headers: {
            "EncryptedBase64": "false",
            "Auth": username  + "<-->" + password,
            "Content-Type": "application/json",
        },
        body: JSON.stringify(new_todo),
    });
}


async function getTodos(username, password) {
    return fetch("/api/todo", {
        method: "GET",
        headers: {
            "EncryptedBase64": "false",
            "Auth": username  + "<-->" + password
        },
    });
}


async function getTodoGroups(username, password) {
    return fetch("/api/group", {
        method: "GET",
        headers: {
            "EncryptedBase64": "false",
            "Auth": username  + "<-->" + password
        },
    });
}

async function deleteTodo(username, password, id) {
    return fetch("/api/todo/"+String(id), {
        method: "DELETE",
        headers: {
            "EnctyptedBase64": "false",
            "Auth": username + "<-->" + password,
        },
    });
}

async function updateTodo(username, password, id, updatedTodo) {
    return fetch("/api/todo/"+String(id), {
        method: "POST",
        headers: {
            "EncryptedBase64": "false",
            "Auth": username + "<-->" + password,
        },
        body: JSON.stringify(updatedTodo),
    });
}

async function getUser(username, password) {
    return fetch("/api/user", {
        method: "GET",
        headers: {
            "EncryptedBase64": "false",
            "Auth": username  + "<-->" + password
        },
    });
}