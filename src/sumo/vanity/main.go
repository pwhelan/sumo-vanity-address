package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	flag "github.com/ogier/pflag"
	monero "github.com/paxos-bankchain/moneroutil"
	"runtime"
	"time"
)

type keyPair struct {
	Priv *monero.Key
	Pub  *monero.Key
}

type wallet struct {
	SpendKey *keyPair
	ViewKey  *keyPair
}

func (k *keyPair) Regen() {
	var reduceFrom [monero.KeyLength * 2]byte
	rand.Read(reduceFrom[:])
	//copy(reduceFrom[:], tmp)
	monero.ScReduce(k.Priv, &reduceFrom)
	k.Pub = k.Priv.PubKey()
}

func newKeyPair() *keyPair {
	priv, pub := monero.NewKeyPair()
	return &keyPair{Priv: priv, Pub: pub}
}

func worker(c chan *keyPair, s chan struct{}, vanity string) {
	var scratch [36]byte

	spend := newKeyPair()
	header := monero.Uint64ToBytes(0x2bb39a)
	copy(scratch[:], header)
	generated := 0

	for {
		pbuf := spend.Pub.ToBytes()
		copy(scratch[len(header)-1:], pbuf[:])
		slug := monero.EncodeMoneroBase58(scratch[:])
		if slug[6:6+len(vanity)] == vanity {
			c <- spend
		}
		spend.Regen()
		generated++
		if generated >= 100 {
			s <- struct{}{}
			generated = 0
		}
	}
}

func (w *wallet) Address() string {
	sbuf := w.SpendKey.Pub.ToBytes()
	vbuf := w.ViewKey.Pub.ToBytes()
	var buf []byte
	buf = append(buf, monero.Uint64ToBytes(0x2bb39a)...)
	buf = append(buf, sbuf[:]...)
	buf = append(buf, vbuf[:]...)
	scratch := append(monero.Uint64ToBytes(0x2bb39a), buf[:]...)
	csum := monero.GetChecksum(scratch)
	address := fmt.Sprintf("%s%s",
		monero.EncodeMoneroBase58(buf[:]),
		monero.EncodeMoneroBase58(csum[:4]))
	return address
}

func (w *wallet) Print() {
	spbuf := w.SpendKey.Priv.ToBytes()
	vpbuf := w.ViewKey.Priv.ToBytes()
	fmt.Printf("[!] Address: %s\n", w.Address())
	fmt.Printf("[!] SpendKey: %s ViewKey: %s\n",
		hex.EncodeToString(spbuf[:]),
		hex.EncodeToString(vpbuf[:]))
}

func main() {
	var w wallet
	var threads int
	var cores int

	flag.IntVar(&threads, "threads", runtime.GOMAXPROCS(0),
		"set the number of threads to use")
	flag.IntVar(&cores, "cores", 0,
		"set the number of cores the machine has (not usually required)")
	flag.Parse()

	if cores > 0 {
		runtime.GOMAXPROCS(cores)
	}

	fmt.Printf("[*] Threads: %d Cores: %d\n", threads, runtime.GOMAXPROCS(0))
	fmt.Printf("[*] Searching for address starting with Sumoo#%s\n", flag.Arg(0))

	s := make(chan struct{})
	k := make(chan *keyPair)
	for i := 0; i < threads; i++ {
		go worker(k, s, flag.Arg(0))
	}

	t := time.NewTicker(250 * time.Millisecond)
	generated := 0.0
	for {
		select {
		case w.SpendKey = <-k:
			w.ViewKey = newKeyPair()
			fmt.Printf("\n")
			w.Print()
			return
		case <-s:
			generated += 100.0
		case <-t.C:
			fmt.Printf("\r[*] Speed: %f keys/s", generated*4)
			generated = 0.0
		}
	}
}
