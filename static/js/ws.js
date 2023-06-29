let socket = null;

function connect() {
    socket = new WebSocket('ws://localhost:8090/ws');

    socket.onopen = function(e) {
        console.log("Соединение установлено");
    };

    socket.onmessage = function(event) {
        let message = JSON.parse(event.data);
        let now = new Date();
        let hours = now.getHours().toString().padStart(2, '0');
        let minutes = now.getMinutes().toString().padStart(2, '0');
        let seconds = now.getSeconds().toString().padStart(2, '0');
        let timeString = hours + ":" + minutes + ":" + seconds;

        const consoleDiv = document.querySelector('.console');
        if(message.name=="[GLOB]"){
            consoleDiv.innerHTML = consoleDiv.innerHTML + `<p class="prompt"><br></p>`
        }
        consoleDiv.innerHTML = consoleDiv.innerHTML + `<p class="prompt">
                                        <span class="grey">${timeString}</span>
                                        <span class=" ${message.color}">${message.name}</span>
                                        <span>⇒ </span> ${message.text}</p>`;

        consoleDiv.scrollTop = consoleDiv.scrollHeight;
    };

    socket.onerror = function(error) {
        console.log(`Ошибка: ${error.message}`);
    };

    socket.onclose = function(event) {
        if (event.wasClean) {
            console.log(`Соединение закрыто чисто, код=${event.code}`);
        } else {
            console.log('Обрыв соединения');
        }

        // Переподключение
        console.log('Пытаемся переподключиться...');
        setTimeout(connect, 5000); // попытка переподключения через 5 секунд
    };
}

// Инициируем подключение
connect();