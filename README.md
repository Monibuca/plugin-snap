# plugin-snap
I帧截图插件

# 安装
1. 修改 Monibuca 项目中的 main.go ，在 import 里添加 `_ "github.com/Monibuca/plugin-snap/v3"`
2. 修改 config.toml 添加 `[Snap]`
3. 运行 `go mod tidy`

# 接口
`/api/snap?stream=视频流&timeout=超时毫秒`  

视频流即流列表中的StreamPath，timeout默认超时2500毫秒，可以不传
