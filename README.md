# plugin-snap
I 帧截图插件，通过命令行调用 ffmpeg 截取 I 帧视频

# 安装
1. 修改 Monibuca 项目中的 main.go ，在 import 里添加 `_ "github.com/Monibuca/plugin-snap/v3"`
2. 修改 config.toml 添加 `[Snap]`
3. 运行 `go mod tidy`
4. 需要 ffmpeg 可执行文件在 **PATH** 中，不依赖 cgo

# 接口
`/api/snap?stream=视频流&timeout=超时毫秒`  

视频流即流列表中的StreamPath，timeout默认超时2500毫秒，可以不传

# 说明
本插件有一定概率截图失败，一般情况下 I 帧 2~4秒 间隔出现一次，所以本插件不适合实时截图，本插件的适用范围为，定时截图或者播放视频前的预览图。 
