package lib

import (
	"fmt"
	"rs_server/lib/util"
	"time"

	log "github.com/cihub/seelog"
	zmq "github.com/zeromq/goczmq"
)

// const FrameWidth = 16
// const FrameHeight = 32
// const FrameDepth = 8
const FrameWidth = 15
const FrameHeight = 50
const FrameDepth = 15


type RealsenseService struct {
	rcvSock      *zmq.Sock
	sndSock      *zmq.Sock
	order        chan string
	enable 		chan bool
	done         chan struct{}
	led565Buffer []byte
}

func NewRealsenseService(rcvEp, sndEp string) *RealsenseService {
	rs := &RealsenseService{}
	rs.led565Buffer = make([]byte, FrameWidth*FrameHeight*FrameDepth*2)

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

func (rs *RealsenseService) Enable(enable bool){
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

	ranges := MakeDepthThresholds(FrameDepth, 128)
	timer := NewTimer(50 * time.Millisecond)
	mesureTicker := time.NewTicker(2 * time.Second)

	defer mesureTicker.Stop()

	var duration time.Duration
	warningMsg := ""
	var enable bool

	for {
		select {
		case <-rs.order:
			rs.done <- struct{}{}
			return
		case enable =<-rs.enable:
		case <-mesureTicker.C:
			if warningMsg != ""{
				log.Warn(warningMsg)
			}else{
				log.Info("Send Frame... last frame duration:", duration)
			}
		default:
			warningMsg = ""
			start := time.Now()
			rs.rcvSock.RecvFrame() // 読み捨て
			data, _, _ := rs.rcvSock.RecvFrame()
			if !enable{
				continue
			}

			if len(data) != FrameWidth * FrameHeight * 4{
				warningMsg = fmt.Sprintf("Invalid Data Length: %v" ,len(data))
				continue
			}

			if !timer.IsPast() {
				continue
			}
			util.ConcurrentEnum(0, len(rs.led565Buffer), func(i int) {
				rs.led565Buffer[i] = 0
			})

			util.ConcurrentEnum(0, FrameHeight, func(y int) {
				for x := 0; x < FrameWidth; x++ {
					idx := y*4 + FrameHeight*4*x

					r, g, b := byte(data[idx+2]), byte(data[idx+1]), byte(data[idx+0])
					depth := uint32(data[idx+3])
					if depth == 0 {
						continue
					}
					for z := FrameDepth - 1; z >= 0; z-- {
						if depth < ranges[z] {
							index565 := z*2 + y*FrameDepth*2 + x*FrameHeight*FrameDepth*2
							rs.led565Buffer[index565+0] = r&0xF8 + g>>5
							rs.led565Buffer[index565+1] = (g<<2)&0xe0 + b>>3
						} else {
							break
						}
					}
				}

			})
			rs.sndSock.SendFrame(rs.led565Buffer, 0)
			duration = time.Since(start)
		}
	}
}
