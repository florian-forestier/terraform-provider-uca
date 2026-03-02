package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
)

func MakeHTTPRequest[T HasContext](payload any, hasOutput bool, ct T, method string, endpoint string, expectedStatuses []int) (output map[string]string, code int, err error) {

	var bytesPayload []byte

	if payload != nil {
		bytesPayload, err = json.Marshal(payload)
	}

	request, _ := http.NewRequest(method, fmt.Sprintf("%s%s", ct.Context().Endpoint, endpoint), bytes.NewBuffer(bytesPayload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Auth-Token", ct.Context().UserToken)

	response, err := ct.Context().HttpClient.Do(request)
	if err != nil {
		return nil, 0, err
	}

	defer response.Body.Close()

	if !slices.Contains(expectedStatuses, response.StatusCode) {
		return nil, 0, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	if hasOutput {
		body, _ := io.ReadAll(response.Body)
		_ = json.Unmarshal(body, &output)
	}

	return output, response.StatusCode, nil
}
