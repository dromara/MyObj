package provider

import "myobj/src/pkg/cloudsync/internal"

// credentialSyncHelper 凭据轮换回写辅助
type credentialSyncHelper struct {
	onUpdate func(string)
}

func (h *credentialSyncHelper) SetCredentialUpdateCallback(fn func(string)) {
	h.onUpdate = fn
}

func (h *credentialSyncHelper) notifyCredential(refreshToken, clientID, clientSecret, defaultClientID, defaultClientSecret string) {
	if h.onUpdate == nil {
		return
	}
	cred := internal.FormatOAuthCredential(refreshToken, clientID, clientSecret, defaultClientID, defaultClientSecret)
	if cred != "" {
		h.onUpdate(cred)
	}
}
