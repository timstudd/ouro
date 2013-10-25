package main

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
    "bytes"
    "fmt"
    "net"
)


func encodeBase64(b []byte) []byte {
    var buf bytes.Buffer
    encoder := base64.NewEncoder(base64.StdEncoding, &buf)
    encoder.Write(b)
    encoder.Close()
    return buf.Bytes()
}
    
func decodeBase64(b []byte) ([]byte, error) {
    encoder := base64.NewDecoder(base64.StdEncoding,  bytes.NewBuffer(b))
    buf := make([]byte, base64.StdEncoding.DecodedLen(len(b)))
    n, err := encoder.Read(buf)
    if err != nil {
        return nil, err
    }
    return buf[:n], nil
}


func encrypt(key, text []byte) []byte {
    block, err := aes.NewCipher(key)
    if err != nil {
        panic(err)
    }
    b := encodeBase64(text)
    ciphertext := make([]byte, aes.BlockSize+len(b))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        panic(err)
    }
    cfb := cipher.NewCFBEncrypter(block, iv)
    cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

    return encodeBase64(ciphertext)
}

func decrypt(key, text []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    text, err = decodeBase64(text)
    if err != nil {
        return nil, err
    }
    if len(text) < aes.BlockSize {
        return nil, err
    }
    iv := text[:aes.BlockSize]
    text = text[aes.BlockSize:]
    cfb := cipher.NewCFBDecrypter(block, iv)
    cfb.XORKeyStream(text, text)

    decoded, err := decodeBase64(text)

    return decoded, err
}


type EncryptedConnection struct {
    Destination io.ReadWriter
}

func (e *EncryptedConnection) Key(in bool) []byte {
    conn := e.Destination.(net.Conn)
    key := make([]byte, 32)
    if in {
        copy(key, fmt.Sprintf("%s%s", conn.LocalAddr().String(), conn.RemoteAddr().String()))
    } else {
        copy(key, fmt.Sprintf("%s%s", conn.RemoteAddr().String(), conn.LocalAddr().String()))
    }
    return key
}

func (e *EncryptedConnection) Write(data []byte) (int, error) {
    out := make([]byte, 65535)
    out = encrypt(e.Key(true), data)

    return e.Destination.Write(out)
}

func (e *EncryptedConnection) Read(in []byte) (int, error) {
    e.Destination.Read(in)

    in = in[:bytes.IndexByte(in, 0)]
    out, err := decrypt(e.Key(false), in)
    if err != nil {
        return 0, err
    }

    // Clear contents
    copy(in, make([]byte, len(in)))
    // Write decrypted contents
    copy(in, out)

    return len(in), nil
}