package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
		return (celsiusValue*9/5 + 32), "F = (C × 9/5) + 32", nil
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	tempconvpb.RegisterTempConverterServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	log.Printf("TempConverter gRPC server running on port %s", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
