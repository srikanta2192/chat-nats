package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/nats-io/nats.go"
)

func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'", i, m.Subject, string(m.Data))
}

func main() {
	var urls = flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")
	var name string
	log.SetFlags(0)
	flag.Parse()

	args := flag.Args()
	nc, err := nats.Connect(*urls)
	scanner := bufio.NewScanner(os.Stdin)
	subj, i := args[0], 0

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Enter your name")

	for scanner.Scan() {
		name = scanner.Text()
		if name != "" {
			nc.Publish("msg.name", []byte(name))
			break
		}
	}

	fmt.Println("You joined the chat")

	go func() {
		nc.Subscribe("name."+name, func(msg *nats.Msg) {
			i++
			printMsg(msg, i)
		})
		nc.Flush()
	}()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	go chatTyper(nc, name)

	log.Printf("Listening on [%s]", subj)

	runtime.Goexit()
}

func chatTyper(nc *nats.Conn, name string) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		text := scanner.Text()
		if text == "" {
			continue
		}
		err := nc.Publish("msg.listen", []byte(name+"."+text))
		if err != nil {
			fmt.Println("error sending message ", err.Error())
			break
		}
	}
}
