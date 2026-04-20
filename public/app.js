const API_BASE = "";

function setTokens(accessToken, refreshToken) {
    localStorage.setItem("access_token", accessToken);
    localStorage.setItem("refresh_token", refreshToken);
}

function getAccessToken() {
    return localStorage.getItem("access_token");
}

async function logout() {
    const token = getAccessToken();
    if (token) {
        await fetch(API_BASE + "/auth/logout", {
            method: "POST",
            headers: { "Authorization": "Bearer " + token }
        }).catch(() => {});
    }
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    window.location.href = "/login.html";
}

// LOGIN
async function loginUser(event) {
    event.preventDefault();

    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;

    const res = await fetch(API_BASE + "/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password })
    });

    const body = await res.json();

    if (res.ok) {
        setTokens(body.data.access_token, body.data.refresh_token);
        window.location.href = "/me.html";
    } else {
        document.getElementById("message").innerText = body.error?.message || "Login failed";
    }
}

// REGISTER
async function registerUser(event) {
    event.preventDefault();

    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;

    const res = await fetch(API_BASE + "/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password })
    });

    const body = await res.json();

    if (res.ok) {
        window.location.href = "/login.html";
    } else {
        document.getElementById("message").innerText = body.error?.message || "Registration failed";
    }
}

// GET ME
async function loadMe() {
    const token = getAccessToken();

    if (!token) {
        window.location.href = "/login.html";
        return;
    }

    const res = await fetch(API_BASE + "/auth/me", {
        headers: { "Authorization": "Bearer " + token }
    });

    if (!res.ok) {
        logout();
        return;
    }

    const body = await res.json();
    const user = body.data;

    document.getElementById("user-info").innerText =
        `Email:   ${user.email}\nUser ID: ${user.user_id}`;
}
