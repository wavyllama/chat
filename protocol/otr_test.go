package protocol

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/otr"
)

func TestOTRProtocol_EndSession(t *testing.T) {
	o := NewOTRProtocol()
	o.EndSession()
	assert.False(t, o.Conv.IsEncrypted())
}


func TestOTRProtocol_Handshake(t *testing.T) {
	alice, bob := NewOTRProtocol(), NewOTRProtocol()
	createOTRLink(t, alice, bob)
}

// Inspired by official OTR tests in Golang here: https://github.com/keybase/go-crypto
func createOTRLink(t *testing.T, alice OTRProtocol, bob OTRProtocol) {
	var aMsg, bMsg [][]byte
	aMsg = append(aMsg, []byte(otr.QueryMessage))
	// Simulate a handshake by just sending messages between two users
	for ; len(aMsg[0]) > 0 || len(bMsg[0]) > 0; {
		var err error
		bMsg = [][]byte{}
		for _, msg := range aMsg {
			bMsg, err = bob.Decrypt(msg)
			assert.Error(t, OTRHandshakeStep{}, err)
		}

		aMsg = [][]byte{}
		for _, msg := range bMsg {
			aMsg, err = alice.Decrypt(msg)
			assert.Error(t, OTRHandshakeStep{}, err)
		}
	}
	assert.True(t, alice.IsEncrypted(), "Alice should be encrypted")
	assert.True(t, bob.IsEncrypted(), "Bob should be encrypted")
	assert.Equal(t, alice.Session.SSID, bob.Session.SSID, "Session IDs should be equal")
	assert.Equal(t, alice.Session.Fingerprint, bob.Session.Fingerprint, "Fingerprints should be equal")
}

func TestOTRProtocol_SendAndReceiveMessages(t *testing.T) {
	andrew, sameet := NewOTRProtocol(), NewOTRProtocol()
	createOTRLink(t, andrew, sameet)

	var testMessage = "Want to play fortnite?"
	cyp, err := andrew.Encrypt([]byte(testMessage))
	assert.Nil(t, err)

	for _, msg := range cyp {
		out, err := sameet.Decrypt(msg)
		assert.Nil(t, err)
		assert.Equal(t, testMessage, string(out[0]), "Strings should be equivalent")
	}
}