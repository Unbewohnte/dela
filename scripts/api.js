/*
	Copyright (c) 2023 Kasyanov Nikolay Alexeyevich (Unbewohnte)
*/


async function post_new_todo(username, password, new_todo) {
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


async function get_todos(username, password) {
    return fetch("/api/todo", {
        method: "GET",
        headers: {
            "EncryptedBase64": "false",
            "Auth": username  + "<-->" + password
        },
    });
}


async function get_todo_groups(username, password) {
    return fetch("/api/group", {
        method: "GET",
        headers: {
            "EncryptedBase64": "false",
            "Auth": username  + "<-->" + password
        },
    });
}

async function delete_todo(username, password, id) {
    return fetch("/api/todo/"+String(id), {
        method: "DELETE",
        headers: {
            "EnctyptedBase64": "false",
            "Auth": username + "<-->" + password,
        },
    });
}

async function get_user(username, password) {
    return fetch("/api/user", {
        method: "GET",
        headers: {
            "EncryptedBase64": "false",
            "Auth": username  + "<-->" + password
        },
    });
}