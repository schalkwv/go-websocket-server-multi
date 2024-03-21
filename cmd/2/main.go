package main

import (
	"fmt"
	"time"
)

func handleNewChannels(
	allReceiversChannel chan []chan uint32,
	oneReceiverChannel chan (chan uint32)) {
	// send empty list of channels to the allReceiversChannel
	currentReceivers := []chan uint32{}
	allReceiversChannel <- currentReceivers
	for {
		// if we get a new receiver on the oneReceiverChannel, add it to the list of receivers
		newReceiver := <-oneReceiverChannel
		currentReceivers = append(currentReceivers, newReceiver)
		// send the updated list of receivers to the allReceiversChannel
		allReceiversChannel <- currentReceivers
	}
}

func sendToChannels(channels chan []chan uint32) {
	tick := time.Tick(1 * time.Second)
	currentChannels := <-channels
	n := uint32(0)
	for {
		n++
		select {
		case <-tick:
			sent := false
			// var n uint32
			// binary.Read(rand.Reader, binary.LittleEndian, &n)
			for i := 0; i < len(currentChannels); i++ {
				currentChannels[i] <- n
				sent = true
			}
			if sent {
				fmt.Println("Sent generated ", n)
			} else {
				fmt.Println("No channels to send to")
			}
		case newChannels := <-channels:
			currentChannels = newChannels
		}
	}
}
func handleChannel(theChannel chan uint32) {
	for {
		val := <-theChannel
		fmt.Println("Got the value ", val)
	}
}

func createChannels(newReceiverChannel chan (chan uint32)) {
	tick := time.Tick(5 * time.Second)
	for {
		<-tick
		fmt.Println("Creating new channel! ")
		newchan := make(chan uint32)
		// this new channel will be received by the handleNewChannels goroutine and added to the list of receivers
		newReceiverChannel <- newchan
		// create a handler for the new channel
		go handleChannel(newchan)
	}
}

func main() {
	// channel that passes around the list of receivers
	allTheReceivers := make(chan []chan uint32)
	// channel for adding a new receiver
	newReceiverChannel := make(chan (chan uint32))
	go handleNewChannels(allTheReceivers, newReceiverChannel)
	go sendToChannels(allTheReceivers)
	createChannels(newReceiverChannel)
}
