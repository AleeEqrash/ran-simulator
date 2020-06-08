// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package topo

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/southbound"
	topodevice "github.com/onosproject/onos-topo/api/device"
	"github.com/onosproject/ran-simulator/api/types"
	"github.com/onosproject/ran-simulator/pkg/utils"
	"google.golang.org/grpc"
	"io"
	"time"
)

var log = logging.GetLogger("southbound", "topo")

// CellCreationHandler a call back function to avoid import cycle
type CellCreationHandler func(device *topodevice.Device) error

// CellDeletionHandler a call back function to avoid import cycle
type CellDeletionHandler func(device *topodevice.Device) error

// ConnectToTopo is a go function that listens for the connection of onos-topo and
// listens out for Cell instances on it
func ConnectToTopo(ctx context.Context, topoEndpoint string,
	serverParams utils.ServerParams, createHandler CellCreationHandler,
	deleteHandler CellDeletionHandler) (topodevice.DeviceServiceClient, error) {

	log.Infof("Connecting to ONOS Topo...%s", topoEndpoint)
	// Attempt to create connection to the Topo
	opts, err := certs.HandleCertPaths(serverParams.CaPath, serverParams.KeyPath, serverParams.CertPath, true)
	if err != nil {
		log.Fatal(err)
	}
	opts = append(opts, grpc.WithStreamInterceptor(southbound.RetryingStreamClientInterceptor(time.Second)))
	conn, err := southbound.Connect(ctx, topoEndpoint, "", "", opts...)
	if err != nil {
		log.Fatal("Failed to connect to %s. Retry. %s", topoEndpoint, err)
	}

	topoClient := topodevice.NewDeviceServiceClient(conn)
	stream, err := topoClient.List(context.Background(), &topodevice.ListRequest{Subscribe: true})
	if err != nil {
		return nil, err
	}
	for {
		in, err := stream.Recv() // Block here and wait for events from topo
		if err == io.EOF {
			// read done.
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if in.GetDevice().GetType() != types.E2NodeType {
			continue
		}
		if in.GetDevice().GetVersion() != types.E2NodeVersion100 {
			log.Warnf("Only version %s of %s is supported", types.E2NodeVersion100, types.E2NodeType)
			continue
		}
		switch in.Type {
		case topodevice.ListResponse_NONE:
			err := createHandler(in.GetDevice())
			if err != nil {
				log.Warnf("Unable to create cell from %s. %s", in.GetDevice().GetID(), err.Error())
				continue
			}
		case topodevice.ListResponse_ADDED:
			err := createHandler(in.GetDevice())
			if err != nil {
				log.Warnf("Unable to create cell from %s. %s", in.GetDevice().GetID(), err.Error())
				continue
			}
		case topodevice.ListResponse_REMOVED:
			err := deleteHandler(in.GetDevice())
			if err != nil {
				log.Warnf("Unable to delete cell from %s. %s", in.GetDevice().GetID(), err.Error())
				continue
			}
		default:
			log.Warnf("topo event type %s not yet handled for %s", in.Type, in.Device.ID)
		}
	}
}
