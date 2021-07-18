// SPDX-License-Identifier: AGPL-3.0-only
// 🄯 2021, Alexey Parfenov <zxed@alkatrazstudio.net>

package src

import (
	"feedmash/util"
	"os"
	"os/signal"
)

func run(cfg Config) {
	nSources := len(cfg.sources)
	sourceFeedsChan := make(chan *FeedChanItem, nSources)
	sourceFeedsReceiverStopped := make(chan bool)
	outXmlChan := make(chan string)
	go startSourceFeedsReceiver(cfg, sourceFeedsChan, sourceFeedsReceiverStopped, outXmlChan)

	feedSources := loadSources(cfg.sources)
	go startWatchingFeeds(feedSources, cfg, sourceFeedsChan)

	srvStop := make(chan bool)
	srvStopped := make(chan bool)
	go runServer(cfg.serverAddr, srvStop, srvStopped, outXmlChan)

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	sourceFeedsReceiverIsStopped := false
	select {
	case <-sigChan:
		util.LogInfo("") // to not print the next message on the same line as ^C
		util.LogWarn("Interrupt received.")

	case <-srvStopped:
		util.LogWarn("Server was stopped abnormally.")

	case <-sourceFeedsReceiverStopped:
		util.LogWarn("Source feeds receiver stopped abnormally.")
		sourceFeedsReceiverIsStopped = true
	}

	sourceFeedsChan <- nil
	srvStop <- true
	stopWatchingFeeds(feedSources)
	<-srvStopped
	if !sourceFeedsReceiverIsStopped {
		<-sourceFeedsReceiverStopped
	}
}

func Main(exampleYaml string) {
	handleCli(run, exampleYaml)
}
