package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "sync"
    "os"
)

var (
    clients = make(map[net.Conn]bool)
    mu      sync.RWMutex
    msgChan = make(chan string, 10)
)

func main() {
    listener, err := net.Listen("tcp", ":8081")
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()

    fmt.Println("Сервер запущен на :8081")

    go func() {
        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
            text := scanner.Text() + "\n"
            msgChan <- fmt.Sprintf("[server] %s", text)
        }
    }()

    go broadcaster()

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Print("accept error:", err)
            continue
        }

        go handleClient(conn)
    }
}

func handleClient(conn net.Conn) {
    defer func() {
        mu.Lock()
        delete(clients, conn)
        mu.Unlock()
        conn.Close()
    }()

    mu.Lock()
    clients[conn] = true
    mu.Unlock()

    fmt.Printf("Клиент подключился: %s\n", conn.RemoteAddr())

    reader := bufio.NewReader(conn)
    for {
        msg, err := reader.ReadString('\n')
        if err != nil {
            return
        }

        msgChan <- fmt.Sprintf("[%s] %s", conn.RemoteAddr(), msg)
    }
}

func broadcaster() {
    for msg := range msgChan {
        mu.RLock()
        for conn := range clients {
            _, _ = conn.Write([]byte(msg))
        }
        mu.RUnlock()
        fmt.Print(msg)
    }
}
