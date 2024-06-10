package encrypt

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "Foo not bar"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := NewEncryptionKey()
	_, err := CopyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	// Check the Encryption.
	fmt.Println(len(payload))
	fmt.Println(len(dst.String()))

	// Check the Descryption.
	out := new(bytes.Buffer) // Dynamic Buffer.
	nw, err := CopyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	if nw != 16+len(payload) {
		t.Fail()
	}

	if out.String() != payload {
		t.Errorf("decryption failed!!!")
	}
}
