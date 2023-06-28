package web

import (
	"authex/model"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

type payloadTest struct {
}

func (payloadTest) Serialize() ([]byte, error) {
	return []byte{0}, nil
}

// TestPostOrder tests the postOrder endpoint
func TestPostOrder(t *testing.T) {

	// create a test order and sign it
	order := &model.SignedRequest[model.Order]{}
	order.From = "test From"

	// create a test request with the signed order
	req := &model.SignedRequest[model.Order]{
		From:      order.From,
		Payload:   model.Order{},
		Signature: order.Signature,
	}
	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	// create a test URL for the endpoint
	url := fmt.Sprintf("%s/order", "server/test")

	// make a POST request to the endpoint
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	// check the status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	// unmarshal the response into a map
	respMap := make(map[string]interface{})
	err = json.Unmarshal(respBody, &respMap)
	require.NoError(t, err)

	// check the response fields
	assert.Equal(t, "ok", respMap["status"])
	assert.NotEmpty(t, respMap["incident"])
	assert.NotEmpty(t, respMap["data"])
	dataMap := respMap["data"].(map[string]interface{})
	assert.Equal(t, order.Payload.ID, dataMap["order_id"])
}
