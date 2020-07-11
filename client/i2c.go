package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/balazsgrill/goiotunnel"
	"google.golang.org/grpc"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

type tunneledI2C struct {
	client goiotunnel.IotunnelClient
	handle uint32
}

func NewClient(conn *grpc.ClientConn) goiotunnel.IotunnelClient {
	return goiotunnel.NewIotunnelClient(conn)
}

func NewI2C(client goiotunnel.IotunnelClient, i2cchannel string) (i2c.BusCloser, error) {
	reply, err := client.I2COpen(context.Background(), &goiotunnel.I2COpenRequest{
		Name: i2cchannel,
	})
	if err != nil {
		return nil, err
	}
	return &tunneledI2C{
		client: client,
		handle: reply.Handle,
	}, nil
}

func (t *tunneledI2C) String() string {
	return fmt.Sprintf("tunnelI2C/%d", t.handle)
}

func (t *tunneledI2C) Tx(addr uint16, w, r []byte) error {
	reply, err := t.client.I2CTx(context.Background(), &goiotunnel.I2CTxRequest{
		Address:  uint32(addr),
		Handle:   t.handle,
		TxData:   w,
		RxLength: uint32(len(r)),
	})
	if err != nil {
		return err
	}
	l := copy(r, reply.RxData)
	if l != len(r) {
		return fmt.Errorf("Incorrect number of bytes received (%d)", l)
	}
	return nil
}

func (*tunneledI2C) SetSpeed(f physic.Frequency) error {
	return errors.New("SetSpeed is not supported")
}

func (t *tunneledI2C) Close() error {
	_, err := t.client.Close(context.Background(), &goiotunnel.CloseRequest{
		Handle: t.handle,
	})
	return err
}
