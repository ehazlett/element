package element

import (
	"time"

	"github.com/sirupsen/logrus"
)

// Start handles cluster events
func (a *Agent) Start() error {
	go func() {
		for range a.peerUpdateChan {
			if err := a.members.UpdateNode(nodeUpdateTimeout); err != nil {
				logrus.Errorf("error updating node metadata: %s", err)
			}
		}
	}()
	if len(a.config.Peers) > 0 {
		doneCh := make(chan bool)
		go func() {
			for {
				logrus.Debugf("attempting to join peers...")
				if _, err := a.members.Join(a.config.Peers); err != nil {
					logrus.WithError(err).Warn("unable to join")
					time.Sleep(1 * time.Second)
					continue
				}
				break
			}
			doneCh <- true
		}()

		select {
		case <-time.After(a.config.LeaderPromotionTimeout):
			logrus.Infof("timeout (%s) trying to join peers; self-electing as leader", a.config.LeaderPromotionTimeout)
		case <-doneCh:
			logrus.Infof("joined peers %s", a.config.Peers)
		}
	}
	return nil
}
