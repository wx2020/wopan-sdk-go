# 短信验证码直接获取 Token — 逆向结论与实现方案

> 2026-11-06 逆向 pan.wo.cn 网页端（app.5d6a7b53446ccd473649.js + chunk-2795042e 登录页）得出。
> **结论：官方网页端存在纯短信验证码登录（免密码），dispatcher key 为 `AppLoginByMobile`，SDK 可直接补齐。**

## 1. 逆向证据链

### 1.1 登录页有两个页签（chunk-2795042e / login.js）

- 页签 1「密码登录」（activeName="1"）→ `dispatch("user/login")` → `PcWebLogin` → 返回 `needSmsCode` → 弹二次验证框 → `PcLoginVerifyCode`（SDK 已有，需密码）
- **页签 2「短信登录」（activeName="2"）** → `dispatch("user/smsLogin", ruleForm2)` → **`AppLoginByMobile`** → 成功后直接 `SET_TOKEN(RSP.DATA.access_token)`

### 1.2 短信登录 payload（免密码）

```js
// login.js 数据定义
ruleForm2: { phone: "", smsCode: "", clientSecret: "" }
// mounted 时填充
this.ruleForm2.clientSecret = queryConfigByClientId(this.clientId).secretKey
// 提交：直接把 ruleForm2 作为 dispatcher param（secret=true → AES 加密）
dispatch("user/smsLogin", this.ruleForm2)
```

即 **`AppLoginByMobile` 的 param = `{"phone": "...", "smsCode": "123456", "clientSecret": "XFmi9GS2hzk98jGX"}`**（AES 加密传输）。
注意字段名是 **`smsCode`**（区别于 `PcLoginVerifyCode` 的 `messageCode`）。

### 1.3 发送验证码端点（非 dispatcher 协议）

```
POST https://panservice.mail.wo.cn/api-user/sendMessageCodeBase
Content-Type: application/json
Origin: https://pan.wo.cn
Referer: https://pan.wo.cn/

{
  "func": "pc_send",
  "clientId": "1001000021",
  "param": "<AES-CBC加密后的base64>"
}
```

param 明文（加密前）：

```json
{ "operateType": "1", "phone": "手机号", "uuid": "", "verifyCode": "" }
```

- `uuid`/`verifyCode` 为图形验证码对，**仅在风控触发时必填**；正常情况传空串即可
- 响应 `RSP_CODE`：
  - `0000` 发送成功
  - `6006` + `DATA.isNeedVerfyCode==1` → 需要图形验证码（先取图，带 uuid+verifyCode 重发）
  - `6007` → 图形验证码错误
- 图形验证码图片地址：`https://panservice.mail.wo.cn/api-user/getverifycode?uuid=<uuid>`（uuid 由前端生成）

### 1.4 加密参数（与 SDK 现有实现完全一致，无需改动 crypto）

网页端密钥来自 `bridge.js`（混淆后还原）：

| 变量 | 值 | SDK 对应 |
|---|---|---|
| `window.obfuscatorAESCodeProd` | `XFmi9GS2hzk98jGX` | `DefaultClientSecret`（api-user 通道 AES key） |
| `window.obfuscatorAESIv` | `wNSOYIB1k1DjY5lA` | `Crypto.iv` |

⇒ SDK 的 `EncryptParam(ChannelAPIUser, param)` 产出与网页端逐字节同构，**直接复用**。

### 1.5 需真机确认的未知数

1. `AppLoginByMobile` 响应 DATA 是否含 `refresh_token`（网页端只取了 `access_token`，无法从 JS 确认）——若无，则该 token 7 天过期后无法静默续期，只能重新发码登录
2. `expires_in` 实际值（对照：PcLoginVerifyCode 为 604799 ≈ 7 天）
3. 发码频控细节（前端 59s 倒计时；每日/每号上限未知）

## 2. SDK 改动设计

### 2.1 consts.go — 新增 key

```go
// api-user methods
const (
	// ... 现有 ...
	KeyAppLoginByMobile = "AppLoginByMobile"
)
```

### 2.2 api-user.go — 新增类型与方法

```go
// AppLoginByMobileData no encrypt
type AppLoginByMobileData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"` // 是否返回需真机确认
	ExpiresIn    int    `json:"expires_in"`
}

// AppLoginByMobile 纯短信验证码登录（免密码）。
// 先用 SendMessageCode 触发短信，再凭收到的验证码换取 token。
func (w *WoClient) AppLoginByMobile(phone, smsCode string, opts ...RestyOption) (*AppLoginByMobileData, error) {
	var resp AppLoginByMobileData
	_, err := w.RequestApiUser(KeyAppLoginByMobile, Json{
		"phone":        phone,
		"smsCode":      smsCode,
		"clientSecret": DefaultClientSecret,
	}, JsonClientIDSecret, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// LoginBySmsCode 登录并自动把 token 装入客户端（便捷方法）
func (w *WoClient) LoginBySmsCode(phone, smsCode string, opts ...RestyOption) (*AppLoginByMobileData, error) {
	resp, err := w.AppLoginByMobile(phone, smsCode, opts...)
	if err != nil {
		return nil, err
	}
	if resp.AccessToken != "" {
		w.SetAccessToken(resp.AccessToken)
	}
	if resp.RefreshToken != "" {
		w.SetRefreshToken(resp.RefreshToken)
	}
	return resp, nil
}

// SendMessageCodeResult no encrypt
type SendMessageCodeResult struct {
	NeedCaptcha bool   // true = 服务端要求图形验证码（6006），需带 uuid+verifyCode 重试
	RspCode     string // 原始 RSP_CODE，便于调用方细分处理
}

// SendMessageCode 发送短信登录验证码（operateType=1）
func (w *WoClient) SendMessageCode(phone string, opts ...RestyOption) (*SendMessageCodeResult, error) {
	return w.sendMessageCode(phone, "", "", opts...)
}

// SendMessageCodeWithCaptcha 风控触发后，带图形验证码重发
func (w *WoClient) SendMessageCodeWithCaptcha(phone, uuid, verifyCode string, opts ...RestyOption) (*SendMessageCodeResult, error) {
	return w.sendMessageCode(phone, uuid, verifyCode, opts...)
}

func (w *WoClient) sendMessageCode(phone, uuid, verifyCode string, opts ...RestyOption) (*SendMessageCodeResult, error) {
	param, err := w.EncryptParam(ChannelAPIUser, Json{
		"operateType": "1",
		"phone":       phone,
		"uuid":        uuid,
		"verifyCode":  verifyCode,
	})
	if err != nil {
		return nil, err
	}
	req := w.NewRequest()
	req.SetHeaders(map[string]string{
		"Origin":      "https://pan.wo.cn",
		"Referer":     "https://pan.wo.cn/",
		"Content-Type": "application/json",
	})
	req.SetBody(Json{
		"func":     "pc_send",
		"clientId": DefaultClientID,
		"param":    param,
	}).SetResult(&Resp{}) // 结构同 dispatcher 的 Resp
	for _, opt := range opts {
		opt(req)
	}
	res, err := req.Post(DefaultBaseURL + "/api-user/sendMessageCodeBase")
	if err != nil {
		return nil, err
	}
	var resp Resp
	if err := w.jsonUnmarshalFunc(res.Body(), &resp); err != nil {
		return nil, err
	}
	switch resp.Rsp.RspCode {
	case "0000":
		return &SendMessageCodeResult{RspCode: resp.Rsp.RspCode}, nil
	case "6006", "6007": // 需要图形验证码 / 图形验证码错误
		return &SendMessageCodeResult{NeedCaptcha: true, RspCode: resp.Rsp.RspCode}, nil
	default:
		return nil, fmt.Errorf("sendMessageCode failed with rsp_code: %s, rsp_desc: %s", resp.Rsp.RspCode, resp.Rsp.RspDesc)
	}
}

// GetVerifyCodeURL 图形验证码图片地址（uuid 自行生成并透传给 SendMessageCodeWithCaptcha）
func GetVerifyCodeURL(uuid string) string {
	return DefaultBaseURL + "/api-user/getverifycode?uuid=" + uuid
}
```

注意：`sendMessageCodeBase` 不走 `w.request()`（该函数写死了 `/{channel}/dispatcher` URL），需要独立小方法；但响应结构与 `Resp` 完全同形，直接复用类型。

### 2.3 docs/api-user.md — 补两节抓包文档

`AppLoginByMobile`（Req/Resp 样例，真机抓包后补 DATA 全字段）与 `sendMessageCodeBase`（端点、明文/密文 param、6006/6007 风控码）。

### 2.4 可选（低优先级）

- `header.go` 的 `reqSeq` 范围比网页端窄（SDK `1e5~1.1e5`，网页 `1e5~1.9e5`），服务端未见校验，可不改
- uuid 生成：网页端是 nanoid 风格随机串，任何唯一随机串即可（服务端只做一次性绑定）

## 3. 真机验证步骤（需收短信，须用户配合）

1. `SendMessageCode("<手机号>")` → 手机收到 6 位短信码 → 若返回 `NeedCaptcha=true`，先 GET `GetVerifyCodeURL(uuid)` 人工识图，再 `SendMessageCodeWithCaptcha`
2. `LoginBySmsCode("<手机号>", "<短信码>")` → 观察返回的 `access_token` / `refresh_token` / `expires_in`
3. 用返回 token 调 `AppQueryUser()` 验证 token 有效
4. 若无 `refresh_token`：记录 7 天后过期行为，SDK 文档注明"此登录方式 token 有效期 7 天，过期需重新短信登录"；若有：走 `AppRefreshToken` 验证续期

## 4. 风险与边界

| 项 | 评估 |
|---|---|
| 加密/签名兼容性 | 无风险（密钥、iv、sign 算法与网页端一致，已逐项核对） |
| `smsCode` 字段名 | 来自网页端实际请求，可信 |
| 响应结构 | access_token 字段确定；refresh_token 待确认 |
| 风控 | 高频发码会触发图形验证码（6006），SDK 已设计为结构化返回而非硬错误 |
| 合规 | 与 PcWebLogin/PcLoginVerifyCode 同源同性质（官方网页端在用能力），无新增攻击面 |
