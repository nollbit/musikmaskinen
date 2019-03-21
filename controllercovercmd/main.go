package main

import (
	"os"
	"os/signal"
	"sync"

	log "github.com/sirupsen/logrus"

	ui "github.com/gizak/termui/v3"

	mmwidgets "github.com/nollbit/musikmaskinen/widgets"
)

func main() {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	termWidth, termHeight := ui.TerminalDimensions()

	block := mmwidgets.NewFadedBlock()
	block.Border = false
	block.SetRect(0, 0, termWidth, termHeight)
	ui.Render(block)

	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
}
