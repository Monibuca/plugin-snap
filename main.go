package snap

import (
	"bytes"
	_ "embed"
	"net/http"
	"os/exec"
	"strings"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/config"
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
	sub.ID = r.RemoteAddr
	sub.SetParentCtx(r.Context())
	sub.SetIO(w)
	if err := plugin.SubscribeBlock(streamPath, sub, SUBTYPE_RAW); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *SnapSubscriber) OnEvent(event any) {
	switch v := event.(type) {
	case VideoDeConf:
		var buff bytes.Buffer
		var errOut bytes.Buffer
		for _, nalu := range v.Raw {
			buff.Write(codec.NALU_Delimiter2)
			buff.Write(nalu)
		}
		buff.Write(codec.NALU_Delimiter2)
		for i, nalu := range s.Video.Track.IDRing.Value.Raw {
			if i > 0 {
				buff.Write(codec.NALU_Delimiter1)
			}
			for _, slice := range nalu {
				buff.Write(slice)
			}
		}
		cmd := exec.Command(conf.FFmpeg, "-i", "pipe:0", "-vframes", "1", "-f", "mjpeg", "pipe:1")
		cmd.Stdin = &buff
		cmd.Stderr = &errOut
		cmd.Stdout = s
		cmd.Run()
		if len(errOut.Bytes()) > 0 {
			s.Info(errOut.String())
		}
		s.Stop()
	default:
		s.Subscriber.OnEvent(event)
	}
}
