package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type roundTripper struct {
	tp    http.RoundTripper
	redis *redis.Client
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + req.URL.String()
	if bs, err := rt.redis.Get(key).Bytes(); err == nil {
		logrus.Debugf("got cache for %s", key)
		respBS, _ := rt.redis.Get(key + ":header").Bytes()
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header, 0),
			Request:    req,
		}
		if err := json.Unmarshal(respBS, &resp.Header); err != nil {
			logrus.Fatalf("unmarshal err: %s", err)
		}
		resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		return resp, nil
	}

	resp, err := rt.tp.RoundTrip(req)
	if err == nil {
		// set cache
		bs, _ := ioutil.ReadAll(resp.Body)
		logrus.Debugf("cache [%s] resp: %d", req.URL.String(), len(bs))
		rt.redis.Set(key, bs, -1)
		headerBS, err := json.Marshal(resp.Header)
		if err != nil {
			logrus.Fatal(err)
		}
		rt.redis.Set(key+":header", headerBS, time.Hour*24)

		// rebuild resp body
		resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
	}

	return resp, err
}
