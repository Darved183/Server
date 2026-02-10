package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "os"
)

func main() {

    host := "localhost"
    port := "8081"

    conn, err := net.Dial("tcp", host+":"+port)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    fmt.Println("Подключено к серверу. Вводите сообщения (Enter чтобы отправить, Ctrl+C чтобы выйти)")

    go func() {
        reader := bufio.NewReader(conn)
        for {
            msg, err := reader.ReadString('\n')
            if err != nil {
                break
            }
            fmt.Println("<<", msg[:len(msg)-1])
        }
    }()

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        text := scanner.Text() + "\n"
        _, err := conn.Write([]byte(text))
        if err != nil {
            log.Print("write error:", err)
            return
        }
    }
}
