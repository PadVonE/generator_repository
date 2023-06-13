
// Создаем новое WebSocket-соединение
const socket = new WebSocket('ws://localhost:8090/ws');

// Устанавливаем обработчики событий
socket.onopen = function(e) {
    console.log("Соединение установлено");
};

socket.onmessage = function(event) {

    let message = JSON.parse(event.data);

    // Создаем объект Date и форматируем его
    let now = new Date();
    let hours = now.getHours().toString().padStart(2, '0');
    let minutes = now.getMinutes().toString().padStart(2, '0');
    let seconds = now.getSeconds().toString().padStart(2, '0');
    let timeString = hours + ":" + minutes + ":" + seconds;




    // Находим div.console и добавляем в него сообщение
    const consoleDiv = document.querySelector('.console');

    if(message.name=="[GLOB]"){
        consoleDiv.innerHTML = consoleDiv.innerHTML + `<p class="prompt"><br></p>`
    }

    consoleDiv.innerHTML = consoleDiv.innerHTML + `<p class="prompt">
                                    <span class="grey">${timeString}</span>
                                    <span class=" ${message.color}">${message.name}</span>
                                    <span>⇒ </span> ${message.text}</p>`;

    // Прокручиваем вниз
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
};