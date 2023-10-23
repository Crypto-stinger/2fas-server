package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/tests"
)

func TestMobileDeviceExtensionIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(MobileDeviceExtensionIntegrationTestSuite))
}

type MobileDeviceExtensionIntegrationTestSuite struct {
	suite.Suite
}

func (s *MobileDeviceExtensionIntegrationTestSuite) SetupTest() {
	tests.RemoveAllMobileDevices(s.T())
	tests.RemoveAllBrowserExtensions(s.T())
	tests.RemoveAllBrowserExtensionsDevices(s.T())
}

func (s *MobileDeviceExtensionIntegrationTestSuite) TestGetPending2FaRequests() {
	device, devicePubKey := tests.CreateDevice(s.T(), "SM-955F", "fcm-token")
	browserExtension := tests.CreateBrowserExtension(s.T(), "go-ext")
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExtension, device)

	var tokenRequest *tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"domain.com"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	var tokenRequestsCollection []*tests.AuthTokenRequestResponse
	tests.DoAPISuccessGet(s.T(), "mobile/devices/"+device.Id+"/browser_extensions/2fa_requests", &tokenRequestsCollection)
	assert.Len(s.T(), tokenRequestsCollection, 1)
}

func (s *MobileDeviceExtensionIntegrationTestSuite) TestDoNotReturnCompleted2FaRequests() {
	device, devicePubKey := tests.CreateDevice(s.T(), "SM-955F", "fcm-token")
	browserExtension := tests.CreateBrowserExtension(s.T(), "go-ext")
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExtension, device)

	var tokenRequest *tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"domain.com"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var tokenRequestsCollection []*tests.AuthTokenRequestResponse
	tests.DoAPISuccessGet(s.T(), "mobile/devices/"+device.Id+"/browser_extensions/2fa_requests", &tokenRequestsCollection)
	assert.Len(s.T(), tokenRequestsCollection, 0)
}
