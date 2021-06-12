package main

import (
	"flag"
	"github.com/grandcat/zeroconf"
	"goplay2/event"
	"goplay2/handlers"
	"goplay2/homekit"
	"goplay2/ptp"
	"goplay2/rtsp"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

const deviceName = "aiwa"

func setLog() {
	file, err := os.OpenFile("goplay.log", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	log.SetOutput(file)
}

func main() {
	var ifName string
	var delay int64

	//setLog()

	flag.StringVar(&ifName, "i", "en0", "Specify interface")
	flag.Int64Var(&delay, "delay", -50, "Specify hardware delay in ms (useful on slow computer)")
	flag.Parse() // after declaring flags we need to call it

	iFace, err := net.InterfaceByName(ifName)
	if err != nil {
		panic(err)
	}
	macAddress := strings.ToUpper(iFace.HardwareAddr.String())
	homekit.Aiwa = homekit.NewAccessory(macAddress, aiwaDevice())
	log.Printf("Aiwa %v", homekit.Aiwa)
	homekit.Server, err = homekit.NewServer(macAddress, deviceName)

	server, err := zeroconf.Register(deviceName, "_airplay._tcp", "local.",
		7000, homekit.Aiwa.ToRecords(), []net.Interface{*iFace})
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	clock := ptp.NewVirtualClock(delay)
	ptp := ptp.NewServer(clock)

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go func() {
		event.RunEventServer()
		wg.Done()
	}()

	go func() {
		ptp.Serve()
	}()

	go func() {
		rtsp.RunRtspServer(handlers.NewRstpHandler(clock))
		wg.Done()
	}()

	wg.Wait()
}
