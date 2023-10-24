package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/tests"
)

func TestMobileDeviceTestSuite(t *testing.T) {
	suite.Run(t, new(MobileDeviceTestSuite))
}

type MobileDeviceTestSuite struct {
	suite.Suite
}

func (s *MobileDeviceTestSuite) SetupTest() {
	tests.RemoveAllMobileDevices(s.T())
}

func (s *MobileDeviceTestSuite) TestCreateMobileDevice() {
	type testCase struct {
		deviceName       string
		fcmToken         string
		expectedHttpCode int
	}
	defaultFCMToken := "some-fake-token"
	testsCases := []testCase{
		{deviceName: "", fcmToken: defaultFCMToken, expectedHttpCode: 400},
		{deviceName: " ", fcmToken: defaultFCMToken, expectedHttpCode: 400},
		{deviceName: "   ", fcmToken: defaultFCMToken, expectedHttpCode: 400},
		{deviceName: "john`s android", fcmToken: defaultFCMToken, expectedHttpCode: 200},
		{deviceName: "john ", fcmToken: defaultFCMToken, expectedHttpCode: 200},
		{deviceName: " john doe", fcmToken: defaultFCMToken, expectedHttpCode: 200},
		// empty FCM token should be also valid.
		{deviceName: " john doe", fcmToken: "", expectedHttpCode: 200},
	}

	for _, tc := range testsCases {
		response := createDevice(s.T(), tc.deviceName, tc.fcmToken)

		assert.Equal(s.T(), tc.expectedHttpCode, response.StatusCode)
	}
}

func createDevice(t *testing.T, name, fcmToken string) *http.Response {
	payload := []byte(fmt.Sprintf(`{"name":"%s","platform":"android","fcm_token":"%s"}`, name, fcmToken))
	return tests.DoAPIRequest(t, "mobile/devices", http.MethodPost, payload, nil)
}
