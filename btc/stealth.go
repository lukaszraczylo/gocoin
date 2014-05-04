package btc

import (
	"bytes"
	"errors"
	"crypto/sha256"
)


func StealthAddressVersion(testnet bool) byte {
	if testnet {
		return 43
	} else {
		return 42
	}
}


type StealthAddr struct {
	Version byte
	Options byte
	ScanKey [33]byte
	SpendKeys [][33]byte
	Sigs byte
	Prefix []byte
}


func NewStealthAddrFromString(hs string) (a *StealthAddr, e error) {
	var tmp byte

	dec := Decodeb58(hs)
	if dec == nil {
		e = errors.New("StealthAddr: Cannot decode b58 string *"+hs+"*")
		return
	}
	if (len(dec)<2+33+33+1+1+4) {
		e = errors.New("StealthAddr: data too short")
		return
	}

	sh := Sha2Sum(dec[0:len(dec)-4])
	if !bytes.Equal(sh[:4], dec[len(dec)-4:len(dec)]) {
		e = errors.New("StealthAddr: Checksum error")
		return
	}

	a = new(StealthAddr)

	b := bytes.NewBuffer(dec[0:len(dec)-4])

	if a.Version, e = b.ReadByte(); e != nil {
		a = nil
		return
	}
	if a.Options, e = b.ReadByte(); e != nil {
		a = nil
		return
	}
	if _, e = b.Read(a.ScanKey[:]); e != nil {
		a = nil
		return
	}
	if tmp, e = b.ReadByte(); e != nil {
		a = nil
		return
	}
	a.SpendKeys = make([][33]byte, int(tmp))
	for i := range a.SpendKeys {
		if _, e = b.Read(a.SpendKeys[i][:]); e != nil {
			a = nil
			return
		}
	}
	if a.Sigs, e = b.ReadByte(); e != nil {
		a = nil
		return
	}
	a.Prefix = b.Bytes()
	return
}


func (a *StealthAddr) String() (string) {
	b := new(bytes.Buffer)
	b.WriteByte(a.Version)
	b.WriteByte(a.Options)
	b.Write(a.ScanKey[:])
	b.WriteByte(byte(len(a.SpendKeys)))
	for i:=range a.SpendKeys {
		b.Write(a.SpendKeys[i][:])
	}
	b.WriteByte(a.Sigs)
	b.Write(a.Prefix)
	sh := Sha2Sum(b.Bytes())
	b.Write(sh[:4])
	return Encodeb58(b.Bytes())
}


// Calculate the stealth difference
func StealthDH(pub, priv []byte) []byte {
	pk, er := NewPublicKey(pub)
	if er != nil {
		println(er.Error())
		return nil
	}
	x, _ := secp256k1.ScalarMult(pk.X, pk.Y, priv)
	s := sha256.New()
	s.Write([]byte{0x03})
	s.Write(x.Bytes())
	return s.Sum(nil)
}


// Calculate the stealth difference
func StealthPub(pub, priv []byte) []byte {
	pk, er := NewPublicKey(pub)
	if er != nil {
		println(er.Error())
		return nil
	}
	x, y := secp256k1.ScalarMult(pk.X, pk.Y, priv)

	return append([]byte{2+byte(y.Bit(0))}, x.Bytes()...)
}