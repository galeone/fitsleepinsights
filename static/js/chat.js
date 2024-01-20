function appendMessage(user, message) {
  const messageElement = document.createElement('div');
  if (user) {
    messageElement.classList.add('chat-message-user');
  } else {
    messageElement.classList.add('chat-message');
  }
  messageElement.innerHTML = message;
  document.querySelector('.chat-messages').appendChild(messageElement);
  // Scroll to the bottom of the chat messages
  const chatMessages = document.querySelector('.chat-messages');
  chatMessages.scrollTop = chatMessages.scrollHeight;
}

document.addEventListener("DOMContentLoaded", function() {
  var loc = window.location;
  var uri = 'ws:';

  if (loc.protocol === 'https:') {
      uri = 'wss:';
  }
  uri += '//' + loc.host;
  uri += location.pathname.replace("/dashboard", "/chat");

  ws = new WebSocket(uri)

  ws.onopen = function () {
    console.log('Connected')
  }

  ws.onmessage = function (evt) {
    appendMessage("Server", evt.data)
  }

  // Send new message
  document.querySelector('input[type="text"]').addEventListener('keyup', (event) => {
  if (event.key === 'Enter') {
    const message = event.target.value;
    ws.send(message);
    appendMessage('You', message);
    event.target.value = '';
  }
});

});