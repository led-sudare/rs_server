package lib

import (
	"fmt"
	"rs_server/lib/util"
	"time"

	log "github.com/cihub/seelog"
	zmq "github.com/zeromq/goczmq"
)

type frameFormat interface {
	GetWidth() int
	GetHeight() int
	GetDepth() int
	GetBuffer() []byte
	GetRanges() []uint32
}

type frameFormatImpl struct {
	Width, Height, Depth int
	Buffer               []byte
	Ranges               []uint32
}

func (f *frameFormatImpl) GetWidth() int {
	return f.Width
}
func (f *frameFormatImpl) GetHeight() int {
	return f.Height
}
func (f *frameFormatImpl) GetDepth() int {
	return f.Depth
}
func (f *frameFormatImpl) GetBuffer() []byte {
	return f.Buffer
}

func (f *frameFormatImpl) GetRanges() []uint32 {
	return f.Ranges
}

func newFrameFormatImpl(width, height, depth int) frameFormat {
	return &frameFormatImpl{
		Width:  width,
		Height: height,
		Depth:  depth,
		Buffer: make([]byte, width*height*depth*2),
		Ranges: MakeDepthThresholds(uint32(depth), 128)}
}

const rsColorPlane = 4

var frameFormat15x50x15 frameFormat
var frameFormat30x100x30 frameFormat

func newFrameFormat(dataLen int) frameFormat {

	switch dataLen {
	case 15 * 50 * rsColorPlane: // w * h * plane
		if frameFormat15x50x15 == nil {
			frameFormat15x50x15 = newFrameFormatImpl(15, 50, 15)
		}
		return frameFormat15x50x15
	case 30 * 100 * rsColorPlane: // w * h * plane
		if frameFormat30x100x30 == nil {
			frameFormat30x100x30 = newFrameFormatImpl(30, 100, 30)
		}
		return frameFormat30x100x30
	default:
		return nil
	}
}

type RealsenseService struct {
	rcvSock *zmq.Sock
	sndSock *zmq.Sock
	order   chan string
	enable  chan bool
	done    chan struct{}
	//	led565Buffer []byte
}

func NewRealsenseService(rcvEp, sndEp string) *RealsenseService {
	rs := &RealsenseService{}
	//rs.led565Buffer = make([]byte, FrameWidth*FrameHeight*FrameDepth*2)

	var err error
	rs.rcvSock = zmq.NewSock(zmq.Sub)
	rs.rcvSock.SetSubscribe("")
	err = rs.rcvSock.Connect(rcvEp)
	if err != nil {
		panic(err)
	}
	log.Info("Sub Socket for RS Start..: ", rcvEp)

	rs.sndSock = zmq.NewSock(zmq.Pub)
	err = rs.sndSock.Connect(sndEp)
	if err != nil {
		panic(err)
	}
	log.Info("Pub Socket for Adapter Start..: ", sndEp)

	rs.order = make(chan string)
	rs.enable = make(chan bool)
	rs.done = make(chan struct{})

	return rs
}

func (rs *RealsenseService) Start() {
	go RealsenseServiceWorker(rs)
}

func (rs *RealsenseService) Enable(enable bool) {
	rs.enable <- enable
}

func (rs *RealsenseService) Stop() {
	rs.order <- ""
	<-rs.done
}
func (rs *RealsenseService) Destory() {
	defer rs.rcvSock.Destroy()
	defer rs.sndSock.Destroy()

}

func MakeDepthThresholds(level, max uint32) []uint32 {

	step := max / level
	thres := make([]uint32, level)
	for i := uint32(0); i < level; i++ {
		thres[i] = i * step
	}
	return thres
}

func RealsenseServiceWorker(rs *RealsenseService) {

	//ranges := MakeDepthThresholds(FrameDepth, 128)
	timer := NewTimer(50 * time.Millisecond)
	mesureTicker := time.NewTicker(2 * time.Second)

	defer mesureTicker.Stop()

	var duration time.Duration = -1
	warningMsg := ""
	var enable bool = true

	for {
		select {
		case <-rs.order:
			rs.done <- struct{}{}
			return
		case enable = <-rs.enable:
		case <-mesureTicker.C:
			if warningMsg != "" {
				log.Warn(warningMsg)
			} else {
				log.Info("Send Frame... last frame duration:", duration)
			}
		default:
			warningMsg = "data waiting..."
			start := time.Now()
			rs.rcvSock.RecvFrame() // 読み捨て
			data, _, _ := rs.rcvSock.RecvFrame()
			warningMsg = ""
			if !enable {
				continue
			}

			frame := newFrameFormat(len(data))
			//			if len(data) != FrameWidth*FrameHeight*4 {
			if frame == nil {
				warningMsg = fmt.Sprintf("Invalid Data Length: %v", len(data))
				continue
			}

			if !timer.IsPast() {
				continue
			}
			util.ConcurrentEnum(0, len(frame.GetBuffer()), func(i int) {
				frame.GetBuffer()[i] = 0
			})

			util.ConcurrentEnum(0, frame.GetHeight(), func(y int) {
				for x := 0; x < frame.GetWidth(); x++ {
					idx := y*4 + frame.GetHeight()*4*x

					r, g, b := byte(data[idx+2]), byte(data[idx+1]), byte(data[idx+0])
					depth := uint32(data[idx+3])
					if depth == 0 {
						continue
					}
					for z := frame.GetDepth() - 1; z >= 0; z-- {
						if depth < frame.GetRanges()[z] {
							index565 := z*2 + y*frame.GetDepth()*2 + x*frame.GetHeight()*frame.GetDepth()*2
							frame.GetBuffer()[index565+0] = r&0xF8 + g>>5
							frame.GetBuffer()[index565+1] = (g<<2)&0xe0 + b>>3
						} else {
							break
						}
					}
				}

			})
			rs.sndSock.SendFrame(frame.GetBuffer(), 0)
			duration = time.Since(start)
		}
	}
}
