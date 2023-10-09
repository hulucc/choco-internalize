package nuget

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// PushNupkg PUTs a .nupkg binary to a NuGet Repository
func Push(host string, apiKey string, file io.Reader) error {
	// If no Source provided, exit
	if host == "" {
		return errors.New("Error: Please specify a Source/Host")
	}

	// Create MultiPart Writer
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	// Create new File part
	p, err := w.CreateFormFile("package", "package.nupkg")
	if err != nil {
		return fmt.Errorf("w.CreateFormFile err: %w", err)
	}
	// Write contents to part
	if _, err := io.Copy(p, file); err != nil {
		return fmt.Errorf("io.Copy err: %w", err)
	}
	// Close the writer
	err = w.Close()
	if err != nil {
		return fmt.Errorf("w.Close() err: %w", err)
	}

	// Create new PUT request
	request, err := http.NewRequest(http.MethodPut, host, body)
	if err != nil {
		return fmt.Errorf("http.NewRequest err: %w", err)
	}
	// Add the ApiKey if supplied
	if apiKey != "" {
		request.Header.Add("X-Nuget-Apikey", apiKey)
	}
	// Add the Content Type header from the reader - includes boundary
	request.Header.Add("Content-Type", w.FormDataContentType())

	// Push to the server
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("client.Do err: %w", err)
	}

	if resp.StatusCode > 300 {
		return fmt.Errorf("client.Do response err code %d", resp.StatusCode)
	}
	return nil
}
