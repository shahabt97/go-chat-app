async function init() {
  // Connect to Socket.IO server
  const socket = new WebSocket("ws://localhost:8080/ws");
  let id = Math.floor(Math.random() * 1000000);
  let username = Math.floor(Math.random() * 1000000).toString();

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
    const userId = id; // Replace with the actual username
    const timestamp = new Date(); // Get the current timestamp

    if (message !== "") {
      // Create an object with the message, username, and timestamp
      const messageData = {
        message,
        username,
        userId,
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

  const data =  JSON.parse(event.data.toString())
    console.log("data: ", data);
    if (data.eventName == "chat message") {
      receiveMessage(data.data);
    }
  };
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
