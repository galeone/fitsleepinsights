function appendMessage(message, user) {
    const messageElement = document.createElement('div');
    if (user) {
        messageElement.classList.add('chat-message-user');
    } else {
        messageElement.classList.add('chat-message-bot');
    }
    messageElement.innerHTML = message;
    var chatMessages = document.querySelector('.chat-messages');

    chatMessages.appendChild(messageElement);

    var chatButton = document.querySelector('.chat-input button');
    if (user) {
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

    // We suppose the chatlist already scrolled to the bottom
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
    uri += location.pathname.replace("/dashboard", "/chat");

    ws = new WebSocket(uri)

    var chatInput = document.querySelector('.chat-input input[type="text"]');
    var chatButton = document.querySelector('.chat-input button');

    ws.onopen = function() {
        document.querySelector('.chat-messages').innerHTML = '';
        appendMessage("Hi! <br> I analyzed your data and I'm ready to answer your questions. <br> Ask me anything!", false)
        chatInput.removeAttribute("disabled");
        chatButton.removeAttribute("disabled");
    }

    ws.onmessage = function(evt) {
        appendMessage(evt.data)
    }

    // Send new message
    chatInput.addEventListener('keyup', (event) => {
        if (event.key === 'Enter') {
            const message = event.target.value;
            if (message === '') {
                return;
            }
            ws.send(message);
            appendMessage(message, true);
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