package httpapi

import "net/http"

// pushRelay is public: mobile clients read it after sign-in to find the
// Pushover-compatible relay they register devices with. It advertises a relay
// only when Pushover delivery is configured, so clients never register with a
// relay this server does not send to.
func (s *Server) pushRelay(w http.ResponseWriter, r *http.Request) {
	url := ""
	if s.pushNotifier != nil {
		url = s.pushRelayURL
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}
