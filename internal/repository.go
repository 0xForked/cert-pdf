package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Repository struct {
	BaseURL string
}

func (r Repository) LoadCert(
	ctx context.Context,
	key, value string,
) (interface{}, error) {
	endpoint := fmt.Sprintf("%s/api/v1/certificates/%s/%s", r.BaseURL, value, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{"Content-Type": []string{"application/json"}}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, resp.Body); err != nil {
		return nil, err
	}
	var res, returnValue interface{}
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		return nil, err
	}
	if rsp, ok := res.(map[string]interface{}); ok {
		// check if error
		if rsp["error"] != nil && rsp["error"].(bool) {
			return nil, errors.New(rsp["data"].(string))
		}
		// return value
		returnValue = rsp["data"]
	}
	return returnValue, nil
}

func (r Repository) ToCertStruct(
	cert *PreGenerateCertificate,
	raw interface{},
) error {
	// marshal the data into raw json
	data, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	// unmarshal the raw json into the response struct
	return json.Unmarshal(data, &cert)
}
