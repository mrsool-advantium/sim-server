package customers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sim-server/internal/services"
	"time"

	"sim-server/internal/models"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "sim-server/internal/genserver/proto"
)

const (
	host               = "localhost:8080"
	sleepBeforeLooping = 20 * time.Second
)

type SimulatedCustomer struct {
	pb.UnimplementedSimulatedCustomerServer
	customer            models.Customer
	lat                 float64
	lng                 float64
	originLat           float64
	originLng           float64
	destinationLat      float64
	destinationLng      float64
	loop                bool
	conn                *websocket.Conn
	requestEstimateData map[string]interface{}
	confirmTripData     map[string]interface{}
	tripId              string
}

// Client Methods

func NewSimulatedCustomer(customer models.Customer, loop bool) {
	sim := &SimulatedCustomer{
		customer: customer,
		loop:     loop,
	}

	sim.serve(customer.Id)
}

func Connect(customerId string) {
	client, conn, err := client(customerId)
	if err != nil {
		log.Printf("Error connecting to customer: %v", err)
		return
	}
	defer conn.Close()
	client.InitConnection(context.Background(), &pb.InitConnectionRequest{})
}

func UpdateLocation(customerId string, lat float64, lng float64) {
	client, conn, err := client(customerId)
	if err != nil {
		log.Printf("Error connecting to customer: %v", err)
		return
	}
	defer conn.Close()
	client.SetLocation(context.Background(), &pb.SetLocationRequest{Lat: lat, Lng: lng})
}

func RequestEstimate(customerId string, originLat, originLng, destinationLat, destinationLng float64) {
	client, conn, err := client(customerId)
	if err != nil {
		log.Printf("Error connecting to customer: %v", err)
		return
	}
	defer conn.Close()
	client.TripEstimate(context.Background(),
		&pb.TripEstimateRequest{
			OriginLng:      originLng,
			OriginLat:      originLat,
			DestinationLat: destinationLat,
			DestinationLng: destinationLng,
		})
}

func ConfirmTrip(customerId string, originLat, originLng, destinationLat, destinationLng float64) {
	client, conn, err := client(customerId)
	if err != nil {
		log.Printf("Error connecting to customer: %v", err)
		return
	}
	defer conn.Close()
	confirmTrip, err := client.ConfirmTrip(context.Background(),
		&pb.ConfirmTripRequest{
			OriginLat:      originLat,
			OriginLng:      originLng,
			DestinationLat: destinationLat,
			DestinationLng: destinationLng,
		})
	if err != nil {
		return
	}
	fmt.Println(confirmTrip)
}

// Server Methods

func (sim *SimulatedCustomer) InitConnection(ctx context.Context, req *pb.InitConnectionRequest) (*pb.InitConnectionResponse, error) {
	// Get token
	token := sim.customer.AccessToken

	address := url.URL{Scheme: "wss", Host: host, Path: "/ws/customer"}
	conn, _, err := websocket.DefaultDialer.Dial(address.String(), http.Header{
		"Authorization": []string{"Bearer " + token},
		"MRSOOL-CLIENT": []string{"Simulation"},
	})
	if err != nil {
		log.Printf("Error connecting to websocket: %v", err)
		return &pb.InitConnectionResponse{Success: false}, nil
	}
	sim.conn = conn

	go sim.blockingSubscribe(conn)
	return &pb.InitConnectionResponse{Success: true}, nil
}

func (sim *SimulatedCustomer) SetLocation(ctx context.Context, req *pb.SetLocationRequest) (*pb.SetLocationResponse, error) {
	sim.lat = req.GetLat()
	sim.lng = req.GetLng()
	return &pb.SetLocationResponse{Success: true}, nil
}

func (sim *SimulatedCustomer) TripEstimate(ctx context.Context, req *pb.TripEstimateRequest) (*pb.TripEstimateResponse, error) {
	tripRequestPayload := models.TripRequestPayload{
		Origin: models.LatLong{
			Latitude:  req.GetOriginLat(),
			Longitude: req.GetOriginLng(),
		},
		Destination: models.LatLong{
			Latitude:  req.GetDestinationLat(),
			Longitude: req.GetDestinationLng(),
		},
	}
	jsonPayload, _ := json.Marshal(tripRequestPayload)

	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.RequestEstimate,
		Payload: jsonPayload,
	})
	if err := sim.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		log.Println("Write error:", err)
		return &pb.TripEstimateResponse{Success: false}, err
	}
	return &pb.TripEstimateResponse{Success: true}, nil
}

func (sim *SimulatedCustomer) ConfirmTrip(ctx context.Context, req *pb.ConfirmTripRequest) (*pb.ConfirmTripResponse, error) {
	sim.originLat = req.GetOriginLat()
	sim.originLng = req.GetOriginLng()
	sim.destinationLat = req.GetDestinationLat()
	sim.destinationLng = req.GetDestinationLng()
	tripRequestPayload := models.TripRequestPayload{
		Origin: models.LatLong{
			Latitude:  sim.originLat,
			Longitude: sim.originLng,
		},
		Destination: models.LatLong{
			Latitude:  sim.destinationLat,
			Longitude: sim.destinationLng,
		},
		VehicleCategoryId: 1, //default selecting the category
	}
	jsonPayload, _ := json.Marshal(tripRequestPayload)

	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.ConfirmTrip,
		Payload: jsonPayload,
	})
	if err := sim.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		log.Println("Write error:", err)
		return &pb.ConfirmTripResponse{Success: false}, err
	}
	return &pb.ConfirmTripResponse{Success: true}, nil
}

// Utility Methods

func (sim *SimulatedCustomer) serve(customerId string) {
	//if checkIfAlreadyServed(customerId) {
	//	return
	//}
	// Register and start a gRPC server for the trip
	lis, err := registerService(customerId)
	if err != nil {
		return
	}

	go sim.grpcLoop(lis)
}

func (sim *SimulatedCustomer) grpcLoop(lis net.Listener) {
	defer lis.Close()

	s := grpc.NewServer()
	pb.RegisterSimulatedCustomerServer(s, sim) // Start with initial state

	if err := s.Serve(lis); err != nil {
		log.Printf("Failed to serve: %v", err)
	}
}

func checkIfAlreadyServed(customerId string) bool {
	value, exists := services.CheckAndGetKey(customerId)

	if !exists {
		return false
	}
	address := value
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to create new client for %s: %v", customerId, err)
	}
	defer conn.Close()

	return true

}

func registerService(customerId string) (lis net.Listener, err error) {
	lis, err = net.Listen("tcp", ":0") // Let the OS provide the port number
	if err != nil {
		log.Printf("Failed to listen: %v", err)
		return
	}

	addr := lis.Addr().(*net.TCPAddr)
	addrString := fmt.Sprintf("%s:%d", "localhost", addr.Port)

	err = services.Set(customerId, []byte(addrString))
	if err != nil {
		log.Printf("Failed to add server to registry: %v", err)
		return
	}
	return
}

func client(customerId string) (pb.SimulatedCustomerClient, *grpc.ClientConn, error) {

	value, exists := services.CheckAndGetKey(customerId)

	if !exists {
		return nil, nil, errors.New("server not found for customer " + customerId)
	}

	// Connect to GenServer instance
	genServerConn, err := grpc.NewClient(value, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to GenServer: %v", err)
		return nil, nil, err
	}

	return pb.NewSimulatedCustomerClient(genServerConn), genServerConn, err
}

func (sim *SimulatedCustomer) blockingSubscribe(conn *websocket.Conn) {
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)
		var payload map[string]interface{}
		json.Unmarshal(message, &payload)
		switch payload["command"] {
		case string(models.RequestEstimate):
			sim.handleRequestEstimate(payload)
		case string(models.ConfirmTrip):
			sim.handleConfirmTrip(payload)
		case string(models.Eta):
			sim.handleEtaPayload(payload)
		case string(models.DriverLocation):
			sim.handleDriverLocation(payload)
		case string(models.CompleteTrip):
			sim.handleTripCompletion(payload)
		}
	}
}

func (sim *SimulatedCustomer) handleRequestEstimate(payload map[string]interface{}) {
	sim.requestEstimateData = payload
}

func (sim *SimulatedCustomer) handleConfirmTrip(payload map[string]interface{}) {
	sim.confirmTripData = payload
	if data, ok := payload["data"].(map[string]interface{}); ok {
		if id, ok := data["id"].(string); ok {
			fmt.Println("Parsed ID:", id)
			sim.tripId = id
		} else {
			fmt.Println("ID is not a string or not present")
		}
	} else {
		fmt.Println("Data is not a map or not present")
		fmt.Printf(payload["message"].(string))
	}
}

func (sim *SimulatedCustomer) handleEtaPayload(payload map[string]interface{}) {
	fmt.Print("Customer getting eta payload after trip acceptance", payload)
}

func (sim *SimulatedCustomer) handleDriverLocation(payload map[string]interface{}) {
	fmt.Print("Customer getting driver current location trip acceptance", payload)
}

func (sim *SimulatedCustomer) handleTripCompletion(_ map[string]interface{}) {
	sim.RateDriver()
	if sim.loop {
		time.Sleep(sleepBeforeLooping)
		ConfirmTrip(sim.customer.Id, sim.destinationLat, sim.destinationLng, sim.originLat, sim.originLng)
	}
}

func (sim *SimulatedCustomer) RateDriver() {
	payload := models.TripRatingPayload{
		TripId: sim.tripId,
		Rating: models.FloatBetweenZeroToOne() * 5,
	}
	jsonPayload, _ := json.Marshal(payload)
	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.RateDriver,
		Payload: jsonPayload,
	})
	if err := sim.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		log.Println("Write error:", err)
		return
	}
}
