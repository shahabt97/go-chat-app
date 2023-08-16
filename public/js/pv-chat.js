async function init() {
  const path = window.location.pathname;
  const hostUser = path.split("/")[2];
  const data = await axios.get(`http://${hostAndPort}/get-user-id`);

  const messages = await axios.get(
    `http://${hostAndPort}/get-messages?hostUser=${hostUser}&username=${data.data.username}&status=pv`
  );
  console.log(messages);
  await getMessages(messages.data);

  // Connect to Socket.IO server
  const socket = io(
    `/pv-chat?hostUser=${hostUser}&username=${data.data.username}`
  );

  const inputBoxForm = document.getElementById("inputBoxForm");
  const submitButton = document.getElementById("submitButton");
  const onlineUsers = document.getElementById("onlineUsers");

  inputBoxForm.addEventListener("submit", (e) => {
    e.preventDefault();
    console.log("hiiiiiirrrrr")
    sendMessage();
  });
  submitButton.addEventListener("click", (e) => {
    e.preventDefault();
    sendMessage();
  });

  socket.on("online", (users) => {
    onlineUsers.innerHTML = "";

    for (const socketId in users) {
      if (users[socketId] !== data.data.username)
        onlineUsers.innerHTML += `
              <li><span></span><a href="/pv-chat/${users[socketId]}" target="_blank">${users[socketId]}</a></li>
            `;
    }
  });

  // Function to send a new message
  function sendMessage() {
    const messageInput = document.getElementById("messageInput");
    const message = messageInput.value.trim();
    const username = data.data.username;
    const userId = data.data.id; // Replace with the actual username
    const timestamp = new Date(); // Get the current timestamp

    if (message !== "") {
      // Create an object with the message, username, and timestamp
      const messageData = {
        message,
        username,
        userId,
        timestamp,
      };
      console.log(messageData);
      // Emit the 'newMessage' event to the server with the message data
      if (data.data.id) {
        socket.emit("chat message", messageData);
        messageInput.value = ""; // Clear the input field
      }
    }
  }

  // Function to handle incoming messages
  function receiveMessage(messageData) {
    const chatBox = document.getElementById("chatBox");
    const messageElement = document.createElement("div");
    //   messageElement.classList.add("message");

    // Create HTML structure for the message
    messageElement.innerHTML = `<div class="message">
              <span class="sender">${messageData.username}:</span>
              <span class="timestamp">${formatDate(
                messageData.timestamp
              )}</span>
              <p>${messageData.message}</p>`;

    chatBox.appendChild(messageElement);
    chatBox.scrollTop = chatBox.scrollHeight;
  }

  // Listen for 'newMessage' event from the server
  socket.on("chat message", (messageData) => {
    receiveMessage(messageData);
  });

  async function getMessages(messages) {
    const chatBox = document.getElementById("chatBox");
    //   messageElement.classList.add("message");
    // console.log(messages);
    for (let i = 0; i < messages.length; i++) {
      const messageElement = document.createElement("div");
      messageElement.innerHTML = `<div class="message">
      <span class="sender">${messages[i].sender.username}:</span>
      <span class="timestamp">${formatDate(messages[i].createdAT)}</span>
      <p>${messages[i].message}</p>`;

      chatBox.appendChild(messageElement);
    }
    chatBox.scrollTop = chatBox.scrollHeight;

    // Create HTML structure for the message
  }
}
function formatDate(dateTimeStr) {
  // const date = new Date();

  // Format in ISO format
  const date = new Date(dateTimeStr);
  // console.log(date);
  const options = { hour: "numeric", minute: "numeric", weekday: "long" };
  const formattedTime = date.toLocaleString("en-US", options);
  // console.log(formattedTime);

  // const [dayOfWeek, time] = formattedTime.split(",");
  // const [hour, minute] = time.split(":");

  return formattedTime;
}
