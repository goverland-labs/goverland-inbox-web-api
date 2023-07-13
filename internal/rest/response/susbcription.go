package response

import (
	"fmt"
	"net/http"
)

const (
	HeaderSubscriptionsCount = "X-Subscriptions-Count"
)

func AddSubscriptionsCountHeaders(w http.ResponseWriter, cnt int) {
	w.Header().Set(HeaderSubscriptionsCount, fmt.Sprintf("%d", cnt))
}
