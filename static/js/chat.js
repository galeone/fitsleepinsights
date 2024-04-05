function appendMessage(message, isUser, marker) {
    var chatMessages = document.querySelector('.chat-messages');
    let messageElement;
    if (marker == "full" || marker == "begin") {
        messageElement = document.createElement('div');
        messageElement.innerHTML = marked.parse(message);
        chatMessages.appendChild(messageElement);
    } else if (marker == "content" || marker == "end") {
        messageElement = document.querySelector('.chat-messages > div:last-of-type');
        messageElement.innerHTML += message;
        if (marker == "end") {
            messageElement.innerHTML = marked.parse(messageElement.innerHTML);
        }
    }
    if (isUser) {
        messageElement.classList.add('chat-message-user');
    } else {
        messageElement.classList.add('chat-message-bot');
    }
    
    var chatButton = document.querySelector('.chat-input button');
    if (isUser) {
        // Scroll to the bottom
        chatMessages.scrollTop = chatMessages.scrollHeight;

        // Replace the input button with a loading spinner
        chatButton.setAttribute("disabled", "disabled");
        chatButton.setAttribute("class", "lds-dual-ring-24");
        chatButton.classList.remove("fa-regular", "fa-paper-plane");
        return;
    }

    // Remove the loading spinner
    chatButton.removeAttribute("disabled");
    chatButton.removeAttribute("class");
    chatButton.classList.add("fa-regular", "fa-paper-plane");

    // We suppose the chat list already scrolled to the bottom
    // so we scroll it down a bit to show the new message
    chatMessages.scrollTop += 100;
}

document.addEventListener("DOMContentLoaded", function() {
    var loc = window.location;
    var uri = 'ws:';

    if (loc.protocol === 'https:') {
        uri = 'wss:';
    }
    uri += '//' + loc.host;
    uri += "/chat/" + document.getElementById("ranges").getAttribute("data-ranges").replaceAll("-", "/");

    ws = new WebSocket(uri)

    var chatInput = document.querySelector('.chat-input input[type="text"]');
    var chatButton = document.querySelector('.chat-input button');

    ws.onopen = function() {
        document.querySelector('.chat-messages').innerHTML = '';
        appendMessage("Hi, I'm your AI assistant 🤖<br>I analyzed the data visualized in the dashboard and I'm ready to answer your questions.", false, "full")
        chatInput.removeAttribute("disabled");
        chatButton.removeAttribute("disabled");
    }

    ws.onmessage = function(evt) {
        data = JSON.parse(evt.data);
        appendMessage(data.message, false, data.marker);     
    }

    // Send new message
    chatInput.addEventListener('keyup', (event) => {
        if (event.key === 'Enter') {
            const message = event.target.value;
            if (message === '') {
                return;
            }
            ws.send(message);
            appendMessage(message, true, "full");
            event.target.value = '';
            chatInput.va
        }
    });
    chatButton.addEventListener('click', (event) => {
        chatInput.dispatchEvent(new KeyboardEvent('keyup', {
            'key': 'Enter'
        }));
    });

});