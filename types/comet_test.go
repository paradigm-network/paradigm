package types

import (
	"testing"
	"github.com/paradigm-network/paradigm/common/crypto"
	"time"
)

func createDummyCometBody() CometBody {
	body := CometBody{}
	body.Transactions = [][]byte{[]byte("abc"), []byte("def")}
	body.Parents = []string{"self", "other"}
	body.Creator = []byte("0x041CC70F28CFC463A5ABC511D7C47228227ECC43CFB1BD82BCC34CAC6382E8E6D03A9E3F50A1CFB34925A20E5346D5C1C569E3191F2C36766ED99114DED9C88C51")
	body.Timestamp = time.Now().UTC()
	body.BlockSignatures = []BlockSignature{
		{
			Validator: body.Creator,
			Index:     0,
			Signature: "ysr70kglonsbb2e0aae92y1nqgds15kuujouf3qgihnj4itp7|3nbqaxcshcacdhes1vmpouxtnsnwzlyy2lw4oigew9oqz5dgcg",
		},
	}
	return body
}

func TestSignComet(t *testing.T) {
	privateKey, _ := crypto.GenerateECDSAKey()
	//publicKeyBytes := crypto.FromECDSAPub(&privateKey.PublicKey)

	body := createDummyCometBody()
	//body.Creator = publicKeyBytes

	comet := Comet{Body: body}
	if err := comet.Sign(privateKey); err != nil {
		t.Fatalf("Error signing Event: %s", err)
	}

	res, err := comet.Verify()
	if err != nil {
		t.Fatalf("Error verifying signature: %s", err)
	}
	if !res {
		t.Fatalf("Verify returned false")
	}
}
