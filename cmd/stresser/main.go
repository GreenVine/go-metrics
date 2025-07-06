package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/greenvine/go-metrics/proto/gen/device/v1"
)

const logPrefix = "[metrics-stresser] "

type config struct {
	Server                  string
	TotalDevices            int
	MinMetricInterval       time.Duration
	MaxMetricInterval       time.Duration
	MinTemperatureThreshold float64
	MaxTemperatureThreshold float64
	MinBatteryThreshold     int
	MaxBatteryThreshold     int
	MinTemperatureMetric    float64
	MaxTemperatureMetric    float64
	MinBatteryMetric        int
	MaxBatteryMetric        int
	MetricsPerDevice        int
}

type deviceConfig struct {
	ID                   uuid.UUID
	ResourceName         string
	TemperatureThreshold float32
	BatteryThreshold     int32
}

func parseFlags() *config {
	cfg := &config{}

	flag.StringVar(&cfg.Server, "server", "127.0.0.1:8080", "Address of the metrics server")
	flag.IntVar(&cfg.TotalDevices, "totalDevices", 1000, "Number of devices to be simulated")
	flag.DurationVar(&cfg.MinMetricInterval, "minMetricInterval", 1*time.Second, "Minimum interval before each metric emission attempt")
	flag.DurationVar(&cfg.MaxMetricInterval, "maxMetricInterval", 5*time.Second, "Maximum interval before each metric emission attempt")
	flag.Float64Var(&cfg.MinTemperatureThreshold, "minTemperatureThreshold", 20, "Minimum temperature threshold for the device")
	flag.Float64Var(&cfg.MaxTemperatureThreshold, "maxTemperatureThreshold", 50, "Maximum temperature threshold for the device")
	flag.IntVar(&cfg.MinBatteryThreshold, "minBatteryThreshold", 30, "Minimum battery threshold for the device")
	flag.IntVar(&cfg.MaxBatteryThreshold, "maxBatteryThreshold", 70, "Maximum battery threshold for the device")
	flag.Float64Var(&cfg.MinTemperatureMetric, "minTemperatureMetric", 0, "Minimum temperature metric emitted for the device")
	flag.Float64Var(&cfg.MaxTemperatureMetric, "maxTemperatureMetric", 70, "Maximum temperature metric emitted for the device")
	flag.IntVar(&cfg.MinBatteryMetric, "minBatteryMetric", 0, "Minimum battery metric emitted for the device")
	flag.IntVar(&cfg.MaxBatteryMetric, "maxBatteryMetric", 100, "Maximum battery metric emitted for the device")
	flag.IntVar(&cfg.MetricsPerDevice, "metricsPerDevice", 10, "Number of metrics to be emitted per device")

	flag.Parse()

	return cfg
}

func randomInterval(min, max time.Duration) time.Duration {
	return min + time.Duration(rand.Int63n(int64(max-min)))
}

func randomFloat(min, max float64) float32 {
	return float32(min + rand.Float64()*(max-min))
}

func randomInt(min, max int) int32 {
	return int32(min + rand.Intn(max-min+1))
}

func buildDeviceResourceName(deviceID uuid.UUID) string {
	return fmt.Sprintf("devices/%s", deviceID.String())
}

func main() {
	log.SetPrefix(logPrefix)

	cfg := parseFlags()
	log.Printf("Starting stresser with %d devices against server %s", cfg.TotalDevices, cfg.Server)

	// Connect to the gRPC server
	conn, clientErr := grpc.NewClient(cfg.Server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if clientErr != nil {
		log.Fatalf("Failed to connect to server: %v", clientErr)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}(conn)

	deviceMgmtClient := devicev1.NewDeviceMgmtServiceClient(conn)

	log.Println("Configuring devices...")

	deviceConfigs := make([]*deviceConfig, 0, cfg.TotalDevices)
	for i := 0; i < cfg.TotalDevices; i++ {
		deviceID := uuid.New()
		deviceResourceName := buildDeviceResourceName(deviceID)

		// Randomise device thresholds
		temperatureThreshold := randomFloat(cfg.MinTemperatureThreshold, cfg.MaxTemperatureThreshold)
		batteryThreshold := randomInt(cfg.MinBatteryThreshold, cfg.MaxBatteryThreshold)

		// Create the device configuration synchronously
		if _, err := deviceMgmtClient.UpsertConfig(context.Background(), &devicev1.UpsertConfigRequest{
			Parent: &deviceResourceName,
			Config: &devicev1.Config{
				TemperatureThreshold: &temperatureThreshold,
				BatteryThreshold:     &batteryThreshold,
			},
		}); err != nil {
			log.Printf("(#%d - %s) Failed to configure: %v", i, deviceResourceName, err)
			continue
		}

		// Store successfully configured device
		deviceConfigs = append(deviceConfigs, &deviceConfig{
			ID:                   deviceID,
			ResourceName:         deviceResourceName,
			TemperatureThreshold: temperatureThreshold,
			BatteryThreshold:     batteryThreshold,
		})
	}

	log.Printf("Successfully configured %d devices", len(deviceConfigs))
	log.Println("Starting emitting metrics...")

	// Now emit metrics in parallel for all configured devices
	var wg sync.WaitGroup
	for i, device := range deviceConfigs {
		wg.Add(1)
		go func(deviceInd int, dev *deviceConfig) {
			defer wg.Done()

			// Emit fake metrics
			for j := 0; j < cfg.MetricsPerDevice; j++ {
				interval := randomInterval(cfg.MinMetricInterval, cfg.MaxMetricInterval)
				time.Sleep(interval)

				// Create random metric values
				temperature := randomFloat(cfg.MinTemperatureMetric, cfg.MaxTemperatureMetric)
				battery := randomInt(cfg.MinBatteryMetric, cfg.MaxBatteryMetric)

				if _, err := deviceMgmtClient.CreateMetric(context.Background(), &devicev1.CreateMetricRequest{
					Parent: &dev.ResourceName,
					Metric: &devicev1.Metric{
						Temperature: &temperature,
						Battery:     &battery,
					},
				}); err != nil {
					log.Printf("(#%d - %s) Failed to emit metric %d: %v",
						deviceInd, dev.ResourceName, j, err)
				}
			}
		}(i, device)
	}

	wg.Wait()

	log.Println("Stressing test done.")
}
