package core

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"sort"
	"time"

	base58 "github.com/jbenet/go-base58"
	"github.com/science-engineering-art/kademlia-grpc/interfaces"
	"github.com/science-engineering-art/kademlia-grpc/pb"
	"github.com/science-engineering-art/kademlia-grpc/structs"
	"github.com/science-engineering-art/kademlia-grpc/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type FullNode struct {
	pb.UnimplementedFullNodeServer
	dht *DHT
}

func NewFullNode(nodeIP string, nodePort, bootstrapPort int, storage interfaces.Persistence, isBootstrapNode bool) *FullNode {

	id, _ := utils.NewID(nodeIP, nodePort)
	node := structs.Node{ID: id, IP: nodeIP, Port: nodePort}
	routingTable := structs.NewRoutingTable(node)
	dht := DHT{Node: node, RoutingTable: routingTable, Storage: storage}
	fullNode := FullNode{dht: &dht}

	go func() {
		for {
			<-time.After(10 * time.Second)
			fmt.Println("\nROUTING TABLE:")
			fmt.Printf("ME: %v\n\n", fullNode.dht.Node)
			fullNode.PrintRoutingTable()
			fmt.Printf("\n")
		}
	}()

	fullNode.joinNetwork(bootstrapPort)

	if isBootstrapNode {
		go fullNode.bootstrap(bootstrapPort)
	}

	return &fullNode
}

// Create gRPC Server
func (fn *FullNode) CreateGRPCServer(grpcServerAddress string) {
	grpcServer := grpc.NewServer()

	pb.RegisterFullNodeServer(grpcServer, fn)
	reflection.Register(grpcServer)

	go fn.Republish()

	listener, err := net.Listen("tcp", grpcServerAddress)
	if err != nil {
		log.Fatal("cannot create grpc server: ", err)
	}

	log.Printf("start gRPC server on %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot create grpc server: ", err)
	}
}

///////////////////////////////////////////////////////
////// 										     //////
//////            RPC KADEMLIA PROTOCOL          //////
//////    									     //////
///////////////////////////////////////////////////////

func (fn *FullNode) Ping(ctx context.Context, sender *pb.Node) (*pb.Node, error) {

	// add the sender to the Routing Table
	_sender := structs.Node{
		ID:   sender.ID,
		IP:   sender.IP,
		Port: int(sender.Port),
	}
	fn.dht.RoutingTable.AddNode(_sender)

	receiver := &pb.Node{
		ID:   fn.dht.ID,
		IP:   fn.dht.IP,
		Port: int32(fn.dht.Port),
	}

	return receiver, nil
}

func (fn *FullNode) Store(stream pb.FullNode_StoreServer) error {
	//fmt.Printf("INIT FullNode.Store()\n\n")
	// defer //fmt.Printf("END FullNode.Store()\n\n")

	key := []byte{}
	buffer := []byte{}
	var init int64 = 0

	for {
		data, err := stream.Recv()
		if data == nil {
			//fmt.Printf("END Streaming\n\n")
			break
		}
		if err != nil {
			//fmt.Printf("EXIT line:133 Store() method\n\n")
			return errors.New("missing chunck")
		}

		if init == 0 {
			//fmt.Printf("INIT Streaming\n\n")
			// add the sender to the Routing Table
			sender := structs.Node{
				ID:   data.Sender.ID,
				IP:   data.Sender.IP,
				Port: int(data.Sender.Port),
			}
			fn.dht.RoutingTable.AddNode(sender)
		}

		key = data.Key
		if init == data.Value.Init {
			buffer = append(buffer, data.Value.Buffer...)
			init = data.Value.End
		} else {
			//fmt.Printf("ERROR missing chunck\n\n")
			return err
		}
		//fmt.Printf("OKKKK ===> FullNode(%s).Recv(%d, %d)\n", fn.dht.IP, data.Value.Init, data.Value.End)
	}
	// //fmt.Println("Received Data:", buffer)

	err := fn.dht.Store(key, &buffer)
	if err != nil {
		//fmt.Printf("ERROR line:140 DHT.Store()\n\n")
		return err
	}
	return nil
}

func (fn *FullNode) FindNode(ctx context.Context, target *pb.Target) (*pb.KBucket, error) {
	// add the sender to the Routing Table
	sender := structs.Node{
		ID:   target.Sender.ID,
		IP:   target.Sender.IP,
		Port: int(target.Sender.Port),
	}
	fn.dht.RoutingTable.AddNode(sender)

	bucket := fn.dht.FindNode(&target.ID)

	return utils.GetKBucketFromNodeArray(bucket), nil
}

func (fn *FullNode) FindValue(target *pb.Target, stream pb.FullNode_FindValueServer) error {
	// add the sender to the Routing Table
	sender := structs.Node{
		ID:   target.Sender.ID,
		IP:   target.Sender.IP,
		Port: int(target.Sender.Port),
	}
	fn.dht.RoutingTable.AddNode(sender)

	value, neighbors := fn.dht.FindValue(&target.ID, target.Init, target.End)
	response := pb.FindValueResponse{}

	if value == nil && neighbors != nil {
		response = pb.FindValueResponse{
			KNeartestBuckets: utils.GetKBucketFromNodeArray(neighbors),
			Value: &pb.Data{
				Init:   0,
				End:    0,
				Buffer: []byte{},
			},
		}
	} else if value != nil && neighbors == nil {
		//fmt.Println("Value from FindValue:", value)
		response = pb.FindValueResponse{
			KNeartestBuckets: &pb.KBucket{Bucket: []*pb.Node{}},
			Value: &pb.Data{
				Init:   0,
				End:    int64(len(*value)),
				Buffer: *value,
			},
		}
	} else {
		return errors.New("check code because this case shouldn't be valid")
	}
	stream.Send(&response)

	return nil
}

///////////////////////////////////////////////////////
////// 										     //////
//////            CORE KADEMLIA PROTOCOL         //////
//////    									     //////
///////////////////////////////////////////////////////

func (fn *FullNode) LookUp(target []byte) ([]structs.Node, error) {
	sl := fn.dht.RoutingTable.GetClosestContacts(structs.Alpha, target, []*structs.Node{})

	contacted := make(map[string]bool)
	contacted[string(fn.dht.ID)] = true

	if len(*sl.Nodes) == 0 {
		return nil, nil
	}

	for {
		addedNodes := 0

		for i, node := range *sl.Nodes {
			if i >= structs.Alpha {
				break
			}
			if contacted[string(node.ID)] {
				continue
			}
			contacted[string(node.ID)] = true

			// get RPC client
			client := NewClientNode(node.IP, 8080)
			if client == nil {
				continue
			}

			// function to add the received nodes into the short list
			addRecvNodes := func(recvNodes *pb.KBucket) {
				kBucket := []*structs.Node{}
				for _, pbNode := range recvNodes.Bucket {
					if !contacted[string(pbNode.ID)] {
						kBucket = append(kBucket, &structs.Node{
							ID:   pbNode.ID,
							IP:   pbNode.IP,
							Port: int(pbNode.Port),
						})
						addedNodes++
					}
				}
				sl.Append(kBucket)
			}
			// //fmt.Println("Before timeout")
			// <-time.After(10 * time.Second)
			// //fmt.Println("After timeout")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			recvNodes, err := client.FindNode(ctx,
				&pb.Target{
					ID: node.ID,
					Sender: &pb.Node{
						ID:   fn.dht.ID,
						IP:   fn.dht.IP,
						Port: int32(fn.dht.Port),
					},
				},
			)
			if err != nil && err.Error() == "rpc error: code = DeadlineExceeded desc = context deadline exceeded" {
				//fmt.Println("Crash connection")
				sl.RemoveNode(&node)
				continue
			}
			if err != nil {
				return nil, err
			}
			addRecvNodes(recvNodes)
		}

		sl.Comparator = fn.dht.ID
		sort.Sort(sl)

		if addedNodes == 0 {
			//fmt.Println("0 added nodes")
			break
		}
	}

	kBucket := []structs.Node{}

	for i, node := range *sl.Nodes {
		if i == structs.K {
			break
		}
		//fmt.Println("append node", node.IP)
		kBucket = append(kBucket, structs.Node{
			ID:   node.ID,
			IP:   node.IP,
			Port: node.Port,
		})
	}
	return kBucket, nil
}

func (fn *FullNode) StoreValue(key string, data *[]byte) (string, error) {
	//fmt.Printf("INIT FullNode.StoreValue(%s) method\n\n", key)
	// defer //fmt.Printf("EXIT FullNode.StoreValue(%s) method\n\n", key)

	keyHash := base58.Decode(key)
	nearestNeighbors, err := fn.LookUp(keyHash)
	if err != nil {
		//fmt.Printf("ERROR LookUP() method\n\n")
		return "", err
	}

	if len(nearestNeighbors) < structs.K {
		err := fn.dht.Store(keyHash, data)
		if err != nil {
			fmt.Printf("ERROR DHT.Store(Me)\n\n")
		}
	}

	for index, node := range nearestNeighbors {
		if index == structs.K-1 && utils.ClosestNodeToKey(keyHash, fn.dht.ID, node.ID) == -1 {
			err := fn.dht.Store(keyHash, data)
			if err != nil {
				fmt.Printf("ERROR DHT.Store(Me)\n\n")
			}
			break
		}

		client := NewClientNode(node.IP, node.Port)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if client == nil {
			continue
		}

		sender, err := client.Store(ctx)
		if err != nil {
			//fmt.Printf("ERROR Store(%v, %d) method", node.IP, node.Port)
			if ctx.Err() == context.DeadlineExceeded {
				// Handle timeout error
				//fmt.Println("Timeout exceeded")
				continue
			}
			//fmt.Println(err.Error())
		}
		// //fmt.Println("data bytes", dataBytes)

		for i := 0; i < len(*data); i += 1024 {
			j := int(math.Min(float64(i+1024), float64(len(*data))))

			err = sender.Send(
				&pb.StoreData{
					Sender: &pb.Node{
						ID:   fn.dht.ID,
						IP:   fn.dht.IP,
						Port: int32(fn.dht.Port),
					},
					Key: keyHash,
					Value: &pb.Data{
						Init:   int64(i),
						End:    int64(j),
						Buffer: (*data)[i:j],
					},
				},
			)
			if err != nil {
				//fmt.Printf("ERROR SendChunck(0, %d) method\n\n", len(*data))
				break
				// return "", err
			}
			//fmt.Printf("OKKKK ===> FullNode(%s).Send(%d, %d)\n", fn.dht.IP, i, j)
		}

	}

	// //fmt.Println("Stored ID: ", key, "Stored Data:", data)
	//fmt.Println("===> OKKKK")
	return key, nil
}

func (fn *FullNode) GetValue(target string, start int64, end int64) ([]byte, error) {
	keyHash := base58.Decode(target)

	val, err := fn.dht.Storage.Read(keyHash, start, end)
	if err == nil {
		return *val, nil
	}

	nearestNeighbors, err := fn.LookUp(keyHash)
	if err != nil {
		return nil, nil
	}
	////fmt.Println(nearestNeighbors)
	buffer := []byte{}

	for _, node := range nearestNeighbors {
		if len(target) == 0 {
			//fmt.Println("Invalid target decoding.")
			continue
		}

		clientChnn := make(chan pb.FullNodeClient)

		go func() {
			client := NewClientNode(node.IP, node.Port)
			if client == nil {
				return
			}
			clientChnn <- client.FullNodeClient
			//fmt.Println("Channel value is: ", clientChnn)
		}()

		//fmt.Println("Init Select-Case")
		select {
		case <-time.After(5 * time.Second):
			//fmt.Println("Timeout")
			continue
		case client := <-clientChnn:
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			if client == nil {
				continue
			}

			//fmt.Println("Init FindValue")
			receiver, err := client.FindValue(ctx,
				&pb.Target{
					ID: keyHash,
					Sender: &pb.Node{
						ID:   fn.dht.ID,
						IP:   fn.dht.IP,
						Port: int32(fn.dht.Port),
					},
					Init: start,
					End:  end,
				},
			)
			if err != nil || receiver == nil {
				continue
			}
			//fmt.Println("End FindValue")
			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					// Handle timeout error
					// //fmt.Println("Timeout exceeded")
					continue
				}
				//fmt.Println(err.Error())
				continue
			}
			var init int64 = 0

			for {
				data, err := receiver.Recv()
				if err != nil {
					//fmt.Println(err.Error())
					break
				}
				if data == nil {
					break
				} else if init == data.Value.Init {
					buffer = append(buffer, data.Value.Buffer...)
					init = data.Value.End
				}
			}
			////fmt.Println("Received value from STREAMING in GetValue():", buffer)
			// Received data
			if len(buffer) > 0 {
				goto RETURN
				// break
			}
		}
	}

RETURN:
	return buffer, nil
}

///////////////////////////////////////////////////////
////// 										     //////
//////                JOIN NETWORK               //////
//////    									     //////
///////////////////////////////////////////////////////

func (fn *FullNode) bootstrap(port int) {
	strAddr := fmt.Sprintf("0.0.0.0:%d", port)

	addr, err := net.ResolveUDPAddr("udp4", strAddr)
	if err != nil {
		fmt.Println("line:494")
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("line:500")
		log.Fatal(err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		_, rAddr, err := conn.ReadFrom(buffer)
		if err != nil {
			fmt.Println("line:510")
			log.Fatal(err)
		}
		// Check if node are in the same port
		portInt := binary.LittleEndian.Uint32(buffer[20:24])
		if int(portInt) != fn.dht.Port {
			fmt.Println("UDP Message with != Ports")
			continue
		}
		host, _, _ := net.SplitHostPort(rAddr.String())
		rAddr = &net.TCPAddr{IP: net.ParseIP(host), Port: port + 1}

		go func(rAddr net.Addr, buffer []byte) {
			connChan := make(chan net.Conn)

			go func() {
				respConn, err := net.Dial("tcp", rAddr.String())
				for err != nil {
					respConn, err = net.Dial("tcp", rAddr.String())
				}
				connChan <- respConn
			}()

			kBucket, err := fn.LookUp(buffer[:20])
			if err != nil {
				fmt.Println("line:529")
				log.Fatal(err)
			}
			kBucket = append(kBucket, fn.dht.Node)

			//Convert port from byte to int
			portInt := binary.LittleEndian.Uint32(buffer[20:24])
			intVal := int(portInt)
			if intVal != fn.dht.Port {
				return
			}

			host, _, _ := net.SplitHostPort(rAddr.String())
			id, _ := utils.NewID(host, intVal)
			fn.dht.RoutingTable.AddNode(structs.Node{ID: id, IP: host, Port: intVal})

			bytesKBucket, err := utils.SerializeMessage(&kBucket)
			if err != nil {
				fmt.Println("line:544")
				log.Fatal(err)
			}

			respConn := <-connChan
			respConn.Write(*bytesKBucket)
		}(rAddr, buffer)
	}
}

func (fn *FullNode) joinNetwork(boostrapPort int) {
	broadcastAddr := net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: boostrapPort,
	}

	fmt.Println("In Dial UDP")
	conn, err := net.DialUDP("udp4", nil, &broadcastAddr)
	if err != nil {
		fmt.Println("line:558")
		log.Fatal(err)
	}

	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		fmt.Println("line:564")
		log.Fatal(err)
	}

	dataToSend := fn.dht.ID
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(fn.dht.Port))
	dataToSend = append(dataToSend, bs...)
	_, err = conn.Write(dataToSend)
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()

	address := fmt.Sprintf("%s:%d", host, boostrapPort+1)
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		fmt.Println("line:579")
		log.Fatal(err)
	}

	fmt.Println("In Listen TCP", tcpAddr)
	lis, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("line:586")
		log.Fatal(err)
	}

	// ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
	// defer cancel()

RECONNECT:
	connChannel := make(chan net.Conn)

	go func() {
		tcpConn, _ := lis.AcceptTCP()
		connChannel <- tcpConn
	}()

	select {
	case <-time.After(5 * time.Second):
		lis.Close()
		break
		// go func() {
		// 	time.After(5 * time.Minute)
		// 	go fn.joinNetwork(boostrapPort)
		// }()
	case tcpConn := <-connChannel:
		kBucket, err := utils.DeserializeMessage(tcpConn)
		if err != nil {
			fmt.Println("line:604")
			log.Fatal(err)
		}
		fmt.Println("line:610 In Deserialize Messages")

		for i := 0; i < len(*kBucket); i++ {
			node := (*kBucket)[i]

			if node.Port != fn.dht.Port {
				goto RECONNECT
			}

			// fmt.Println("Before NewClient")
			client := NewClientNode(node.IP, node.Port)
			// fmt.Println("After NewClient")
			if client == nil {
				continue
			}

			// fmt.Println("Before Ping")
			recvNode, err := client.Ping(fn.dht.Node)
			// fmt.Println("After Ping")
			if err != nil {
				log.Fatal(err)
			}

			if recvNode.Equal(node) {
				fn.dht.RoutingTable.AddNode(node)
			} else {
				fmt.Println("Recv Node", *recvNode)
				fmt.Println("Node", node)
				fmt.Println("Me", fn.dht.Node)
				log.Fatal(errors.New("bad ping"))
			}
		}
		fmt.Println("Finish join network")
	}
}

func (fn *FullNode) PrintRoutingTable() {
	KBuckets := fn.dht.RoutingTable.KBuckets

	for i := 0; i < len(KBuckets); i++ {
		for j := 0; j < len(KBuckets[i]); j++ {
			fmt.Println(KBuckets[i][j])
		}
	}
}

///////////////////////////////////////////////////////
////// 										     //////
//////                REPLICATION                //////
//////    									     //////
///////////////////////////////////////////////////////

func (fn *FullNode) Republish() {
	for {
		<-time.After(time.Hour)
		keys := fn.dht.Storage.GetKeys()
		if keys == nil {
			break
		}
		for _, key := range keys {
			data, _ := fn.dht.Storage.Read(key, 0, 0)
			keyStr := base58.Encode(key)
			go func() {
				fn.StoreValue(keyStr, data)
			}()
		}
	}
}
