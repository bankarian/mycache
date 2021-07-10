package core

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github/mycache/pb"
	"io/ioutil"
	"net/http"
	"net/url"
)

type httpFetcher struct {
	baseURL string
}

func (h *httpFetcher) Fetch(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.Group), url.QueryEscape(in.Key))
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returns: %v", res.Status)
	}

	byts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading http response body err: %v", err)
	}
	if err := proto.Unmarshal(byts, out); err != nil {
		return fmt.Errorf("decoding rpc response body err: %v", err)
	}
	return nil
}

var _ Peer = (*httpFetcher)(nil)
