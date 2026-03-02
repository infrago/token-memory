# token-memory

`token-memory` 是 `token` 模块的内存 `Driver` 实现（driver: `github.com/infrago/token-memory`）。

## 包定位

- 类型：驱动（Driver）
- 作用：开发和单进程场景的默认存储实现

## 主要功能

- 生命周期：`Open/Close`（内存实现，无外部连接）
- `payload` 存储：`SavePayload/LoadPayload/DeletePayload`
- `revoke` 存储：`RevokeToken/RevokeTokenID/Revoked*`
- TTL 过期惰性清理

## 配置

`token-memory` 不依赖外部连接，默认无需额外配置。

## 使用

```go
import _ "github.com/infrago/token-memory"
```

```toml
[token]
driver = "memory"
payload = "token" # token | store | hybrid
```
