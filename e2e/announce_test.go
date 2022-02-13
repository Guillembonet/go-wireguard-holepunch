//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/guillembonet/go-wireguard-holepunch/e2e/common/server/endpoints"
	"github.com/stretchr/testify/assert"
)

func TestAnnounce(t *testing.T) {

	t.Run("announces", func(t *testing.T) {
		err := sendAnnounceRequest(t, CLIENT0_BASE_URL, &endpoints.AnnounceReq{
			ServerIp:      "10.100.1.4",
			ServerPort:    2001,
			InterfaceCIDR: "15.14.13.12/24",
		})
		assert.Nil(t, err)
		resp, err := sendListRequest(t, SERVER_BASE_URL)
		assert.Nil(t, err)
		assert.Len(t, *resp, 1)
		assert.Equal(t, "15.14.13.12/24", (*resp)[0].CIDR)
	})
}

func sendAnnounceRequest(t *testing.T, baseurl string, request *endpoints.AnnounceReq) *endpoints.Error {
	url := baseurl + "/announce"

	blob, err := json.Marshal(request)
	assert.NoError(t, err)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(blob))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	return parseError(t, resp)
}

func sendListRequest(t *testing.T, baseurl string) (*endpoints.ListResp, *endpoints.Error) {
	url := baseurl + "/list"

	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	var lr endpoints.ListResp
	return &lr, parseResp(t, resp, &lr)
}

func parseResp(t *testing.T, resp *http.Response, obj interface{}) *endpoints.Error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseError(t, resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(body, &obj)
	assert.NoError(t, err)

	return nil
}

func parseError(t *testing.T, resp *http.Response) *endpoints.Error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var result endpoints.Error
		err := json.NewDecoder(resp.Body).Decode(&result)
		result.Code = resp.StatusCode
		assert.NoError(t, err)
		return &result
	}

	return nil
}
