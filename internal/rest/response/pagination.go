package response

import (
	"fmt"
	"net/http"
	"net/url"
)

const (
	HeaderTotalCount   = "X-Total-Count"
	HeaderUnreadCount  = "X-Unread-Count"
	HeaderVpTotal      = "X-Total-Avg-Vp"
	HeaderOffset       = "X-Offset"
	HeaderLimit        = "X-Limit"
	HeaderPrevPageLink = "X-Prev-Page"
	HeaderNextPageLink = "X-Next-Page"
)

func AddPaginationHeaders(w http.ResponseWriter, r *http.Request, offset, limit, totalCnt int) {
	w.Header().Set(HeaderTotalCount, fmt.Sprintf("%d", totalCnt))
	w.Header().Set(HeaderOffset, fmt.Sprintf("%d", offset))
	w.Header().Set(HeaderLimit, fmt.Sprintf("%d", limit))

	if (offset + limit) < totalCnt {
		w.Header().Set(HeaderNextPageLink, replaceGetParameter(r.URL, "offset", fmt.Sprintf("%d", offset+limit)).String())
	}

	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}

		w.Header().Set(HeaderPrevPageLink, replaceGetParameter(r.URL, "offset", fmt.Sprintf("%d", prevOffset)).String())
	}
}

func AddUnreadHeader(w http.ResponseWriter, count int) {
	w.Header().Set(HeaderUnreadCount, fmt.Sprintf("%d", count))
}

func AddVpTotalHeader(w http.ResponseWriter, total float32) {
	w.Header().Set(HeaderVpTotal, fmt.Sprintf("%f", total))
}

func AddTotalCounterHeaders(w http.ResponseWriter, totalCnt int) {
	w.Header().Set(HeaderTotalCount, fmt.Sprintf("%d", totalCnt))
}

func replaceGetParameter(uri *url.URL, param string, value string) *url.URL {
	uriCopy, _ := url.Parse(uri.String())
	query := uriCopy.Query()
	query.Set(param, value)
	uriCopy.RawQuery = query.Encode()

	return uriCopy
}
