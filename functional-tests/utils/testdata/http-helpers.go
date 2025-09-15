package testdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	FAILED_TO_DECODE_RESPONSE = "failed to decode response: %w"
)

// MakePostRequest sends a POST request with a JSON-encoded request entity to the specified URL,
// and decodes the JSON response into the provided response entity type.
//
// Type Parameters:
//
//	REQ_E - The type of the request entity to be sent in the POST body.
//	RES_E - The type of the response entity to decode the response into.
//
// Parameters:
//
//	url           - The endpoint to send the POST request to.
//	requestEntity - The request payload to be marshaled into JSON.
//
// Returns:
//
//	ResponseResult[RES_E] - A struct containing the HTTP response, status code, and decoded response entity.
//	error                 - An error if the request fails, or if marshaling or decoding fails.
func MakePostRequest[REQ_E any, RES_E any](url string, requestEntity REQ_E) (ResponseResult[RES_E], error) {
	// Marshal the request entity to JSON
	payload, err := json.Marshal(requestEntity)
	if err != nil {
		return ResponseResult[RES_E]{}, fmt.Errorf("failed to marshal request entity: %w", err)
	}

	// Make the POST request
	resp, err := http.Post(url, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		return ResponseResult[RES_E]{}, fmt.Errorf("failed to make POST request: %w", err)
	}
	defer resp.Body.Close()

	// prepare the responseEntity placeholder for decoding
	var re RES_E
	responseEntity := &re
	// Decode the response into responseEntity
	if err := json.NewDecoder(resp.Body).Decode(responseEntity); err != nil {
		return ResponseResult[RES_E]{}, fmt.Errorf(FAILED_TO_DECODE_RESPONSE, err)
	}

	return NewResponseResult(resp, resp.StatusCode, responseEntity), nil

}

// MakeGetRequest sends an HTTP GET request to the specified URL and decodes the response body into the provided generic type RES_E.
// It returns a ResponseResult containing the HTTP response, status code, and the decoded result, or an error if the request or decoding fails.
//
// Type Parameters:
//
//	RES_E - The type into which the response body will be decoded.
//
// Parameters:
//
//	url string - The URL to send the GET request to.
//
// Returns:
//
//	ResponseResult[RES_E] - The result containing the HTTP response and decoded body.
//	error - An error if the request fails or the response cannot be decoded.
func MakeGetRequest[RES_E any](url string) (ResponseResult[RES_E], error) {
	resp, err := http.Get(url)
	if err != nil {
		return ResponseResult[RES_E]{}, err
	}
	defer resp.Body.Close()

	var kanbanBoard RES_E
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&kanbanBoard)
	if err != nil {
		return ResponseResult[RES_E]{}, err
	}
	return NewResponseResult(resp, resp.StatusCode, &kanbanBoard), nil
}

// MakeDeleteRequest sends an HTTP DELETE request to the specified URL and decodes the response body into the provided generic type RES_E.
// It returns a ResponseResult containing the decoded response entity and any error encountered during the request or decoding process.
// If the response body is empty (EOF), it returns a zero value of RES_E without error.
//
// Parameters:
//
//	url string: The URL to which the DELETE request is sent.
//
// Returns:
//
//	ResponseResult[RES_E]: The result containing the HTTP response and decoded entity.
//	error: An error if the request fails or the response cannot be decoded.
func MakeDeleteRequest[RES_E any](url string) (ResponseResult[RES_E], error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return ResponseResult[RES_E]{}, fmt.Errorf("failed to create DELETE request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ResponseResult[RES_E]{}, fmt.Errorf("failed to execute DELETE request: %w", err)
	}
	defer resp.Body.Close()

	var responseEntity RES_E
	if err := json.NewDecoder(resp.Body).Decode(&responseEntity); err != nil && !errors.Is(err, io.EOF) {
		return ResponseResult[RES_E]{}, fmt.Errorf(FAILED_TO_DECODE_RESPONSE, err)
	}

	return NewResponseResult(resp, resp.StatusCode, &responseEntity), nil
}

// MakePutRequest sends an HTTP PUT request to the specified URL with the provided request entity.
// The request entity is marshaled to JSON and included in the request body.
// The response body is decoded into the specified response entity type.
// Returns a ResponseResult containing the response entity and status code, or an error if the request fails.
//
// Type Parameters:
//
//	REQ_E - The type of the request entity to be sent.
//	RES_E - The type of the response entity to be received.
//
// Parameters:
//
//	url           - The endpoint to send the PUT request to.
//	requestEntity - The request payload to be marshaled and sent.
//
// Returns:
//
//	ResponseResult[RES_E] - The result containing the response entity and status code.
//	error                 - An error if the request or decoding fails.
func MakePutRequest[REQ_E any, RES_E any](url string, requestEntity REQ_E) (ResponseResult[RES_E], error) {
	payload, err := json.Marshal(requestEntity)
	if err != nil {
		return ResponseResult[RES_E]{}, fmt.Errorf("failed to marshal request entity: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(string(payload)))
	if err != nil {
		return ResponseResult[RES_E]{}, fmt.Errorf("failed to create PUT request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ResponseResult[RES_E]{}, fmt.Errorf("failed to execute PUT request: %w", err)
	}
	defer resp.Body.Close()

	var responseEntity RES_E
	if err := json.NewDecoder(resp.Body).Decode(&responseEntity); err != nil && !errors.Is(err, io.EOF) {
		return ResponseResult[RES_E]{}, fmt.Errorf(FAILED_TO_DECODE_RESPONSE, err)
	}

	return NewResponseResult(resp, resp.StatusCode, &responseEntity), nil
}
