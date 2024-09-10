package main

import (
	"fmt"

	blst "github.com/supranational/blst/bindings/go"
	"golang.org/x/exp/rand"
)

var blsMode = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

type PrivateKey = blst.SecretKey
type PublicKey = blst.P2Affine
type Signature = blst.P1Affine
type SignatureCompress = []byte
type PublicKeyCompress = []byte

type AggregatePublicKey = blst.P2Aggregate
type AggregateSignature = blst.P1Aggregate

const (
	PubkeyLength    = blst.BLST_P2_COMPRESS_BYTES
	SignatureLength = blst.BLST_P1_COMPRESS_BYTES
)

// Generate a random private key
func generatePrivateKey() *PrivateKey {
	var raw [32]byte
	_, _ = rand.Read(raw[:])
	return blst.KeyGenV3(raw[:])
}

// Generate a public key from a private key
func generatePublicKey(privateKey *PrivateKey) PublicKeyCompress {
	var pk PublicKey
	pk.From(privateKey)
	return pk.Compress()
}

// Sign a message using a private key
func signMessage(privateKey *PrivateKey, message []byte) (SignatureCompress, error) {
	sig := new(Signature).Sign(privateKey, message, blsMode)
	if sig == nil {
		return nil, fmt.Errorf("Failed to generate signature")
	}
	fmt.Printf("Generated signature: %v\n", sig)
	return sig.Compress(), nil
}

// Aggregate multiple signatures into one
func aggregateSignatures(signatures []SignatureCompress) SignatureCompress {
	if len(signatures) == 0 {
		return nil
	}

	aggSignature := new(AggregateSignature)
	success := aggSignature.AggregateCompressed(signatures, true)
	if !success {
		fmt.Println("Error aggregating signatures")
		return nil
	}
	return aggSignature.ToAffine().Compress()
}

// Verify an aggregated signature
func verifyAggregatedSignature(aggSignatureCompress SignatureCompress, publicKeys []PublicKeyCompress, message []byte) bool {
	if len(publicKeys) == 0 {
		return false
	}

	signature := new(Signature).Uncompress(aggSignatureCompress)
	if signature == nil {
		return false
	}

	pubkeys := make([]*PublicKey, 0, len(publicKeys))
	for _, v := range publicKeys {
		pk := new(PublicKey).Uncompress(v)
		if pk == nil {
			return false
		}
		pubkeys = append(pubkeys, pk)
	}
	return signature.FastAggregateVerify(true, pubkeys, message, blsMode)
}

func main() {
	// Generate keys and sign message
	privateKeys := []*PrivateKey{}
	publicKeys := []PublicKeyCompress{}
	signatures := []SignatureCompress{}
	message := []byte("Hello, BLS signature in Golang!")

	// Generate 3 signers
	for i := 0; i < 3; i++ {
		privateKey := generatePrivateKey()
		publicKey := generatePublicKey(privateKey)
		signature, err := signMessage(privateKey, message)
		if err != nil {
			fmt.Println("Failed to generate signature:", err)
			return
		}

		privateKeys = append(privateKeys, privateKey)
		publicKeys = append(publicKeys, publicKey)
		signatures = append(signatures, signature)
	}

	// Aggregate signatures
	aggSignature := aggregateSignatures(signatures)

	// Verify aggregated signature
	isValid := verifyAggregatedSignature(aggSignature, publicKeys, message)
	if isValid {
		fmt.Println("Aggregated signature is valid.")
	} else {
		fmt.Println("Aggregated signature is invalid.")
	}
}
