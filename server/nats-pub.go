package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/nats-io/nats.go"
)

var users [10]string

func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'", i, m.Subject, string(m.Data))
}

func main() {
	var urls = flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")
	log.SetFlags(0)
	flag.Parse()
	args := flag.Args()
	nc, err := nats.Connect(*urls)
	if err != nil {
		log.Fatal(err)
	}
	subj, msg := args[0], []byte(args[1])

	defer nc.Close()
	go subscription(nc)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		go sendMessageToNonAuthor("", (text), nc)
		if err != nil {
			fmt.Println("error sending message ", err.Error())
			break
		}
	}
	nc.Flush()
	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Published [%s] : '%s'\n", subj, msg)
	}
	runtime.Goexit()
}

func sendMessageToNonAuthor(user string, msg string, nc *nats.Conn) {
	for _, s := range users {
		if s != user {
			nc.Publish("name."+s, []byte(msg))
		}
	}
}

func subscription(nc *nats.Conn) {
	k, i := 0, 0
	nc.Subscribe("msg.name", func(msg *nats.Msg) {
		users[k] = string(msg.Data)
		k++
		printMsg(msg, k)
	})

	nc.Subscribe("msg.listen", func(msg *nats.Msg) {
		i++
		var userMessage []string
		userMessage = strings.Split(string(msg.Data), ".")
		go sendMessageToNonAuthor(userMessage[0], userMessage[1], nc)
		printMsg(msg, i)
	})
}
