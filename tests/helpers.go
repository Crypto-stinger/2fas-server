package tests

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/twofas/2fas-server/internal/common/crypto"
)

func CreateDevice(t *testing.T, name, fcmToken string) (*DeviceResponse, string) {
	keyPair := crypto.GenerateKeyPair(2048)
	devicePubKey := crypto.PublicKeyToBase64(keyPair.PublicKey)

	payload := []byte(fmt.Sprintf(`{"name":"%s","platform":"android","fcm_token":"%s"}`, name, fcmToken))

	device := new(DeviceResponse)

	DoAPISuccessPost(t, "mobile/devices", payload, device)

	return device, devicePubKey
}

func CreateBrowserExtension(t *testing.T, name string) *BrowserExtensionResponse {
	keyPair := crypto.GenerateKeyPair(2048)

	pubKey := crypto.PublicKeyToBase64(keyPair.PublicKey)

	payload := []byte(fmt.Sprintf(`{"name":"%s","browser_name":"go-browser","browser_version":"0.1","public_key":"%s"}`, name, pubKey))

	browserExt := new(BrowserExtensionResponse)

	DoAPISuccessPost(t, "/browser_extensions", payload, browserExt)

	return browserExt
}

func CreateBrowserExtensionWithPublicKey(t *testing.T, name, publicKey string) *BrowserExtensionResponse {
	payload := []byte(fmt.Sprintf(`{"name":"%s","browser_name":"go-browser","browser_version":"0.1","public_key":"%s"}`, name, publicKey))

	browserExt := new(BrowserExtensionResponse)

	DoAPISuccessPost(t, "/browser_extensions", payload, browserExt)

	return browserExt
}

func PairDeviceWithBrowserExtension(t *testing.T, devicePubKey string, browserExtension *BrowserExtensionResponse, device *DeviceResponse) *PairingResultResponse {
	payload := struct {
		ExtensionId     string `json:"extension_id"`
		DeviceName      string `json:"device_name"`
		DevicePublicKey string `json:"device_public_key"`
	}{
		ExtensionId:     browserExtension.Id,
		DeviceName:      device.Name,
		DevicePublicKey: devicePubKey,
	}

	pairingResult := new(PairingResultResponse)

	payloadJson, _ := json.Marshal(payload)

	DoAPISuccessPost(t, "/mobile/devices/"+device.Id+"/browser_extensions", payloadJson, pairingResult)

	return pairingResult
}

func GetExtensionDevices(t *testing.T, extensionId string) []*ExtensionPairedDeviceResponse {
	var extensionDevices []*ExtensionPairedDeviceResponse

	DoAPISuccessGet(t, "/browser_extensions/"+extensionId+"/devices", &extensionDevices)

	return extensionDevices
}

func Request2FaToken(t *testing.T, domain, extensionId string) *AuthTokenRequestResponse {
	var response *AuthTokenRequestResponse

	payload := []byte(fmt.Sprintf(`{"domain":"%s"}`, domain))

	DoAPISuccessPost(t, "browser_extensions/"+extensionId+"/commands/request_2fa_token", payload, &response)

	return response
}

func Send2FaTokenToExtension(t *testing.T, extensionId, deviceId, requestId, token string) {
	j := fmt.Sprintf(`{"token_request_id":"%s","extension_id":"%s","token":"%s"}`, requestId, extensionId, token)

	DoAPISuccessPost(t, "mobile/devices/"+deviceId+"/commands/send_2fa_token", []byte(j), nil)
}

func RemoveAllBrowserExtensionsDevices(t *testing.T) {
	DoAdminSuccessDelete(t, "browser_extensions/devices")
}

func RemoveAllBrowserExtensions(t *testing.T) {
	DoAdminSuccessDelete(t, "browser_extensions")
}

func RemoveAllMobileDevices(t *testing.T) {
	DoAdminSuccessDelete(t, "/mobile/devices")
}

func RemoveAllMobileIconsCollections(t *testing.T) {
	DoAdminSuccessDelete(t, "mobile/icons/collections")
}

func RemoveAllMobileWebServices(t *testing.T) {
	DoAdminSuccessDelete(t, "mobile/web_services")
}

func RemoveAllMobileIcons(t *testing.T) {
	DoAdminSuccessDelete(t, "mobile/icons")
}

func RemoveAllMobileIconsRequests(t *testing.T) {
	DoAdminSuccessDelete(t, "mobile/icons/requests")
}

func RemoveAllMobileNotifications(t *testing.T) {
	DoAdminSuccessDelete(t, "mobile/notifications")
}
