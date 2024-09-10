package main

import (
	"anvztor/bls-signer/config"
	"anvztor/bls-signer/crypto"
	"anvztor/bls-signer/p2p"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	// Create custom Formatter
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
		PadLevelText:  true,
		ForceQuote:    true,
		DisableQuote:  false,
		FieldMap: log.FieldMap{
			log.FieldKeyTime:  "time",
			log.FieldKeyLevel: "level",
			log.FieldKeyMsg:   "message",
		},
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		},
	})

	// Set log level
	log.SetLevel(log.InfoLevel)

	// Add process PID field
	log.AddHook(&pidHook{})
}

// pidHook is a custom logrus Hook
type pidHook struct{}

// Levels defines which log levels this Hook applies to
func (h *pidHook) Levels() []log.Level {
	return log.AllLevels
}

// Fire is called every time a log entry is made
func (h *pidHook) Fire(entry *log.Entry) error {
	entry.Data["PID"] = strconv.Itoa(os.Getpid())
	return nil
}

func processCollectedSignatures(signatures []p2p.SignatureMessage, signDoc []byte) {
	var publicKeys []crypto.PublicKeyCompress
	var signatureCompresses []crypto.SignatureCompress

	for _, sig := range signatures {
		publicKeys = append(publicKeys, sig.PublicKey)
		signatureCompresses = append(signatureCompresses, sig.Signature)
	}

	// Aggregate signatures
	aggSignature := crypto.AggregateSignatures(signatureCompresses)

	// Verify aggregated signature
	isValid := crypto.VerifyAggregatedSignature(aggSignature, publicKeys, signDoc)

	if isValid {
		fmt.Println("Aggregated signature is valid.")
	} else {
		fmt.Println("Aggregated signature is invalid.")
	}
}

func main() {
	config.InitConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	libp2pService := p2p.NewLibP2PService()
	go libp2pService.Start(ctx)

	// Wait for P2P network initialization
	time.Sleep(15 * time.Second)
	// Generate message to be signed
	// Broadcast SignatureMessage
	// Wait for signature collection
	timeout := 30 * time.Second
	// Process collected signatures
	// TODO: Use better data structure to handle multi sigDoc signatures
	var collectedSignatures []p2p.SignatureMessage

	timeoutChan := time.After(timeout)
	if config.AppConfig.Proposer == true {
		log.Infof("Proposer initiates signature message...")
		message := []byte("Hello, BLS signature in Golang!")
		// Create SignDoc
		signDoc := crypto.CreateSignDoc(message)
		privateKey := crypto.GeneratePrivateKey()
		publicKey := crypto.GeneratePublicKey(privateKey)
		signature, err := crypto.SignMessage(privateKey, signDoc)
		signMsg := p2p.SignatureMessage{
			PeerID:    libp2pService.GetPeerID(),
			SignDoc:   signDoc,
			Signature: signature,
			PublicKey: publicKey,
		}
		collectedSignatures = append(collectedSignatures, signMsg)
		// TODO: Use crypto serialize and deserialize
		signMsgBytes, err := json.Marshal(signMsg)
		if err != nil {
			log.Errorf("Failed to marshal message: %v", err)
			return
		}
		msg := p2p.Message{
			MessageType: p2p.MessageTypeSignature,
			Content:     string(signMsgBytes),
		}
		libp2pService.PublishMessage(ctx, msg)
	}

	log.Infof("Start waiting for signature messages...")
	for {
		select {
		case receivedMsg := <-libp2pService.MessageChan:
			switch receivedMsg.MessageType {
			case p2p.MessageTypeSignature:
				log.Info("Received message to be signed")
				signMsg := p2p.SignatureMessage{}
				// TODO: Use crypto serialize and deserialize
				json.Unmarshal([]byte(receivedMsg.Content), &signMsg)
				privateKey := crypto.GeneratePrivateKey()
				publicKey := crypto.GeneratePublicKey(privateKey)
				signature, err := crypto.SignMessage(privateKey, signMsg.SignDoc)
				if err != nil {
					log.Errorf("Failed to generate signature: %v", err)
					continue
				}

				signedMsg := p2p.SignatureMessage{
					PeerID:    libp2pService.GetPeerID(),
					SignDoc:   signMsg.SignDoc,
					Signature: signature,
					PublicKey: publicKey,
				}
				signedMsgBytes, err := json.Marshal(signedMsg)
				if err != nil {
					log.Errorf("Failed to marshal message: %v", err)
					return
				}
				wrappedSignedMsg := p2p.Message{
					MessageType: p2p.MessageTypeResponse,
					Content:     string(signedMsgBytes),
				}
				log.Info("Preparing to broadcast signed message to p2p network")
				libp2pService.PublishMessage(ctx, wrappedSignedMsg)
				log.Info("Signed message broadcasted to p2p network")
			case p2p.MessageTypeResponse:
				log.Info("Received completed signed message")
				if config.AppConfig.Proposer {
					log.Info("Proposer processing")
					respSignedMsg := p2p.SignatureMessage{}
					json.Unmarshal([]byte(receivedMsg.Content), &respSignedMsg)
					// TODO: Use better data structure to handle multi sigDoc signatures
					// TODO: Verify signature
					collectedSignatures = append(collectedSignatures, respSignedMsg)
					if len(collectedSignatures) >= config.AppConfig.Threshold {
						log.Info("Signature message sent to MessageChan")
						processCollectedSignatures(collectedSignatures, respSignedMsg.SignDoc)
					}
				}
				log.Info("Regular node processing")
			}
		case <-timeoutChan:
			log.Infof("Signature collection timeout")
		case <-ctx.Done():
			return
		}
	}
}
