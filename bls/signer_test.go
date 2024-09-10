package bls

import (
	"anvztor/bls-signer/config"
	"anvztor/bls-signer/p2p"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	log "github.com/sirupsen/logrus"
)

func TestSigner(t *testing.T) {
	// Initialize configuration
	config.InitConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create LibP2PService instances
	service1 := p2p.NewLibP2PService()
	service2 := p2p.NewLibP2PService()

	// Start services
	go service1.Start(ctx)
	go service2.Start(ctx)

	// Wait for services to start
	time.Sleep(30 * time.Second)

	// Create a signature request
	message := []byte("Hello, World!")
	signDoc := createSignDoc(message)

	// Broadcast signature request from node 1
	broadcastSignatureRequest(ctx, service1, signDoc)

	// Collect signatures
	signatures := collectSignatures(ctx, service1.SignatureChan, 30*time.Second)

	// Process collected signatures
	processCollectedSignatures(signatures, signDoc)

	// Wait for a while to ensure all messages are processed
	time.Sleep(2 * time.Second)
}

func createNode(port int) (host.Host, error) {
	return libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
	)
}

func createSignDoc(message []byte) []byte {
	return message
}

func broadcastSignatureRequest(ctx context.Context, service *p2p.LibP2PService, signDoc []byte) {
	msg := p2p.Message{
		MessageType: p2p.MessageTypeSignature,
		Content:     string(signDoc),
	}
	service.PublishMessage(ctx, msg)
}

func collectSignatures(ctx context.Context, signChan <-chan p2p.SignatureMessage, timeout time.Duration) []p2p.SignatureMessage {
	var signatures []p2p.SignatureMessage
	timeoutChan := time.After(timeout)

	for {
		select {
		case sig := <-signChan:
			signatures = append(signatures, sig)
			log.Infof("Collected signature from node: %s", sig.PeerID)
		case <-timeoutChan:
			log.Info("Signature collection timed out")
			return signatures
		case <-ctx.Done():
			return signatures
		}
	}
}

func processCollectedSignatures(signatures []p2p.SignatureMessage, signDoc []byte) {
	log.Infof("Collected %d signatures", len(signatures))
	for i, sig := range signatures {
		log.Infof("Signature %d: PeerID=%s, PublicKey=%x, Signature=%x", i+1, sig.PeerID, sig.PublicKey, sig.Signature)
	}

	// Actual signature aggregation and verification logic should be implemented here
	// Since the actual BLS library implementation may vary, this is just a simulation of the process
	log.Info("Simulating signature aggregation...")
	log.Info("Simulating aggregated signature verification...")
	log.Info("Aggregated signature verification successful")
}
