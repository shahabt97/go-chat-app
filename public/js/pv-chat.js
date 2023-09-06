async function init() {
  const path = window.location.pathname;
  const hostUser = path.split("/")[3];

  // Connect to Socket.IO server
  function connection() {
    const socket = new WebSocket(`ws://${hostAndPort}/ws/pv?host=${hostUser}`);

    const inputBoxForm = document.getElementById("inputBoxForm");
    const submitButton = document.getElementById("submitButton");

    inputBoxForm.addEventListener("submit", (e) => {
      e.preventDefault();
      sendMessage();
    });
    
    submitButton.addEventListener("click", (e) => {
      e.preventDefault();
      sendMessage();
    });

    // Function to send a new message
    function sendMessage() {
      const messageInput = document.getElementById("messageInput");
      const message = messageInput.value.trim();
      const timestamp = new Date(); // Get the current timestamp

      if (message !== "") {
        // Create an object with the message, username, and timestamp
        const messageData = {
          message,
          receiver: hostUser,
          timestamp,
        };

        // Emit the 'newMessage' event to the server with the message data
        socket.send(
          JSON.stringify({
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

      // Create HTML structure for the message
      messageElement.innerHTML = `<div class="message">
              <span class="sender">${messageData.sender}:</span>
              <span class="timestamp">${formatDate(
                messageData.timestamp
              )}</span>
              <p>${messageData.message}</p>`;

      chatBox.appendChild(messageElement);
      chatBox.scrollTop = chatBox.scrollHeight;
    }

    // Listen for 'newMessage' event from the server
    socket.onmessage = async (event) => {
      const data = JSON.parse(event.data.toString());
      console.log("data: ", data);
      if (data.eventName == "chat message") {
        receiveMessage(data.data);
      }
      if (data.eventName == "all messages") {
        await getMessages(data.data);
      }
    };
  }

  async function getMessages(messages) {
    const chatBox = document.getElementById("chatBox");

    for (let i = 0; i < messages.length; i++) {
      const messageElement = document.createElement("div");
      messageElement.innerHTML = `<div class="message">
      <span class="sender">${messages[i].sender}:</span>
      <span class="timestamp">${formatDate(messages[i].createdAt)}</span>
      <p>${messages[i].message}</p>`;
      chatBox.appendChild(messageElement);
    }

    chatBox.scrollTop = chatBox.scrollHeight;
  }
  connection();
}
function formatDate(dateTimeStr) {
  // Format in ISO format
  const date = new Date(dateTimeStr);
  const options = { hour: "numeric", minute: "numeric", weekday: "long" };
  const formattedTime = date.toLocaleString("en-US", options);
  return formattedTime;
}
