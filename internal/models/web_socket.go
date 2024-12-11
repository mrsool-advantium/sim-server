package models

import (
	"encoding/json"
)

type Command string

const (
	GoOnline             Command = "goOnline"
	GoOffline            Command = "goOffline"
	RequestEstimate      Command = "requestEstimate"
	ConfirmTrip          Command = "confirmTrip"
	DriverLocation       Command = "driverLocation"
	AcceptTrip           Command = "acceptTrip"
	RejectTrip           Command = "rejectTrip"
	ArrivedForPickup     Command = "arrivedForPickup"
	StartTrip            Command = "startTrip"
	CompleteTrip         Command = "completeTrip"
	Ack                  Command = "ack"
	NewTripOffer         Command = "newTripOffer"
	NoDriverFound        Command = "noDriverFound"
	NoDriverAcceptedTrip Command = "noDriverAcceptedTrip"
	CancelTrip           Command = "cancelTrip"
	CancellationReasons  Command = "cancellationReasons"
	RateDriver           Command = "rateDriver"
	RateCustomer         Command = "rateCustomer"
	LocationUpdate       Command = "locationUpdate"
	Sync                 Command = "sync"
	Eta                  Command = "eta"
	Reroute              Command = "reroute"
	TripTimedOut         Command = "tripTimedOut"
)

// LATEST AND NEW CODE

// IncomingMessage represents to fetch only the command from the incoming.
type IncomingMessage struct {
	Command Command         `json:"command,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

// DriverLocationPayload represents the incoming location data from drivers.
type DriverLocationPayload struct {
	UserType          string      `json:"userType,omitempty"`
	RideType          string      `json:"rideType,omitempty"`
	RawLocation       RawLocation `json:"rawLocation,omitempty"`
	Accuracy          float64     `json:"accuracy,omitempty"`
	Altitude          float64     `json:"altitude,omitempty"`
	DriverID          string      `json:"driverId,omitempty"`
	CustomerID        string      `json:"customerId,omitempty"`
	TripId            string      `json:"tripId,omitempty"`
	Floor             float64     `json:"floor,omitempty"`
	Source            string      `json:"source,omitempty"`
	Speed             float64     `json:"speed,omitempty"`
	SpeedAccuracy     float64     `json:"speedAccuracy,omitempty"`
	H3Cells           []H3Cell    `json:"h3Cells,omitempty"`
	Heading           float64     `json:"heading,omitempty"`
	Timestamp         int64       `json:"timestamp,omitempty"`
	VehicleCategoryId int         `json:"vehicleCategoryId,omitempty"`
}

// RawLocation represents the geographical coordinates.
type RawLocation struct {
	Type        string             `json:"type,omitempty"`
	Coordinates LatLongCoordinates `json:"coordinates,omitempty"`
}

// LatLongCoordinates represents the geographical coordinates.
type LatLongCoordinates struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

// H3Cell represents the geographical coordinates.
type H3Cell struct {
	Resolution int    `json:"resolution,omitempty"`
	CellId     string `json:"cellId,omitempty"`
}

// TripRequestPayload payload and ConfirmTripPayload would be the same
type TripRequestPayload struct {
	Origin            LatLong `json:"origin" validate:"required"`
	OriginName        string  `json:"origin_name"`
	Destination       LatLong `json:"destination" validate:"required"`
	DestinationName   string  `json:"destination_name"`
	VehicleCategoryId int     `json:"category_id,omitempty"`
}

type LatLong struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CancelTripPayload struct {
	TripId   string `json:"trip_id" validate:"required"`
	ReasonId int    `json:"reason_id" validate:"required"`
}

type TripRatingPayload struct {
	TripId string  `json:"trip_id" validate:"required"`
	Rating float64 `json:"rating" validate:"required"`
}

// TripActionPayload represents accepted trip message by the driver for a trip offering id
type TripActionPayload struct {
	TripId string `json:"trip_id" validate:"required"`
}
