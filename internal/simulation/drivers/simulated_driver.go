package drivers

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
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"googlemaps.github.io/maps"
	pb "sim-server/internal/genserver/proto"

	"sim-server/internal/models"
)

const (
	scheme                  = "wss"
	host                    = "rh-core.advantium.in"
	sleepBeforeArrival      = 10 * time.Second
	sleepPingLocation       = 50 * time.Second
	sleepBeforeStartTrip    = 5 * time.Second
	sleepBeforeCompleteTrip = 5 * time.Second
	sleepForTripPing        = 2 * time.Second
)

type SimulatedDriver struct {
	pb.UnimplementedSimulatedDriverServer
	driver         models.Driver
	lat            float64
	lng            float64
	conn           *websocket.Conn
	tripOfferData  map[string]interface{}
	tripId         string
	acceptanceRate float64
	writeLock      sync.Mutex
}

// Client Methods

func NewSimulatedDriver(driver models.Driver, lat, lng, acceptanceRate float64) {

	sim := &SimulatedDriver{
		driver:         driver,
		lat:            lat,
		lng:            lng,
		acceptanceRate: acceptanceRate,
	}

	sim.serve(driver.Id)
}

func CheckAndGoOnline(driverId string) {
	client, conn, err := client(driverId)
	if err != nil {
		log.Printf("Error connecting to driver: %v", err)
		return
	}
	defer conn.Close()
	client.GoOnline(context.Background(), &pb.GoOnlineRequest{})
}

func Connect(driverId string) {
	client, conn, err := client(driverId)
	if err != nil {
		log.Printf("Error connecting to driver: %v", err)
		return
	}
	defer conn.Close()
	client.InitConnection(context.Background(), &pb.InitConnectionRequest{})
}

func UpdateLocation(driverId string, lat, lng float64) (err error) {
	client, conn, err := client(driverId)
	if err != nil {
		log.Printf("Error connecting to driver: %v", err)
		return
	}
	defer conn.Close()
	_, err = client.SetLocation(context.Background(), &pb.SetLocationRequest{Lat: lat, Lng: lng})
	if err != nil {
		return err
	}
	return nil
}

// Server Methods

func (sim *SimulatedDriver) GoOnline(ctx context.Context, req *pb.GoOnlineRequest) (*pb.GoOnlineResponse, error) {
	// Get token
	token := sim.driver.AccessToken

	commonResponse, err := services.CheckShiftStatus(token)
	if err != nil {
		return &pb.GoOnlineResponse{Success: false}, err
	}
	if commonResponse.Data.(map[string]interface{})["has_active_shift"] == true {
		return &pb.GoOnlineResponse{Success: true}, nil
	}
	commonResponse, err = services.StartNewShift(token)
	if err != nil {
		return &pb.GoOnlineResponse{Success: false}, err
	}
	if commonResponse.Data.(map[string]interface{})["new_shift_started"] == true {
		return &pb.GoOnlineResponse{Success: true}, nil
	} else {
		return &pb.GoOnlineResponse{Success: false}, nil
	}
}

func (sim *SimulatedDriver) InitConnection(ctx context.Context, req *pb.InitConnectionRequest) (*pb.InitConnectionResponse, error) {
	// Get token
	token := sim.driver.AccessToken

	address := url.URL{Scheme: scheme, Host: host, Path: "/ws/driver"}
	conn, _, err := websocket.DefaultDialer.Dial(address.String(), http.Header{
		"Authorization": []string{"Bearer " + token},
		"MRSOOL-CLIENT": []string{"Simulation"},
	})
	if err != nil {
		log.Printf("Error connecting to websocket: %v", err)
		return &pb.InitConnectionResponse{Success: false}, nil
	}
	sim.conn = conn
	log.Print("web socket connected")

	go sim.blockingSubscribe(conn)
	go sim.pingDriverLocationLoop() //pinging location to websocket once the connection gets established
	return &pb.InitConnectionResponse{Success: true}, nil
}

func (sim *SimulatedDriver) SetLocation(ctx context.Context, req *pb.SetLocationRequest) (*pb.SetLocationResponse, error) {
	sim.lat = req.GetLat()
	sim.lng = req.GetLng()
	log.Println(sim.lat, sim.lng)
	payload := models.DriverLocationPayload{
		RawLocation: models.RawLocation{
			Type: "Point",
			Coordinates: models.LatLongCoordinates{
				Latitude:  sim.lat,
				Longitude: sim.lng,
			},
		},
	}
	jsonPayload, _ := json.Marshal(payload)
	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.DriverLocation,
		Payload: jsonPayload,
	})
	if !sim.sendMessageToClient(message) {
		return &pb.SetLocationResponse{Success: false}, nil
	}
	return &pb.SetLocationResponse{Success: true}, nil
}

func (sim *SimulatedDriver) serve(driverId string) {
	//if checkIfAlreadyServed(driverId) {
	//	return
	//}

	// Register and start a gRPC server for the trip
	lis, err := registerService(driverId)
	if err != nil {
		return
	}

	go sim.grpcLoop(lis)
}

func (sim *SimulatedDriver) grpcLoop(lis net.Listener) {
	defer lis.Close()

	s := grpc.NewServer()
	pb.RegisterSimulatedDriverServer(s, sim) // Start with initial state

	if err := s.Serve(lis); err != nil {
		log.Printf("Failed to serve: %v", err)
	}
}

func checkIfAlreadyServed(driverId string) bool {
	value, exists := services.CheckAndGetKey(driverId)

	if !exists {
		return false
	}
	address := value

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to create new client for %s: %v", driverId, err)
	}
	defer conn.Close()

	return true

}

// Utility Methods

func (sim *SimulatedDriver) sendMessageToClient(bytes []byte) bool {
	sim.writeLock.Lock() // Acquire lock before writing
	defer sim.writeLock.Unlock()
	if err := sim.conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
		log.Println("Write error:", err)
		return false
	}

	return true
}

func registerService(driverId string) (lis net.Listener, err error) {
	lis, err = net.Listen("tcp", ":0") // Let the OS provide the port number
	if err != nil {
		log.Printf("Failed to listen: %v", err)
		return
	}

	addr := lis.Addr().(*net.TCPAddr)
	addrString := fmt.Sprintf("%s:%d", "localhost", addr.Port)

	err = services.Set(driverId, []byte(addrString))
	if err != nil {
		log.Printf("Failed to add server to registry: %v", err)
		return
	}
	return
}

func client(driverId string) (pb.SimulatedDriverClient, *grpc.ClientConn, error) {
	value, exists := services.CheckAndGetKey(driverId)

	if !exists {
		return nil, nil, errors.New("server not found for driver " + driverId)
	}

	// Connect to GenServer instance
	genServerConn, err := grpc.NewClient(value, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to GenServer: %v", err)
		return nil, nil, err
	}

	return pb.NewSimulatedDriverClient(genServerConn), genServerConn, err
}

func (sim *SimulatedDriver) blockingSubscribe(conn *websocket.Conn) {
	// Loop over incoming websocket messages
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("driver read:", err)
			return
		}
		log.Printf("driver recv: %s", message)
		var payload map[string]interface{}
		json.Unmarshal(message, &payload)
		switch payload["command"] {
		case string(models.NewTripOffer):
			sim.handleNewTripOffer(payload)
		case string(models.Eta):
			sim.handleEtaPayload(payload)
		case string(models.CompleteTrip):
			sim.handleTripCompletion(payload)
		}
	}
}

// pinging location to websocket once the connection gets established
func (sim *SimulatedDriver) pingDriverLocationLoop() {
	for {
		sim.pingDriverLocation()
		time.Sleep(sleepPingLocation)
	}
}

func (sim *SimulatedDriver) pingDriverLocation() {
	locationPayload := models.DriverLocationPayload{
		RawLocation: models.RawLocation{
			Type: "Point",
			Coordinates: models.LatLongCoordinates{
				Latitude:  sim.lat,
				Longitude: sim.lng,
			},
		},
		VehicleCategoryId: 2,
	}
	jsonPayload, _ := json.Marshal(locationPayload)
	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.DriverLocation,
		Payload: jsonPayload,
	})

	sim.sendMessageToClient(message)
}

func (sim *SimulatedDriver) handleNewTripOffer(payload map[string]interface{}) {
	sim.tripOfferData = payload
	//tripId := sim.tripOfferData["data"].(map[string]interface{})["trip_offer"].(map[string]interface{})["trip_id"]
	if data, ok := payload["data"].(map[string]interface{}); ok {
		if tripOffer, ok := data["trip_offer"].(map[string]interface{}); ok {
			if id, ok := tripOffer["trip_id"].(string); ok {
				fmt.Println("Parsed ID:", id)
				sim.tripId = id
			} else {
				fmt.Println("ID is not a string or not present")
			}
		}
	} else {
		fmt.Println("Data is not a map or not present")
		fmt.Printf(payload["message"].(string))
	}

	if sim.tripId != "" {
		if shouldAccept(sim.acceptanceRate) {
			sim.AcceptTrip(sim.tripId)
		} else {
			sim.RejectTrip(sim.tripId)
		}
	}
}

func (sim *SimulatedDriver) AcceptTrip(tripId string) {
	payload := models.TripActionPayload{
		TripId: tripId,
	}
	jsonPayload, _ := json.Marshal(payload)
	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.AcceptTrip,
		Payload: jsonPayload,
	})
	sim.sendMessageToClient(message)
	sim.tripId = tripId
	// from where to trigger driver arrival
	time.Sleep(sleepBeforeArrival)
	sim.handleDriverArrival()
}

func (sim *SimulatedDriver) RejectTrip(tripId string) {
	payload := models.TripActionPayload{
		TripId: tripId,
	}
	jsonPayload, _ := json.Marshal(payload)
	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.RejectTrip,
		Payload: jsonPayload,
	})
	sim.sendMessageToClient(message)
}

func (sim *SimulatedDriver) handleEtaPayload(payload map[string]interface{}) {
	fmt.Print("Driver getting eta payload after trip acceptance", payload)
}

func (sim *SimulatedDriver) handleDriverArrival() {
	trip := sim.tripOfferData["data"].(map[string]interface{})["trip_offer"].(map[string]interface{})["trip"].(map[string]interface{})
	pickUpPolyline := sim.tripOfferData["data"].(map[string]interface{})["pickup_estimate"].(map[string]interface{})["route"].(map[string]interface{})["polyline"].(map[string]interface{})["encodedPolyline"]
	sim.decodeAndPingOnPolyline(pickUpPolyline.(string))

	sim.lat = trip["origin_lat"].(float64)
	sim.lng = trip["origin_lng"].(float64)
	sim.pingDriverLocation()
	sim.DriverArrival()

	time.Sleep(sleepBeforeStartTrip)
	sim.StartTrip()
	dropPolyline := sim.tripOfferData["data"].(map[string]interface{})["trip_estimate"].(map[string]interface{})["route"].(map[string]interface{})["polyline"].(map[string]interface{})["encodedPolyline"]
	sim.decodeAndPingOnPolyline(dropPolyline.(string))
	sim.lat = trip["destination_lat"].(float64)
	sim.lng = trip["destination_lng"].(float64)
	sim.pingDriverLocation()
	time.Sleep(sleepBeforeCompleteTrip)
	sim.CompleteTrip()
}

func (sim *SimulatedDriver) decodeAndPingOnPolyline(polyline string) {
	coordinates, _ := maps.DecodePolyline(polyline)
	for _, coordinate := range coordinates {
		sim.lat = coordinate.Lat
		sim.lng = coordinate.Lng
		sim.pingDriverLocation()
		time.Sleep(sleepForTripPing)
	}
}

func (sim *SimulatedDriver) DriverArrival() {
	payload := models.TripActionPayload{
		TripId: sim.tripId,
	}
	jsonPayload, _ := json.Marshal(payload)
	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.ArrivedForPickup,
		Payload: jsonPayload,
	})
	sim.sendMessageToClient(message)
}

func (sim *SimulatedDriver) StartTrip() {
	payload := models.TripActionPayload{
		TripId: sim.tripId,
	}
	jsonPayload, _ := json.Marshal(payload)
	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.StartTrip,
		Payload: jsonPayload,
	})
	sim.sendMessageToClient(message)
}

func (sim *SimulatedDriver) CompleteTrip() {
	payload := models.TripActionPayload{
		TripId: sim.tripId,
	}
	jsonPayload, _ := json.Marshal(payload)
	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.CompleteTrip,
		Payload: jsonPayload,
	})
	sim.sendMessageToClient(message)
}

func (sim *SimulatedDriver) handleTripCompletion(_ map[string]interface{}) {
	sim.RateCustomer()
}

func (sim *SimulatedDriver) RateCustomer() {
	payload := models.TripRatingPayload{
		TripId: sim.tripId,
		Rating: models.FloatBetweenZeroToOne() * 5,
	}
	jsonPayload, _ := json.Marshal(payload)
	message, _ := json.Marshal(models.IncomingMessage{
		Command: models.RateCustomer,
		Payload: jsonPayload,
	})
	sim.sendMessageToClient(message)
}

func shouldAccept(probability float64) bool {
	return models.FloatBetweenZeroToOne() < probability
}
