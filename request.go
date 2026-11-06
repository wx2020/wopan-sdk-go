package wopan

import (
	"fmt"
	"strings"
)

func (w *WoClient) request(channel string, key string, param, other Json, resp interface{}, retry bool, opts ...RestyOption) ([]byte, error) {
	req := w.NewRequest()
	req.SetHeader("User-Agent", w.ua)
	req.SetHeaders(map[string]string{
		"Origin":  "https://pan.wo.cn",
		"Referer": "https://pan.wo.cn/",
		// 浏览器指纹头：缺失会被 WAF 判定为非浏览器流量而 403 拦截（2026-11-06 真机实测，
		// 仅 Origin/Referer/UA 不够，登录类 dispatcher 必须带完整 Client Hints / Sec-Fetch 家族）
		"Accept":             "application/json, text/plain, */*",
		"sec-ch-ua":          `"Not.A/Brand";v="8", "Chromium";v="114", "Microsoft Edge";v="114"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
	})
	if w.accessToken != "" {
		req.SetHeader("Accesstoken", w.accessToken)

	}
	header := calHeader(channel, key)
	body, err := w.NewBody(channel, param, other)
	if err != nil {
		return nil, err
	}
	var _resp Resp
	req.SetBody(Req[interface{}]{
		Header: header,
		Body:   body,
	}).SetResult(&_resp)
	for _, opt := range opts {
		opt(req)
	}
	res, err := req.Post(fmt.Sprintf("%s/%s/dispatcher", DefaultBaseURL, channel))
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return res.Body(), fmt.Errorf("request failed with status: %s", res.Status())
	}
	if _resp.Status != "200" {
		return res.Body(), fmt.Errorf("request failed with status: %s, msg: %s", _resp.Status, _resp.Msg)
	}
	if _resp.Rsp.RspCode != "0000" {
		if channel != ChannelAPIUser && retry && _resp.Rsp.RspCode == "9999" {
			err := w.RefreshToken()
			if err != nil {
				return res.Body(), err
			}
			return w.request(channel, key, param, other, resp, false, opts...)
		}
		return res.Body(), fmt.Errorf("request failed with rsp_code: %s,rep_desc: %s", _resp.Rsp.RspCode, _resp.Rsp.RspDesc)
	}
	if resp != nil {
		data := string(_resp.Rsp.Data)
		if strings.HasSuffix(data, "\"") && strings.HasPrefix(data, "\"") {
			data, err = w.crypto.Decrypt(data[1:len(data)-1], channel)
			if err != nil {
				return res.Body(), err
			}
		}
		err = w.jsonUnmarshalFunc([]byte(data), resp)
		if err != nil {
			return res.Body(), err
		}
	}
	return res.Body(), nil
}

func (w *WoClient) Request(channel string, key string, param, other Json, resp interface{}, opts ...RestyOption) ([]byte, error) {
	return w.request(channel, key, param, other, resp, true, opts...)
}

func (w *WoClient) RequestApiUser(key string, param, other Json, resp interface{}, opts ...RestyOption) ([]byte, error) {
	return w.Request(ChannelAPIUser, key, param, other, resp, opts...)
}

func (w *WoClient) RequestWoHome(key string, param, other Json, resp interface{}, opts ...RestyOption) ([]byte, error) {
	return w.Request(ChannelWoHome, key, param, other, resp, opts...)
}
