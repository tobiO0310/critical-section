package main

import (
	"context"
	proto "cs/proto"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// check env variable "CS_ID" for id, and lookup table for which port to listen on, and which ports to connect to

var portOfId0 = 3000
var serverAmount = 10

func CRITICAL_SECTION_RUNNER() {
	log.Printf("[CRITICAL] very critical guys!!")
}

type CriticalSectionService struct {
	proto.UnimplementedCriticalSectionServiceServer

	id        int64
	timestamp uint64
	mutex     sync.Mutex

	connections map[int64]proto.CriticalSectionServiceClient

	doIWantIt      bool
	whenDidIWantIt uint64

	queue          []int64
	replies_gotten uint64
}

func (s *CriticalSectionService) CanI(ctx context.Context, in *proto.CanIRequest) (*proto.CanIResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.timestamp = max(s.timestamp, uint64(in.Time))
	s.timestamp += 1

	log.Printf("Gotten request from %d", in.Id)

	if !s.doIWantIt || in.Time < int64(s.whenDidIWantIt) || in.Time == int64(s.whenDidIWantIt) && in.Id < s.id {
		log.Printf("%d is allowed to run, either because I don't want it or they asked first", in.Id)
		s.timestamp += 1
		go s.connections[in.Id].YouMayNow(
			context.Background(),
			&proto.YouMayNowRequest{Time: int64(s.timestamp)},
		)
		return &proto.CanIResponse{Time: int64(s.timestamp)}, nil
	}
	log.Printf("%d must wait for me!", in.Id)

	s.queue = append(s.queue, in.Id)
	s.timestamp += 1
	return &proto.CanIResponse{Time: int64(s.timestamp)}, nil
}

func (s *CriticalSectionService) YouMayNow(ctx context.Context, in *proto.YouMayNowRequest) (*proto.YouMayNowResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.timestamp = max(s.timestamp, uint64(in.Time))
	s.timestamp += 1

	log.Printf("Gotten acceptance from a server, will count it")

	s.replies_gotten += 1

	s.timestamp += 1
	return &proto.YouMayNowResponse{Time: int64(s.timestamp)}, nil
}

func listen(id int) net.Listener {
	lis, netErr := net.Listen("tcp", fmt.Sprintf("localhost:%d",  portOfId0+id))
	if netErr != nil {
		log.Fatalf("[SERVER] failed to listen: %v\n", netErr)
	} else {
		log.Printf("[SERVER]\tserver is listening on rpc://localhost:%d\n\t\t\t\tPress Ctrl+C to exit.\n", portOfId0+id)
	}
	return lis
}

func main() {
	id, idErr := strconv.ParseInt(os.Getenv("CS_ID"), 10, 16)
	if idErr != nil {
		fmt.Println("Invalid id given in CS_ID environment variable")
		return
	}
	logFile, netErr := os.OpenFile(
		fmt.Sprintf("output_server_%d.log", id),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if netErr != nil {
		log.Fatalf("Failed to open log file: %v", netErr)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	lis := listen(int(id))
	defer lis.Close()

	go func() {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		// Block until the context is canceled
		<-ctx.Done()

		fmt.Println("\n[SERVER] Interrupt received, shutting down gracefully...")
		lis.Close()
	}()

	service := CriticalSectionService{
		id: id,
		connections: make(map[int64]proto.CriticalSectionServiceClient, serverAmount),
	}

	go func() {
		// connect to other servers
		for i := range serverAmount {
			if int64(i) == id {
				continue // don't try to listen to self
			}
			log.Printf("Trying to connect to server with id %d, at localhost:%d ...\n", i, portOfId0+i)
			s := fmt.Sprintf("localhost:%d", portOfId0+i)
			conn, err := grpc.NewClient(s, grpc.WithTransportCredentials(insecure.NewCredentials()))
			for err != nil {
				log.Printf("Failed to connect, error:%v\t\tretrying connection...\n", err)
				time.Sleep(250 * time.Millisecond)
				conn, err = grpc.NewClient(s, grpc.WithTransportCredentials(insecure.NewCredentials()))
			}
			log.Printf("Connected, will save connection now.\n")

			client := proto.NewCriticalSectionServiceClient(conn)
			service.mutex.Lock()
			service.connections[int64(i)] = client
			service.mutex.Unlock()
		}
	}()

	go func() {
		// ask at some point ! :D
		for len(service.connections) < serverAmount-1 { // -1 for nirself
			// wait
			time.Sleep(time.Second)
		}

		sleepTime := time.Duration(rand.Intn(1_000) * 1_000_000);
		log.Printf("Sleeping for %d milliseconds (for random access) and then trying to get into critical section", sleepTime.Milliseconds())
		time.Sleep(sleepTime)
		for i, client := range service.connections {
			log.Printf("Sending request to %d... ", i)
			service.mutex.Lock()
			service.timestamp += 1
			req := &proto.CanIRequest{ Time: int64(service.timestamp), Id: id };
			service.mutex.Unlock()
			res, err := client.CanI(
				context.Background(),
				req,
			)
			if err != nil {
				log.Fatalf("failed\n");
			}
			service.mutex.Lock()
			service.timestamp = max(service.timestamp, uint64(res.Time));
			service.timestamp += 1;
			service.mutex.Unlock()
			log.Printf("done\n")
		}
		log.Println("Sent all requests, waiting for enough responses");

		for service.replies_gotten < uint64(serverAmount)-1 { // -1 for nirself!
			time.Sleep(10 * time.Millisecond) // wait till has enough replies
		}

		service.mutex.Lock()
		CRITICAL_SECTION_RUNNER() // OMG I CAN RUN THE CRITICAL SECTION?!?!?!

		for asked := range service.queue { // now i'll let the others waiting on me get in on the action :)
			service.timestamp += 1
			service.connections[int64(asked)].YouMayNow(context.Background(), &proto.YouMayNowRequest{})
		}
		service.queue = make([]int64, 0);
		service.mutex.Unlock()
	}()

	grpcServer := grpc.NewServer()
	proto.RegisterCriticalSectionServiceServer(grpcServer, &service)
	grpcServer.Serve(lis)
}
