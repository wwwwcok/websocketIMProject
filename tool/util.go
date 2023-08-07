package tool

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
)

func BaseEncode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func BaseDecode(data string) []byte {
	decodeString, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil
	}
	return decodeString
}

func GobEncode(data []byte) []byte {
	var encodeData bytes.Buffer
	encoder := gob.NewEncoder(&encodeData)
	encoder.Encode(data)
	return encodeData.Bytes()
}

func GobDecode(data []byte, Node any) any {

	source := bytes.NewReader(data)
	newDecoder := gob.NewDecoder(source)
	err := newDecoder.Decode(Node)
	if err != nil {
		return nil
	}
	return Node
}
