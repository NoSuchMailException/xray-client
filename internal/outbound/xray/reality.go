package xray

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"golang.org/x/crypto/curve25519"
)

func buildSessionID(serverPub []byte, shortID []byte) (sessionID, ePub, ePriv [32]byte, err error) {
	if _, err = rand.Read(ePriv[:]); err != nil {
		return sessionID, ePub, ePriv, fmt.Errorf("rand: %w", err)
	}

	ePriv[0] &= 248
	ePriv[31] &= 127
	ePriv[31] |= 64

	curve25519.ScalarBaseMult(&ePub, &ePriv)
	sharedSecret, err := curve25519.X25519(ePriv[:], serverPub)
	if err != nil {
		return sessionID, ePub, ePriv, fmt.Errorf("x25519: %w", err)
	}

	window := time.Now().Unix() / 30

	h := hmac.New(sha256.New, sharedSecret)
	h.Write([]byte("REALITY"))

	var windowBytes [8]byte
	binary.BigEndian.PutUint64(windowBytes[:], uint64(window))
	h.Write(windowBytes[:])

	authToken := h.Sum(nil)

	copy(sessionID[0:24], authToken[0:24])
	copy(sessionID[24:32], shortID)
	return sessionID, ePub, ePriv, nil
}
