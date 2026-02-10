package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
    "sync"
)

var (
	MessageLog = make([]string, 0, 200)
	logMu      sync.RWMutex
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", GetMessage)
	mux.HandleFunc("/message", SendMessage)
	mux.HandleFunc("/poll", PollMessage)

	fmt.Println("Сервер запущен на localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}

func GetMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="utf-8">
	<title>Чат</title>
	<style>
		body { font-family: Arial, sans-serif; padding: 15px; }
		#chatForm { margin-bottom: 15px; }
		#username, #text { margin-right: 10px; padding: 6px; }
		button { padding: 6px 12px; }
		#chatHistory {
			max-height: 45vh;
			overflow-y: auto;
			border: 1px solid #ccc;
			padding: 10px;
			background-color: #f9f9f9;
			white-space: pre-wrap;
			font-family: monospace;
		}
	</style>
</head>
<body>
	<div id="chatForm">
		<input type="text" id="username" placeholder="Ваше имя" />
		<input type="text" id="text" placeholder="Сообщение" />
		<button onclick="send()">Отправить</button>
	</div>
	<div id="chatHistory"></div>
	<script>
		const history = document.getElementById("chatHistory");
		let prevCount = 0;

		function send() {
			const user = encodeURIComponent(
				document.getElementById("username").value.trim() || "anon"
			);
			const text = encodeURIComponent(
				document.getElementById("text").value.trim()
			);
			if (!text) return;

			fetch("/message?user=" + user + "&text=" + text, { method: "GET" })
				.then(r => r.text())
				.then(t => {
					console.log(t);
					document.getElementById("text").value = "";
				});
		}

		function loadMessages() {
			fetch("/poll")
				.then(r => r.json())
				.then(data => {
					if (!data || data.count === prevCount) return;
					
					prevCount = data.count;
					history.textContent = data.messages.join("\n");
					history.scrollTop = history.scrollHeight;
				});
		}

		setInterval(loadMessages, 300);
		loadMessages();
	</script>
</body>
</html>`)
}

func SendMessage(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	msg := r.URL.Query().Get("text")

	if msg == "" {
		http.Error(w, "Пустое сообщение", http.StatusBadRequest)
		return
	}

	fullMsg := fmt.Sprintf("[%s]: %s", user, msg)

	logMu.Lock()
	MessageLog = append(MessageLog, fullMsg)
	if len(MessageLog) > 100 {
		MessageLog = MessageLog[len(MessageLog)-100:]
	}
	logMu.Unlock()

	fmt.Println("📢", fullMsg)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Сообщение отправлено: %s", msg)
}

func PollMessage(w http.ResponseWriter, r *http.Request) {
	logMu.RLock()
	msgs := make([]string, len(MessageLog))
	copy(msgs, MessageLog)
	count := len(msgs)
	logMu.RUnlock()

	response := map[string]interface{}{
		"count":    count,
		"messages": msgs,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}