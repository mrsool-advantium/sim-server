package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sim-server/internal/models"
	"sim-server/internal/services"
	"sim-server/internal/simulation/customers"
	"sim-server/internal/simulation/drivers"
	"strconv"
	"time"
)

type SimHandler struct {
}

func (handler SimHandler) SimulateScenario(context *gin.Context) {
	type request struct {
		NumDrivers          int     `json:"num_drivers"`
		NumCustomers        int     `json:"num_customers"`
		CenterLat           float64 `json:"center_lat"`
		CenterLng           float64 `json:"center_lng"`
		Radius              float64 `json:"radius"`
		Loop                bool    `json:"loop"`
		AcceptanceRate      float64 `json:"acceptance_rate"`
		DriverSeriesStart   int     `json:"driver_series_start"`
		CustomerSeriesStart int     `json:"customer_series_start"`
	}

	var req request
	err := context.ShouldBindJSON(&req)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Initial coordinates
	lat := req.CenterLat        // CP Lat
	lng := req.CenterLng        // CP Lng
	radius := req.Radius * 1000 // Radius in meters

	for i := 1; i <= req.NumDrivers; i++ {
		simSeriesNumbers := 1111100000 + req.DriverSeriesStart
		phoneNumber := simSeriesNumbers + i

		// Generate random point
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		newLat, newLng := generateRandomPoint(lat, lng, float64(radius), rng)
		go simNewDriver(phoneNumber, newLat, newLng, req.AcceptanceRate)
	}

	for i := 1; i <= req.NumCustomers; i++ {
		simSeriesNumbers := 1111100000 + req.CustomerSeriesStart
		phoneNumber := simSeriesNumbers + i

		// Generate random point
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		orgLat, orgLng := generateRandomPoint(lat, lng, float64(radius), rng)

		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
		desLat, desLng := generateRandomPoint(lat, lng, float64(radius)*4, rng)

		go simNewCustomer(phoneNumber, req.Loop, orgLat, orgLng, desLat, desLng)
	}

}

func simulateDriver(driver *models.Driver, lat, lng, acceptanceRate float64) {
	drivers.NewSimulatedDriver(*driver, lat, lng, acceptanceRate)
	drivers.CheckAndGoOnline(driver.Id)
	drivers.Connect(driver.Id)
}

func simulateCustomer(customer *models.Customer, loop bool) {
	customers.NewSimulatedCustomer(*customer, loop)
	customers.Connect(customer.Id)
}

func simNewCustomer(phoneNumber int, loop bool, orgLat, orgLng, desLat, desLng float64) {
	response, err := services.CustomerLogin(strconv.Itoa(phoneNumber))
	if err != nil {
		log.Printf("error logging in: %v", err)
		return
	}
	if response.Status {
		var customer models.Customer
		customer.Id = response.Data.(map[string]interface{})["id"].(string)
		customer.Name = response.Data.(map[string]interface{})["name"].(string)
		customer.PhoneNumber = response.Data.(map[string]interface{})["phone_number"].(string)
		customer.AccessToken = response.Data.(map[string]interface{})["access_token"].(string)
		simulateCustomer(&customer, loop)
		customers.ConfirmTrip(customer.Id, orgLat, orgLng, desLat, desLng)
	} else {
		log.Printf("error: %v", response.Message)
	}
}

func simNewDriver(phoneNumber int, lat, lng, acceptanceRate float64) {
	response, err := services.DriverLogin(strconv.Itoa(phoneNumber))
	if err != nil {
		log.Printf("error logging in: %v", err)
		return
	}
	log.Printf("response: %v", response)
	if response.Status {
		var driver models.Driver
		driver.Id = response.Data.(map[string]interface{})["id"].(string)
		driver.Name = response.Data.(map[string]interface{})["name"].(string)
		driver.PhoneNumber = response.Data.(map[string]interface{})["phone_number"].(string)
		driver.AccessToken = response.Data.(map[string]interface{})["access_token"].(string)
		simulateDriver(&driver, lat, lng, acceptanceRate)
	} else {
		log.Printf("error: %v", response.Message)
	}
}

const EarthRadius = 6371000.0 // Earth's radius in meters

// toRadians converts degrees to radians
func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

// toDegrees converts radians to degrees
func toDegrees(rad float64) float64 {
	return rad * 180 / math.Pi
}

// generateRandomPoint generates a random latitude and longitude within a radius from a given point
func generateRandomPoint(lat, lng, radius float64, rng *rand.Rand) (float64, float64) {
	// Convert latitude and longitude from degrees to radians
	latRad := toRadians(lat)
	lngRad := toRadians(lng)

	// Random distance in meters within the radius
	distance := rng.Float64() * radius

	// Random bearing (angle in radians)
	bearing := rng.Float64() * 2 * math.Pi

	// New latitude in radians
	newLatRad := math.Asin(math.Sin(latRad)*math.Cos(distance/EarthRadius) +
		math.Cos(latRad)*math.Sin(distance/EarthRadius)*math.Cos(bearing))

	// New longitude in radians
	newLngRad := lngRad + math.Atan2(
		math.Sin(bearing)*math.Sin(distance/EarthRadius)*math.Cos(latRad),
		math.Cos(distance/EarthRadius)-math.Sin(latRad)*math.Sin(newLatRad))

	// Convert new coordinates back to degrees
	newLat := toDegrees(newLatRad)
	newLng := toDegrees(newLngRad)

	return newLat, newLng

	//28.632837,77.219567
}
