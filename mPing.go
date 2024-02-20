package main

import (
    "log"
    "net"
    "fmt"
    "os"
    "time"
    "golang.org/x/net/icmp"
    "golang.org/x/net/ipv4"
)

func main() {
    c1 := make(chan string)
    c2 := make(chan string)
    for {
        go myPing(os.Args[1], c1)
        chnnl1, ch1Ok := <- c1
        go myPing(os.Args[2], c2)
        chnnl2, ch2Ok := <- c2
	fmt.Println(chnnl1)
	fmt.Println(chnnl2)
        time.Sleep(1 * time.Second)
	if !ch1Ok && !ch2Ok {
            break
	}
    }   
}

func myPing(targetIP string, channel1 chan string) {
    con, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
    if err != nil {
        log.Fatalf("listen err, %s", err)
    }
    defer con.Close()

    start := time.Now()
    wm := icmp.Message{
        Type: ipv4.ICMPTypeEcho, Code: 0,
        Body: &icmp.Echo{
            ID: os.Getpid() & 0xffff, Seq: 1,
            Data: []byte("Some test bytes"),
        },
    }
    
    wb, err := wm.Marshal(nil)
    if err != nil {
        log.Fatal(err)
    }
    if _, err := con.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(targetIP)}); err != nil {
        log.Fatalf("WriteTo err, %s", err)
    }
    
    rb := make([]byte, 1500)
    n, peer, err := con.ReadFrom(rb)
    if err != nil {
        log.Fatal(err)
    }
    rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), rb[:n])
    if err != nil {
        log.Fatal(err)
    }
    
    returnTime := time.Since(start)
    switch rm.Type {
    case ipv4.ICMPTypeEchoReply:
        myTest, _ := rm.Body.Marshal(ipv4.ICMPTypeEchoReply.Protocol())
        rtLen := rm.Body.Len(ipv4.ICMPTypeEchoReply.Protocol())
        pr := fmt.Sprintf("Reply from: %v, Length: %v, Seq: %v, Return Code: %v, Return Time: %v", peer, rtLen, myTest[3], rm.Code, returnTime)
        channel1 <- pr
    default:
        fail := fmt.Sprintf("got %+v; error", rm)
        channel1 <- fail
    }
}
