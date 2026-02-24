package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"tempconv/gen/tempconvpb"
)

type server struct {
	tempconvpb.UnimplementedTempConverterServer
}

func convert(value float64, fromUnit tempconvpb.TemperatureUnit, toUnit tempconvpb.TemperatureUnit) (float64, string, error) {
	if fromUnit == toUnit {
		return value, "same unit, no conversion", nil
	}

	celsiusValue := value

	switch fromUnit {
	case tempconvpb.TemperatureUnit_CELSIUS:
	case tempconvpb.TemperatureUnit_FAHRENHEIT:
		celsiusValue = (value - 32) * 5 / 9
	case tempconvpb.TemperatureUnit_KELVIN:
		celsiusValue = value - 273.15
	default:
		return 0, "", fmt.Errorf("unsupported source unit")
	}

	switch toUnit {
	case tempconvpb.TemperatureUnit_CELSIUS:
		return celsiusValue, "C = source converted to celsius", nil
	case tempconvpb.TemperatureUnit_FAHRENHEIT:
		return (celsiusValue*9/5 + 32), "F = (C * 9/5) + 32", nil
	case tempconvpb.TemperatureUnit_KELVIN:
		return (celsiusValue + 273.15), "K = C + 273.15", nil
	default:
		return 0, "", fmt.Errorf("unsupported destination unit")
	}
}

func (s *server) ConvertTemperature(_ context.Context, request *tempconvpb.ConvertRequest) (*tempconvpb.ConvertResponse, error) {
	result, formula, err := convert(request.GetValue(), request.GetFromUnit(), request.GetToUnit())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &tempconvpb.ConvertResponse{
		ConvertedValue: result,
		FormulaUsed:    formula,
	}, nil
}

func main() {
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	tempconvpb.RegisterTempConverterServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	go func() {
		log.Printf("TempConverter gRPC server running on port %s", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	gatewayMux := runtime.NewServeMux()
	ctx := context.Background()
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := tempconvpb.RegisterTempConverterHandlerFromEndpoint(ctx, gatewayMux, "127.0.0.1:"+grpcPort, dialOptions); err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	log.Printf("TempConverter REST gateway running on port %s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, gatewayMux); err != nil {
		log.Fatalf("failed to serve gateway: %v", err)
	}
}
