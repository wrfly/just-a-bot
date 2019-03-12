package client

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/go-redis/redis"

	"github.com/sirupsen/logrus"
)

type roundTripper struct {
	tp    *http.Transport
	redis *redis.Client
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.URL.String()
	if bs, err := rt.redis.Get(key).Bytes(); err == nil {
		logrus.Debugf("got cache for %s", key)
		return &http.Response{
			Body: ioutil.NopCloser(bytes.NewReader(bs)),
		}, nil
	}

	resp, err := rt.tp.RoundTrip(req)
	if err == nil {
		// set cache
		bs, _ := ioutil.ReadAll(resp.Body)
		logrus.Debugf("cache [%s] resp: %d", req.URL.String(), len(bs))
		rt.redis.Set(key, bs, -1)
		// rebuild cache
		resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
	}
	return resp, err
}
