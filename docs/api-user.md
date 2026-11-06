# PcWebLogin

### Req

```json
{
  "header": {
    "key": "PcWebLogin",
    "resTime": 1686118416490,
    "reqSeq": 178232,
    "channel": "api-user",
    "sign": "2df72fb5b35f9b5535e06b786a5810c0",
    "version": ""
  },
  "body": {
    "param": "zb2sV3uNfNxSdGJlZIuto/07ECO5+XwXgwh3biexRDsKpMImBQOeFEza0BH0OlfCvXScOZwcue7HU68qzIsB4rV2dQF6gecJRFtUf6CF+jHdfgmW5IVTvvGiKeYiUCIXqwOzfdrgWCiCHqG81brYS3P3Z9ZnnBSAYQD6q9Yw4qw=",
    "clientId": "1001000021",
    "secret": true
  }
}
```

#### param

```json
{
  "phone": "186****5244",
  "password": "this is password",
  "uuid": "",
  "verifyCode": "",
  "clientSecret": "XFmi9GS2hzk98jGX"
}
```

### Resp

```json
{
  "STATUS": "200",
  "MSG": "服务调用成功！",
  "LOGID": "d7e87d6a95614e8eaef4a2f22bfaf7a6",
  "RSP": {
    "RSP_CODE": "0000",
    "RSP_DESC": "成功",
    "DATA": {
      "needSmsCode": "1"
    }
  }
}
```

# AppLoginByMobile

纯短信验证码登录（免密码），逆向自网页端 pan.wo.cn 登录页「短信登录」页签（2026-11-06）。
先用 [sendMessageCodeBase](#sendmessagecodebase) 发送短信，再凭收到的验证码调用本接口换取 token。

### Req

```json
{
  "header": {
    "key": "AppLoginByMobile",
    "resTime": 1762400000000,
    "reqSeq": 123456,
    "channel": "api-user",
    "sign": "<md5(key+resTime+reqSeq+channel+version)>",
    "version": ""
  },
  "body": {
    "param": "<AES加密>",
    "clientId": "1001000021",
    "secret": true
  }
}
```

#### param

```json
{
  "phone": "186****5244",
  "smsCode": "123456",
  "clientSecret": "XFmi9GS2hzk98jGX"
}
```

注意字段名是 `smsCode`（区别于 PcLoginVerifyCode 的 `messageCode`）。

### Resp

```json
{
  "STATUS": "200",
  "MSG": "服务调用成功！",
  "LOGID": "...",
  "RSP": {
    "RSP_CODE": "0000",
    "RSP_DESC": "成功",
    "DATA": {
      "access_token": "<uuid>",
      "refresh_token": "<uuid，是否返回待真机确认>",
      "expires_in": 604799
    }
  }
}
```

网页端只取 `access_token`；`refresh_token`/`expires_in` 实际返回情况需真机验证。

# sendMessageCodeBase

发送短信登录验证码。**独立端点，非 dispatcher 协议**（无 header/sign）。

### Req

```
POST https://panservice.mail.wo.cn/api-user/sendMessageCodeBase
Content-Type: application/json
Origin: https://pan.wo.cn
Referer: https://pan.wo.cn/
```

```json
{
  "func": "pc_send",
  "clientId": "1001000021",
  "param": "<AES加密>"
}
```

#### param 明文（api-user 通道 AES 加密前）

```json
{
  "operateType": "1",
  "phone": "186****5244",
  "uuid": "",
  "verifyCode": ""
}
```

`uuid`/`verifyCode` 为图形验证码对，仅风控触发时必填；正常传空串。
图形验证码图片：`GET https://panservice.mail.wo.cn/api-user/getverifycode?uuid=<uuid>`。

### Resp

```json
{ "STATUS": "200", "MSG": "服务调用成功！", "RSP": { "RSP_CODE": "0000", "RSP_DESC": "成功" } }
```

| RSP_CODE | 含义 | SDK 行为 |
|---|---|---|
| `0000` | 发送成功 | `NeedCaptcha=false` |
| `6006` | 需要图形验证码（`DATA.isNeedVerfyCode==1`） | `NeedCaptcha=true` |
| `6007` | 图形验证码错误 | `NeedCaptcha=true` |

# PcLoginVerifyCode

### Req

```json
{
  "header": {
    "key": "PcLoginVerifyCode",
    "resTime": 1686118433825,
    "reqSeq": 153713,
    "channel": "api-user",
    "sign": "c17d5b143168fce8dd5e86a4ed5d5d74",
    "version": ""
  },
  "body": {
    "param": "zb2sV3uNfNxSdGJlZIuto4JcpoFt3BgCZ6hy0mZI+IyuItwv33Ftg6SGaDc4c65fHQVowbewSxDw4h06CrLu1FiwSy+2+Q6T1soAI5RnMA4p3gXcwNcJZRwoYbZSyVZ31zs4PrXdxRbOC5/xPlQWZZ03lkT+JutfI6OH+7SvYkIK0Yg6y8dgZrxOAppLrjwp",
    "clientId": "1001000021",
    "secret": true
  }
}
```

#### param

```json
{
  "phone": "186****5244",
  "messageCode": "634978",
  "verifyCode": null,
  "uuid": null,
  "clientSecret": "XFmi9GS2hzk98jGX",
  "password": "this is password"
}
```

### Resp

```json
{
  "STATUS": "200",
  "MSG": "服务调用成功！",
  "LOGID": "1e664e05d9124feabcf08e3dba3d10cb",
  "RSP": {
    "RSP_CODE": "0000",
    "RSP_DESC": "成功",
    "DATA": {
      "access_token": "c3f0362b-xxxx-4997-a9c2-0478702a5fea",
      "refresh_token": "457192a6-xxxx-485e-8ab1-0922c17cca69",
      "expires_in": 604799
    }
  }
}
```

# AppQueryUser

### Req

```json
{
  "header": {
    "key": "AppQueryUser",
    "resTime": 1686118434179,
    "reqSeq": 106207,
    "channel": "api-user",
    "sign": "ce928821b32a3808925bd1343f26e78f",
    "version": ""
  },
  "body": {
    "param": "/OP/co9mYqLkkDks24f5wyqYIeq7+K+bJcKXsaRgM61hjy8vBTDNvjfqXhGvMcdTcVudN8Tf6m6nOOGpRPY5PA==",
    "clientId": "1001000021",
    "secret": true
  }
}
```

#### param

```json
{ "accessToken": "c3f0362b-xxxx-4997-a9c2-0478702a5fea" }
```

### Resp

```json
{
  "STATUS": "200",
  "MSG": "服务调用成功！",
  "LOGID": "2d38989e8d1145e8a9fd911b60775ff9",
  "RSP": {
    "RSP_CODE": "0000",
    "RSP_DESC": "成功",
    "DATA": "OUiQMdAaILKTyrcyS0RSd9Kdd86NnTndTnZ45C1iz1qqLlaMbPA7IEhEsyowFNSgj5lsTwXTWnA8e/gZsV1QW39oK1ZUn0Nw8ACNwHgUNsXSc/UjZ2vPf+mlXqCN2HL9JQLEddvvZER319qeaxrHI1A/t7XZOgX9YChhO9QR1tiakwLxKWaIDH4AnAmR7WSBmo2dWFoRzvJBibTCD3eSAjqGX6hOsZ2h5qyINDtS303qWhQRsYPIu3AVQ9RSub5OJpiLPUSIKS1vNQ9r/m2AgT5lLoSfNkOdoIHNhWbJqfC1elT1JC4IlybttsXQPqj60FK36ujvSLwsvXse+SKa2Q=="
  }
}
```

#### data

```json
{
  "userId": "186****5244",
  "headUrl": "https://panservice.mail.wo.cn/upload/headPortrait/person_1600762418973144.png",
  "userName": "wo_Kbc7Hb",
  "sex": "",
  "birthday": "",
  "isModify": "0",
  "isHeadModify": "0",
  "isSetPassword": "1",
  "registerTime": "20220221185416"
}
```
