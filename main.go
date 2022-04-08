package snap

import (
	"bytes"
	"errors"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	. "github.com/Monibuca/engine/v3"
	. "github.com/Monibuca/utils/v3"

	"github.com/Monibuca/utils/v3/codec"
)

type SnapProcess = func(*Stream) ([]byte, error)

var snapKind = map[string]SnapProcess{
	"iframe": snapIFrame,
	"rtmp":   snapStream,
	"rtsp":   snapStream,
	"flv":    snapStream,
	"hls":    snapStream,
}

func init() {

	pc := PluginConfig{
		Name:   "Snap",
		Config: &struct{}{},
		// Version: "v3.0.0",
	}
	pc.Install(nil)
	http.HandleFunc("/api/snap", snap)
}

func snap(w http.ResponseWriter, r *http.Request) {
	CORS(w, r)
	timeout := r.URL.Query().Get("timeout")
	t, err := strconv.Atoi(timeout)
	if timeout == "" || t == 0 || err != nil {
		t = 2500
	}
	kind := r.URL.Query().Get("kind")
	if kind == "" {
		kind = "iframe"
	}
	snapProc, ok := snapKind[kind]
	if !ok {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("kind param is wrong"))
		return
	}

	if streamPath := r.URL.Query().Get("stream"); streamPath != "" {
		s := FindStream(streamPath)
		if s == nil {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte("stream not found"))
			return
		}
		data, err := snapWithTimeOut(s, snapProc, t)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Add("Content-Type", "image/jpeg")
		w.Write(data)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("stream param is empty"))
	}
}

func snapWithTimeOut(s *Stream, snap SnapProcess, timeout int) ([]byte, error) {

	after := time.After(time.Duration(timeout) * time.Millisecond)
	var data []byte
	var err error
L:
	for {
		select {
		case <-after:
			if err == nil {
				err = errors.New("timeout")
			}
			break L
		default:
			data, err = snap(s)
			if err == nil {
				break L
			}
		}
	}
	return data, err
}

func snapIFrame(s *Stream) ([]byte, error) {
	if v := s.WaitVideoTrack(); v != nil {
		buf := bytes.NewBuffer(nil)

		header := *v.ExtraData
		for _, h := range header.NALUs {
			buf.Write(codec.NALU_Delimiter2)
			buf.Write(h)
		}

		idr := v.IDRing.Value.(*AVItem).Value.(*VideoPack)
		for _, h := range idr.NALUs {
			buf.Write(codec.NALU_Delimiter2)
			buf.Write(h)
		}

		cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-vframes", "1", "-f", "mjpeg", "pipe:1")
		cmd.Stdin = bytes.NewReader(buf.Bytes())
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		if len(out.Bytes()) == 0 {
			return nil, errors.New("snap failed")
		}
		return out.Bytes(), nil
	} else {
		return nil, errors.New("stream no track")
	}
}

func snapStream(s *Stream) ([]byte, error) {

	return nil, errors.New("not implement yet")
}
