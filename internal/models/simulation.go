package models

import (
	"math"
	"math/rand"
	"time"
)

type Customer struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	AccessToken string `json:"access_token"`
}

type Driver struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	AccessToken string `json:"access_token"`
}

type CommonResponse struct {
	Status  bool        `json:"status"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// FloatBetweenZeroToOne Generate a random float between 0.0 and 1.0
func FloatBetweenZeroToOne() float64 {
	seed := time.Now().UnixNano() + int64(rand.Intn(1000))
	r := rand.New(rand.NewSource(seed))
	return math.Round(r.Float64()*100) / 100
}
