package main

import (
	"context"
	"log"
	"net"

	"github.com/balazsgrill/goiotunnel"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

type server struct {
	goiotunnel.UnimplementedIotunnelServer
	handleCounter uint32
	i2cBuses      map[uint32]i2c.BusCloser
}

func (s *server) Close(ctx context.Context, req *goiotunnel.CloseRequest) (*empty.Empty, error) {
	bus, ok := s.i2cBuses[req.Handle]
	if !ok {
		return nil, status.Errorf(codes.Unknown, "Given handle is not a valid handle")
	}
	err := bus.Close()
	return &empty.Empty{}, err
}
func (s *server) I2COpen(ctx context.Context, req *goiotunnel.I2COpenRequest) (*goiotunnel.OpenReply, error) {
	bus, err := i2creg.Open(req.Name)
	if err != nil {
		return nil, err
	}

	handle := s.handleCounter
	s.handleCounter++
	s.i2cBuses[handle] = bus
	return &goiotunnel.OpenReply{
		Handle: handle,
	}, nil
}
func (s *server) I2CTx(ctx context.Context, req *goiotunnel.I2CTxRequest) (*goiotunnel.I2CTxReply, error) {
	bus, ok := s.i2cBuses[req.Handle]
	if !ok {
		return nil, status.Errorf(codes.Unknown, "Given handle is not a valid I2C Bus")
	}

	rxData := make([]byte, req.RxLength)
	err := bus.Tx(uint16(req.Address&0xFFFF), req.TxData, rxData)
	return &goiotunnel.I2CTxReply{
		RxData: rxData,
	}, err
}

func main() {
	_, err := host.Init()
	if err != nil {
		log.Fatal(err)
	}
	lis, err := net.Listen("0.0.0.0", "1234")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	goiotunnel.RegisterIotunnelServer(grpcServer, &server{})
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}
