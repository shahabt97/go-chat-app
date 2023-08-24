async function init() {
  // find user ID
  const data = await axios.get("http://localhost:8080/user/get-user-id");
  console.log(data);

  const messages = await axios.get(
    "http://localhost:8080/chat/get-messages?&status=public"
  );
  await getMessages(messages.data);
  console.log("messages: ", messages);

  // Connect to Socket.IO server
  let id = data.data.id;
  let username = data.data.username;
  const socket = new WebSocket(
    `ws://localhost:8080/ws?username=${username}&id=${id}`
  );

  const inputBoxForm = document.getElementById("inputBoxForm");
  const submitButton = document.getElementById("submitButton");
  const onlineUsers = document.getElementById("onlineUsers");

  inputBoxForm.addEventListener("submit", (e) => {
    e.preventDefault();
    sendMessage();
  });
  submitButton.addEventListener("click", (e) => {
    e.preventDefault();
    sendMessage();
  });

  // socket.on("online", (users) => {
  //   onlineUsers.innerHTML = "";
  //   for (const socketId in users) {
  //     if (users[socketId] !== data.data.username)
  //       onlineUsers.innerHTML += `
  //         <li><span></span><a href="/pv-chat/${users[socketId]}" target="_blank">${users[socketId]}</a></li>
  //       `;
  //   }
  // });

  // Function to send a new message
  function sendMessage() {
    const messageInput = document.getElementById("messageInput");
    const message = messageInput.value.trim();
    // const userId = id; // Replace with the actual username
    const timestamp = new Date(); // Get the current timestamp

    if (message !== "") {
      // Create an object with the message, username, and timestamp
      const messageData = {
        message,
        username,
        // userId,
        timestamp,
      };

      // Emit the 'newMessage' event to the server with the message data
      socket.send(
        JSON.stringify({
          id,
          eventName: "chat message",
          data: messageData,
        })
      );
      messageInput.value = ""; // Clear the input field
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
          <span class="timestamp">${formatDate(messageData.timestamp)}</span>
          <p>${messageData.message}</p>`;

    chatBox.appendChild(messageElement);
    chatBox.scrollTop = chatBox.scrollHeight;
  }

  // Listen for 'newMessage' event from the server
  socket.onmessage = (event) => {
    const data = JSON.parse(event.data.toString());
    console.log("data: ", data);
    if (data.eventName == "chat message") {
      receiveMessage(data.data);
    }
  };

  async function getMessages(messages) {
    const chatBox = document.getElementById("chatBox");
    //   messageElement.classList.add("message");
    // console.log(messages);
    for (let i = 0; i < messages.length; i++) {
      const messageElement = document.createElement("div");
      messageElement.innerHTML = `<div class="message">
  <span class="sender">${messages[i].sender.username}:</span>
  <span class="timestamp">${formatDate(messages[i].createdAt)}</span>
  <p>${messages[i].message}</p>`;

      chatBox.appendChild(messageElement);
    }
    chatBox.scrollTop = chatBox.scrollHeight;

    // Create HTML structure for the message
  }
}

function formatDate(dateTimeStr) {
  // const date = new Date();
console.log("dateTimeStr: ",dateTimeStr)
  // Format in ISO format
  const date = new Date(dateTimeStr);
  console.log("date: ",date);
  const options = { hour: "numeric", minute: "numeric", weekday: "long" };
  const formattedTime = date.toLocaleString("en-US", options);
  // console.log(formattedTime);

  // const [dayOfWeek, time] = formattedTime.split(",");
  // const [hour, minute] = time.split(":");

  return formattedTime;
}
