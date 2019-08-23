package main

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	core "github.com/libp2p/go-libp2p-core"
)

// setupHosts is responsible for creating libp2p hosts.
func setupHosts(n int, initialPort int) ([]*libp2pPubSub, []*core.Host, *map[string]int) {
	// hosts used in libp2p communications
	hosts := make([]*core.Host, n)
	pubSubs := make([]*libp2pPubSub, n)
	nodeIds := make(map[string]int)
	for i := range hosts {

		pubsub := new(libp2pPubSub)

		// creating libp2p hosts
		host := pubsub.createPeer(i, initialPort+i)
		hosts[i] = host
		// creating pubsubs
		pubsub.initializePubSub(*host)
		pubSubs[i] = pubsub
		nodeIds[(*host).ID().Pretty()] = i
	}
	return pubSubs, hosts, &nodeIds
}

// setupNetworkTopology sets up a simple network topology for test.
func setupNetworkTopology(hosts []*core.Host, n int) {

	// Connect hosts to each other in a topology
	// host0 ---- host1 ---- host2 ----- host3 ----- host4
	//	 			|		   				|    	   |
	//	            ------------------------------------
	for i := 0; i < n-1; i++ {
		connectHostToPeer(*hosts[i], getLocalhostAddress(*hosts[i+1]))
		//connectHostToPeer(*hosts[2], getLocalhostAddress(*hosts[1]))
		//connectHostToPeer(*hosts[3], getLocalhostAddress(*hosts[2]))
		//connectHostToPeer(*hosts[4], getLocalhostAddress(*hosts[3]))
		//connectHostToPeer(*hosts[4], getLocalhostAddress(*hosts[1]))
		//connectHostToPeer(*hosts[3], getLocalhostAddress(*hosts[1]))
		//connectHostToPeer(*hosts[4], getLocalhostAddress(*hosts[1]))
	}
	// Wait so that subscriptions on topic will be done and all peers will "know" of all other peers
	time.Sleep(time.Second * 2)

}

func startListening(pubSubs []*libp2pPubSub, hosts []*core.Host, nodeIds *map[string]int, n int) {
	wg := &sync.WaitGroup{}

	for i, host := range hosts {

		wg.Add(1)
		go func(host *core.Host, pubSub *libp2pPubSub) {
			for i := 0; i < n*n+n; i++ {
				_, msg := pubSub.Receive()
				id := (*nodeIds)[(*host).ID().Pretty()]
				fmt.Printf("Node %d received Message: '%s'\n", id, msg)
				match, err := regexp.MatchString("msg [0-9] ack [0-9]", msg)
				if err != nil {
					panic(err)
				}
				if !match {
					pubSub.Broadcast(fmt.Sprintf("%s ack %d", msg, id))
				}
			}
			wg.Done()
		}(host, pubSubs[i])
		//fmt.Printf("Broadcasting a message from node %d\n",i)
		pubSubs[i].Broadcast(fmt.Sprintf("msg %d", i))
	}
	wg.Wait()
	fmt.Println("The END")
}

func main() {

	n := 10
	initialPort := 9900

	// Create hosts in libp2p
	pubSubs, hosts, nodeIds := setupHosts(n, initialPort)

	defer func() {
		fmt.Println("Closing hosts")
		for _, h := range hosts {
			_ = (*h).Close()
		}
	}()

	setupNetworkTopology(hosts, n)

	startListening(pubSubs, hosts, nodeIds, n)

}
