// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: proto/prysm/v1alpha1/light_update.proto

/*
Package eth is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package eth

import (
	"context"
	"io"
	"net/http"

	"github.com/golang/protobuf/ptypes/empty"
	emptypb "github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	github_com_prysmaticlabs_eth2_types "github.com/prysmaticlabs/eth2-types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = metadata.Join
var _ = github_com_prysmaticlabs_eth2_types.Epoch(0)
var _ = emptypb.Empty{}
var _ = empty.Empty{}

func request_LightUpdate_GetUpdates_0(ctx context.Context, marshaler runtime.Marshaler, client LightUpdateClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq emptypb.Empty
	var metadata runtime.ServerMetadata

	msg, err := client.GetUpdates(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_LightUpdate_GetUpdates_0(ctx context.Context, marshaler runtime.Marshaler, server LightUpdateServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq emptypb.Empty
	var metadata runtime.ServerMetadata

	msg, err := server.GetUpdates(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_LightUpdate_GetSkipSyncUpdate_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_LightUpdate_GetSkipSyncUpdate_0(ctx context.Context, marshaler runtime.Marshaler, client LightUpdateClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq SkipSyncRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_LightUpdate_GetSkipSyncUpdate_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetSkipSyncUpdate(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_LightUpdate_GetSkipSyncUpdate_0(ctx context.Context, marshaler runtime.Marshaler, server LightUpdateServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq SkipSyncRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_LightUpdate_GetSkipSyncUpdate_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetSkipSyncUpdate(ctx, &protoReq)
	return msg, metadata, err

}

func request_LightUpdate_DebugGetTrustedCurrentCommitteeRoot_0(ctx context.Context, marshaler runtime.Marshaler, client LightUpdateClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq emptypb.Empty
	var metadata runtime.ServerMetadata

	msg, err := client.DebugGetTrustedCurrentCommitteeRoot(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_LightUpdate_DebugGetTrustedCurrentCommitteeRoot_0(ctx context.Context, marshaler runtime.Marshaler, server LightUpdateServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq emptypb.Empty
	var metadata runtime.ServerMetadata

	msg, err := server.DebugGetTrustedCurrentCommitteeRoot(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterLightUpdateHandlerServer registers the http handlers for service LightUpdate to "mux".
// UnaryRPC     :call LightUpdateServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterLightUpdateHandlerFromEndpoint instead.
func RegisterLightUpdateHandlerServer(ctx context.Context, mux *runtime.ServeMux, server LightUpdateServer) error {

	mux.Handle("GET", pattern_LightUpdate_GetUpdates_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/ethereum.eth.v1alpha1.LightUpdate/GetUpdates")
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_LightUpdate_GetUpdates_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_LightUpdate_GetUpdates_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_LightUpdate_GetSkipSyncUpdate_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/ethereum.eth.v1alpha1.LightUpdate/GetSkipSyncUpdate")
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_LightUpdate_GetSkipSyncUpdate_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_LightUpdate_GetSkipSyncUpdate_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_LightUpdate_DebugGetTrustedCurrentCommitteeRoot_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/ethereum.eth.v1alpha1.LightUpdate/DebugGetTrustedCurrentCommitteeRoot")
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_LightUpdate_DebugGetTrustedCurrentCommitteeRoot_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_LightUpdate_DebugGetTrustedCurrentCommitteeRoot_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterLightUpdateHandlerFromEndpoint is same as RegisterLightUpdateHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterLightUpdateHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterLightUpdateHandler(ctx, mux, conn)
}

// RegisterLightUpdateHandler registers the http handlers for service LightUpdate to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterLightUpdateHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterLightUpdateHandlerClient(ctx, mux, NewLightUpdateClient(conn))
}

// RegisterLightUpdateHandlerClient registers the http handlers for service LightUpdate
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "LightUpdateClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "LightUpdateClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "LightUpdateClient" to call the correct interceptors.
func RegisterLightUpdateHandlerClient(ctx context.Context, mux *runtime.ServeMux, client LightUpdateClient) error {

	mux.Handle("GET", pattern_LightUpdate_GetUpdates_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/ethereum.eth.v1alpha1.LightUpdate/GetUpdates")
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_LightUpdate_GetUpdates_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_LightUpdate_GetUpdates_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_LightUpdate_GetSkipSyncUpdate_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/ethereum.eth.v1alpha1.LightUpdate/GetSkipSyncUpdate")
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_LightUpdate_GetSkipSyncUpdate_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_LightUpdate_GetSkipSyncUpdate_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_LightUpdate_DebugGetTrustedCurrentCommitteeRoot_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/ethereum.eth.v1alpha1.LightUpdate/DebugGetTrustedCurrentCommitteeRoot")
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_LightUpdate_DebugGetTrustedCurrentCommitteeRoot_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_LightUpdate_DebugGetTrustedCurrentCommitteeRoot_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_LightUpdate_GetUpdates_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3}, []string{"eth", "v1alpha1", "light-update", "updates"}, ""))

	pattern_LightUpdate_GetSkipSyncUpdate_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3}, []string{"eth", "v1alpha1", "light-update", "skip-sync"}, ""))

	pattern_LightUpdate_DebugGetTrustedCurrentCommitteeRoot_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3}, []string{"eth", "v1alpha1", "light-update", "debug-get-trusted-sync-comm-root"}, ""))
)

var (
	forward_LightUpdate_GetUpdates_0 = runtime.ForwardResponseMessage

	forward_LightUpdate_GetSkipSyncUpdate_0 = runtime.ForwardResponseMessage

	forward_LightUpdate_DebugGetTrustedCurrentCommitteeRoot_0 = runtime.ForwardResponseMessage
)
