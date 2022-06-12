package snap

import (
	"bytes"
	"net/http"
	"os/exec"
	"strings"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/config"
)

type SnapSubscriber struct {
	Subscriber
}
type SnapConfig struct {
	config.Subscribe
	FFmpeg string // ffmpeg的路径
	Path   string //存储路径
	Filter string //过滤器
	cmd    string
}

func (snap *SnapConfig) OnEvent(event any) {

}

var conf = &SnapConfig{
	FFmpeg: "ffmpeg",
}
var plugin = InstallPlugin(conf)

func (snap *SnapConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	streamPath := strings.TrimPrefix(r.RequestURI, "/snap/")
	w.Header().Set("Content-Type", "image/jpeg")
	sub := &SnapSubscriber{}
	sub.ID = r.RemoteAddr
	sub.SetParentCtx(r.Context())
	sub.SetIO(w)
	if err := plugin.SubscribeBlock(streamPath, sub); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *SnapSubscriber) OnEvent(event any) {
	switch v := event.(type) {
	case *VideoFrame:
		if v.IFrame {
			var buff bytes.Buffer
			c := VideoDeConf(s.Video.Track.DecoderConfiguration).GetAnnexB()
			c.WriteTo(&buff)
			c = v.GetAnnexB()
			c.WriteTo(&buff)
			cmd := exec.Command(conf.FFmpeg, "-i", "pipe:0", "-vframes", "1", "-f", "mjpeg", "pipe:1")
			cmd.Stdin = &buff
			cmd.Stdout = s
			cmd.Run()
			s.Stop()
		}
	default:
		s.Subscriber.OnEvent(event)
	}
}
