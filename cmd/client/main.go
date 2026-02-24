package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"tempconv/gen/tempconvpb"
)

func parseUnit(value string) (tempconvpb.TemperatureUnit, error) {
	switch strings.ToUpper(value) {
	case "C":
		return tempconvpb.TemperatureUnit_CELSIUS, nil
	case "F":
		return tempconvpb.TemperatureUnit_FAHRENHEIT, nil
	case "K":
		return tempconvpb.TemperatureUnit_KELVIN, nil
	default:
		return tempconvpb.TemperatureUnit_TEMPERATURE_UNIT_UNSPECIFIED, fmt.Errorf("invalid unit: %s", value)
	}
}

func main() {
	host := flag.String("host", "localhost:50051", "gRPC server host:port")
	value := flag.Float64("value", 0, "input temperature value")
	fromUnitText := flag.String("from", "C", "source unit: C, F, K")
	toUnitText := flag.String("to", "F", "destination unit: C, F, K")
	flag.Parse()

	fromUnit, err := parseUnit(*fromUnitText)
	if err != nil {
		log.Fatal(err)
	}

	toUnit, err := parseUnit(*toUnitText)
	if err != nil {
		log.Fatal(err)
	}

	connection, err := grpc.Dial(*host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer connection.Close()

	client := tempconvpb.NewTempConverterClient(connection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.ConvertTemperature(ctx, &tempconvpb.ConvertRequest{
		Value:    *value,
		FromUnit: fromUnit,
		ToUnit:   toUnit,
	})
	if err != nil {
		log.Fatalf("RPC failed: %v", err)
	}

	fmt.Printf("Converted value: %.4f\n", response.GetConvertedValue())
	fmt.Printf("Formula: %s\n", response.GetFormulaUsed())
}
