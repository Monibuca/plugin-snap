package snap

import (
	_ "embed"
	"net/http"
	"os/exec"
	"strings"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/util"
)

//go:embed default.yaml
var defaultYaml DefaultYaml

type SnapSubscriber struct {
	Subscriber
}
type SnapConfig struct {
	DefaultYaml
	config.Subscribe
	FFmpeg string // ffmpeg的路径
	Path   string //存储路径
	Filter string //过滤器
}

func (snap *SnapConfig) OnEvent(event any) {

}

var conf = &SnapConfig{
	DefaultYaml: defaultYaml,
}

var plugin = InstallPlugin(conf)

func (snap *SnapConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	streamPath := strings.TrimPrefix(r.RequestURI, "/snap/")
	w.Header().Set("Content-Type", "image/jpeg")
	sub := &SnapSubscriber{}
	sub.IsInternal = true
	sub.ID = r.RemoteAddr
	sub.SetParentCtx(r.Context())
	sub.SetIO(w)
	if err := plugin.SubscribeBlock(streamPath, sub, SUBTYPE_RAW); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *SnapSubscriber) OnEvent(event any) {
	switch v := event.(type) {
	case VideoFrame:
		s.Stop()
		var errOut util.Buffer
		firstFrame := v.GetAnnexB()
		cmd := exec.Command(conf.FFmpeg, "-hide_banner", "-i", "pipe:0", "-vframes", "1", "-f", "mjpeg", "pipe:1")
		cmd.Stdin = &firstFrame
		cmd.Stderr = &errOut
		cmd.Stdout = s
		cmd.Run()
		if errOut.CanRead() {
			s.Info(string(errOut))
		}
	default:
		s.Subscriber.OnEvent(event)
	}
}
