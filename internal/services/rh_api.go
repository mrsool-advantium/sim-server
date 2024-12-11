package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"sim-server/internal/models"
	"time"
)

const (
	host        = "localhost:8080"
	backdoorOtp = "1234"
)

func DriverLogin(phoneNumber string) (*models.CommonResponse, error) {
	address := url.URL{Scheme: "http", Host: host, Path: "/api/v1/admin/simulate/driver"}
	payload, err := json.Marshal(map[string]interface{}{"phone_number": phoneNumber})
	if err != nil {
		return &models.CommonResponse{}, err
	}
	request, err := http.NewRequest("POST", address.String(), bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Error creating HTTP request:", err)
		return &models.CommonResponse{}, err
	}
	request.Header.Set("MRSOOL-CLIENT", "Simulation")
	client := &http.Client{Timeout: 100 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("Error sending request : %v", err)
		return &models.CommonResponse{}, err
	}
	defer resp.Body.Close()

	//Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return &models.CommonResponse{}, err
	}
	var commonResponse models.CommonResponse
	err = json.Unmarshal(body, &commonResponse)
	if err != nil {
		log.Println("Error parsing response body:", err)
		return &models.CommonResponse{}, err
	}
	return &commonResponse, nil
}

func CustomerLogin(phoneNumber string) (*models.CommonResponse, error) {
	address := url.URL{Scheme: "http", Host: host, Path: "/api/v1/admin/simulate/customer"}
	payload, err := json.Marshal(map[string]interface{}{"phone_number": phoneNumber})
	if err != nil {
		return &models.CommonResponse{}, err
	}
	request, err := http.NewRequest("POST", address.String(), bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Error creating HTTP request:", err)
		return &models.CommonResponse{}, err
	}
	request.Header.Set("MRSOOL-CLIENT", "Simulation")
	client := &http.Client{Timeout: 100 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("Error sending request : %v", err)
		return &models.CommonResponse{}, err
	}
	defer resp.Body.Close()

	//Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return &models.CommonResponse{}, err
	}
	var commonResponse models.CommonResponse
	err = json.Unmarshal(body, &commonResponse)
	if err != nil {
		log.Println("Error parsing response body:", err)
		return &models.CommonResponse{}, err
	}
	return &commonResponse, nil
}

func CheckShiftStatus(token string) (*models.CommonResponse, error) {
	address := url.URL{Scheme: "http", Host: host, Path: "/api/v1/driver/shift_status"}
	request, err := http.NewRequest("GET", address.String(), nil)
	if err != nil {
		log.Println("Error creating HTTP request:", err)
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("MRSOOL-CLIENT", "Simulation")
	client := &http.Client{Timeout: 100 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("Error sending request : %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	//Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, err
	}
	var commonResponse models.CommonResponse
	err = json.Unmarshal(body, &commonResponse)
	if err != nil {
		log.Println("Error parsing response body:", err)
		return nil, err
	}
	if commonResponse.Status {
		return &commonResponse, nil
	} else {
		return nil, errors.New(commonResponse.Message)
	}
}

func StartNewShift(token string) (*models.CommonResponse, error) {
	address := url.URL{Scheme: "http", Host: host, Path: "/api/v1/driver/go_online"}
	request, err := http.NewRequest("POST", address.String(), nil)
	if err != nil {
		log.Println("Error creating HTTP request:", err)
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("MRSOOL-CLIENT", "Simulation")
	client := &http.Client{Timeout: 100 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("Error sending request : %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	//Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, err
	}
	var commonResponse models.CommonResponse
	err = json.Unmarshal(body, &commonResponse)
	if err != nil {
		log.Println("Error parsing response body:", err)
		return nil, err
	}
	if commonResponse.Status {
		return &commonResponse, nil
	} else {
		return nil, errors.New(commonResponse.Message)
	}
}
