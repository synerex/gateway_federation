package main

// Federation gateway
// open REST port for federation.

import (
	"flag"
	"log"
	"strings"
	"sync"

	api "github.com/synerex/synerex_api"
	"github.com/synerex/synerex_nodeapi"
	sxutil "github.com/synerex/synerex_sxutil"
	"golang.org/x/net/context"
)

var (
	nodesrv = flag.String("nodesrv", "127.0.0.1:9990", "Node ID Server")
	//	port    = flag.Int("rest_port", 10099, "federation port")
	name   = flag.String("name", "Federation-Gateway", "Name of Fedration Gateway")
	server = flag.String("server", "", "Speficy Synerex Server name")

	idlist          []uint64
	spMap           map[uint64]*sxutil.SupplyOpts
	mu              sync.Mutex
	sxServerAddress string
)

func listenGatewayMsg(sg api.Synerex_SubscribeGatewayClient) {
	//	ctx := context.Background() //
	for {
		msg, err := sg.Recv()
		if err == nil {
			log.Printf("Receive Gateway message :%v", msg)

		} else {
			log.Printf("Error on gateway receive! :%v", err)
		}
	}
}

func main() {
	flag.Parse()
	go sxutil.HandleSigInt()
	sxutil.RegisterDeferFunction(sxutil.UnRegisterNode)

	sxo := &sxutil.SxServerOpt{
		ServerInfo: "",
		NodeType:   synerex_nodeapi.NodeType_GATEWAY,
		ClusterId:  0,
		AreaId:     "Default",
		GwInfo:     *server,
	}

	channelTypes := []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	// obtain synerex server address from nodeserv
	srvs, err := sxutil.RegisterNode(*nodesrv, *name, channelTypes, sxo)
	if err != nil {
		log.Fatal("Can't register node...")
	}
	log.Printf("Connecting Servers [%s]\n", srvs)
	servers := strings.Split(srvs, ",")

	wg := sync.WaitGroup{} // for syncing other goroutines
	client0 := sxutil.GrpcConnectServer(servers[0])

	channels := []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8}

	gi := &api.GatewayInfo{
		ClientId:    sxutil.GenerateIntID(),        // new client_ID
		GatewayType: api.GatewayType_BIDIRECTIONAL, /// default
		Channels:    channels,
	}
	ctx := context.Background() //
	sg0, err := client0.SubscribeGateway(ctx, gi)
	if err != nil {
		log.Printf("Synerex subscribe Error %v\n", err)
	} else {

		wg.Add(1)
		go listenGatewayMsg(sg0)
	}

	wg.Wait()
	sxutil.CallDeferFunctions() // cleanup!

}
