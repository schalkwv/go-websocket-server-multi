package main

import (
	"fmt"
	"time"
)

func handleNewChannels(arrchangen chan [](chan uint32),
	intchangen chan (chan uint32)) {
	currarr := []chan uint32{}
	arrchangen <- currarr
	for {
		newchan := <-intchangen
		currarr = append(currarr, newchan)
		arrchangen <- currarr
	}
}

func sendToChannels(arrchangen chan [](chan uint32)) {
	tick := time.Tick(1 * time.Second)
	currarr := <-arrchangen
	n := uint32(0)
	for {
		n++
		select {
		case <-tick:
			sent := false
			// var n uint32
			// binary.Read(rand.Reader, binary.LittleEndian, &n)
			for i := 0; i < len(currarr); i++ {
				currarr[i] <- n
				sent = true
			}
			if sent {
				fmt.Println("Sent generated ", n)
			}
		case newarr := <-arrchangen:
			currarr = newarr
		}
	}
}
func handleChannel(tchan chan uint32) {
	for {
		val := <-tchan
		fmt.Println("Got the value ", val)
	}
}

func createChannels(intchangen chan (chan uint32)) {
	othertick := time.Tick(5 * time.Second)
	for {
		<-othertick
		fmt.Println("Creating new channel! ")
		newchan := make(chan uint32)
		intchangen <- newchan
		go handleChannel(newchan)
	}
}

func main() {
	arrchangen := make(chan []chan uint32)
	intchangen := make(chan (chan uint32))
	go handleNewChannels(arrchangen, intchangen)
	go sendToChannels(arrchangen)
	createChannels(intchangen)
}
