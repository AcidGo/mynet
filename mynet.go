package main

import (
    "flag"
    "fmt"
    "log"
    "net"
    "os"
    "strconv"
    "time"
)


var (
    isHelp bool

    isServer bool
    isClient bool
    isTCP bool
    isUDP bool
    port int
    connectAddress string
)

func usage() {
    fmt.Fprintf(os.Stderr, `mynet version: 0.0.1
Author: AcidGo
Usage: mynet [-h] [-c|s] [-u|t] [-d <Address>] -p <Port>

Options:
`)
    flag.PrintDefaults()
}

func parseFlag() {
    flag.BoolVar(&isHelp, "h", false, "show the help message")

    flag.BoolVar(&isServer, "s", false, "become the socket server")
    flag.BoolVar(&isClient, "c", false, "become the socket client")
    flag.BoolVar(&isTCP, "t", false, "use the TCP protocol")
    flag.BoolVar(&isUDP, "u", false, "use the UDP protocol")
    flag.IntVar(&port, "p", -1, "set the `Port` for server listening or client dail, it must between 1024 and 65535")
    flag.StringVar(&connectAddress, "d", "", "for client, set the `ConnectAddress` to dail")
    flag.Usage = usage

    flag.Parse()


    if isHelp {
        flag.Usage()
        os.Exit(0)
    }

    if flag.NArg() >= 1 {
        log.Println("The number of args must be zero")
        os.Exit(1)
    }

    if port < 1024 || port > 65535 {
        log.Println("The port must be between 1024 and 65535")
        os.Exit(1)
    }

    if (isUDP && isTCP) || (!isUDP && !isTCP) {
        log.Println("The protocol must be TCP or UDP")
        os.Exit(1)
    }

    if isServer {
        if isClient || (connectAddress != "") || port == 0 {
            log.Println("As server, it cannot be client or has connectAddress")
            os.Exit(1)
        }
    }

    if isClient {
        if isServer || (connectAddress == "") || port == 0 {
            log.Println("As client, it must has connectAddress")
            os.Exit(1)
        }
    }
}

func startTCP4Server(port int) {
    addr := fmt.Sprintf("0.0.0.0:%d", port)
    tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
    if err != nil {
        log.Println("error:", err)
        return
    }

    tcpListener, err := net.ListenTCP("tcp4", tcpAddr)
    if err != nil {
        log.Println("error:", err)
        return
    }
    defer tcpListener.Close()
    log.Println("start listen:", tcpAddr)

    for {
        con, err := tcpListener.AcceptTCP()
        if err != nil {
            log.Println("error:", err)
            return
        }
        
        go func(con net.Conn) {
            defer con.Close()
            for {
                recvData := make([]byte, 4096)
                n, err := con.Read(recvData)
                if err != nil {
                    log.Println("error:", err)
                    break
                }
                recvStr := string(recvData[:n])
                log.Printf("recv from %v: %s\n", con.RemoteAddr(), recvStr)
                sendData := fmt.Sprintf("OK:%s", recvStr)
                con.Write([]byte(sendData))
            }
        }(con)
    }
}

func startTCP4Client(addr string, port int) {
    addr = fmt.Sprintf("%s:%d", addr, port)
    tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
    if err != nil {
        log.Println("error:", err)
        return
    }

    con, err := net.DialTCP("tcp4", nil, tcpAddr)
    if err != nil {
        log.Println("error:", err)
        return
    }
    defer con.Close()

    count := 1
    for {
        sendData := strconv.Itoa(count)
        n, err := con.Write([]byte(sendData))
        if err != nil {
            log.Println("error:", err)
            return
        }
        log.Printf("send %d byte data: %s\n", n, string(sendData))

        recvData := make([]byte, 4096)
        n, err = con.Read(recvData)
        if err != nil {
            log.Println("error:", err)
            return
        }
        recvStr := string(recvData[:n])
        log.Printf("response from %v: %s\n", con.RemoteAddr(), recvStr)
        time.Sleep(time.Second)
        count++
    }
}

func startUDP4Server(port int) {
    addr := fmt.Sprintf("0.0.0.0:%d", port)
    udpAddr, err := net.ResolveUDPAddr("udp4", addr)
    if err != nil {
        log.Println("error:", err)
        return
    }

    udpListener, err := net.ListenUDP("udp4", udpAddr)
    if err != nil {
        log.Println("error:", err)
        return
    }
    defer udpListener.Close()
    log.Println("start listen:", udpAddr)

    for {
        var data [4096]byte
        n, cAddr, err := udpListener.ReadFromUDP(data[:])
        if err != nil {
            fmt.Println("error:", err)
            continue
        }
        log.Printf("recv from %v: %s\n", cAddr, string(data[:n]))
        sendData := fmt.Sprintf("OK:%s", string(data[:n]))
        _, err = udpListener.WriteToUDP([]byte(sendData), cAddr)
        if err != nil {
            log.Println("error:", err)
            continue
        }
    }
}

func startUDP4Client(addr string, prot int) {
    addr = fmt.Sprintf("%s:%d", addr, port)
    udpAddr, err := net.ResolveUDPAddr("udp4", addr)
    if err != nil {
        log.Println("error:", err)
        return
    }

    con, err := net.DialUDP("udp4", nil, udpAddr)
    if err != nil {
        log.Println("error:", err)
        return
    }
    defer con.Close()

    count := 1
    for {
        sendData := strconv.Itoa(count)
        n, err := con.Write([]byte(sendData))
        if err != nil {
            log.Println("error:", err)
            return
        }
        log.Printf("send %d byte data: %s\n", n, string(sendData))

        recvData := make([]byte, 4096)
        n, sAddr, err := con.ReadFromUDP(recvData)
        if err != nil {
            log.Println("error:", err)
            return
        }
        recvStr := string(recvData[:n])
        log.Printf("response from %v: %s\n", sAddr, recvStr)
        time.Sleep(time.Second)
        count++
    }
}

func init() {
    log.SetFlags(log.Ldate|log.Lmicroseconds)
}

func main() {
    parseFlag()

    if isTCP {
        if isServer {
            startTCP4Server(port)
        } else if isClient {
            startTCP4Client(connectAddress, port)
        } else {
            log.Println("it is not both server and client")
            os.Exit(1)
        }
    } else if isUDP {
        if isServer {
            startUDP4Server(port)
        } else if isClient {
            startUDP4Client(connectAddress, port)
        } else {
            log.Println("tt is not both server and client")
            os.Exit(1)
        }
    } else {
        log.Println("the protocol must be TCP or UPD")
    }
}