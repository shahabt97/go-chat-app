const nickname = localStorage.getItem("nickname");

const socket = io(`/pv-chat?nickname=${nickname}`);

//Query DOM
const messageInput = document.getElementById("messageInput"),
    chatForm = document.getElementById("chatForm"),
    chatBox = document.getElementById("chat-box"),
    feedback = document.getElementById("feedback"),
    onlineUsers = document.getElementById("online-users-list"),
    chatContainer = document.getElementById("chatContainer"),
    pvChatForm = document.getElementById("pvChatForm"),
    pvMessageInput = document.getElementById("pvMessageInput"),
    modalTitle = document.getElementById("modalTitle"),
    pvChatMessage = document.getElementById("pvChatMessage");

// const nickname = localStorage.getItem("nickname");
let socketId;
// Emit Events
socket.emit("login", nickname);

// socket.on("hi",(data)=>{
//     console.log("dataaaa: ",data)
// })

chatForm.addEventListener("submit", (e) => {
    e.preventDefault();
    if (messageInput.value) {
        socket.emit("chat message", {
            message: messageInput.value,
            name: nickname,
        });
        messageInput.value = "";
    }
});

messageInput.addEventListener("keypress", () => {
    socket.emit("typing", { name: nickname });
});

pvChatForm.addEventListener("submit", (e) => {
    e.preventDefault();

    socket.emit("pvChat", {
        message: pvMessageInput.value,
        name: nickname,
        to: socketId,
        from: socket.id,
    });

    $("#pvChat").modal("hide");
    pvMessageInput.value = "";
});

socket.on("chat message", (data) => {
    console.log(`data has been recieved from ${socket.id}`)
    feedback.innerHTML = "";
    chatBox.innerHTML += `
                        <li class="alert alert-light">
                            <span
                                class="text-dark font-weight-normal"
                                style="font-size: 13pt"
                                >${data.name}</span
                            >
                            <span
                                class="
                                    text-muted
                                    font-italic font-weight-light
                                    m-2
                                "
                                style="font-size: 9pt"
                                >ساعت 12:00</span
                            >
                            <p
                                class="alert alert-info mt-2"
                                style="font-family: persian01"
                            >
                            ${data.message}
                            </p>
                        </li>`;
    chatContainer.scrollTop =
        chatContainer.scrollHeight - chatContainer.clientHeight;
});