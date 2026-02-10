package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
    "sync"
)

var (
    clients = make(map[string]bool)
    mu      sync.RWMutex
    Tasks   = make(chan string, 10)
    lastMsg string
    msgMu   sync.Mutex
)

func main() {
    go Printer()
   
    mux := http.NewServeMux()
    mux.HandleFunc("/", GetMessage)
    mux.HandleFunc("/message", SendMessage)
    mux.HandleFunc("/poll", PollMessage)
   
    fmt.Println("Сервер запущен на localhost:8081")
    log.Fatal(http.ListenAndServe(":8081", mux))
}

func GetMessage(w http.ResponseWriter, r *http.Request) {
    session := r.RemoteAddr
   
    mu.Lock()
    if !clients[session] {
        fmt.Printf("🟢 Подключен клиент: %s\n", session)
    }
    clients[session] = true
    mu.Unlock()
   
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprint(w, `
        <script>
            let lastMsgId = "";
            setInterval(() => {
                fetch('/poll').then(r=>r.text()).then(t=> {
                    if(t && t !== lastMsgId) {
                        eval(t);
                        lastMsgId = t;
                    }
                });
            }, 500);
        </script>`)
}

func SendMessage(w http.ResponseWriter, r *http.Request) {
    
    msg := r.URL.Query().Get("text")
    if msg == "" {
        http.Error(w, "Пустое сообщение", http.StatusBadRequest)
        return
    }

    sender := r.RemoteAddr
    fmt.Printf("📤 Клиент %s отправил: %s\n", sender, msg)
    
    fullMsg := fmt.Sprintf("[%s]: %s", sender, msg)
    
    go func(msg string) {
        Tasks <- msg
    }(fullMsg)

    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    fmt.Fprintf(w, "%s", msg)
}

func PollMessage(w http.ResponseWriter, r *http.Request) {
    msgMu.Lock()
    defer msgMu.Unlock()
   
    if lastMsg != "" {
        fmt.Fprintf(w, `alert("%s");`, lastMsg)
    } else {
        w.Write([]byte(""))
    }
}

func Printer() {
    for msg := range Tasks {
        msgMu.Lock()
        lastMsg = msg
        msgMu.Unlock()
        time.Sleep(3 * time.Second)
        
        msgMu.Lock()
        lastMsg = ""
        msgMu.Unlock()
    }
}
