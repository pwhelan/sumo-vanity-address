package main

import (
	"encoding/hex"
	"fmt"
	flag "github.com/ogier/pflag"
	monero "github.com/paxos-bankchain/moneroutil"
	"runtime"
)

type keyPair struct {
	Priv *monero.Key
	Pub  *monero.Key
}

type wallet struct {
	SpendKey *keyPair
	ViewKey  *keyPair
}

func newKeyPair() *keyPair {
	priv, pub := monero.NewKeyPair()
	return &keyPair{Priv: priv, Pub: pub}
}

func worker(c chan *keyPair, vanity string) {
	for {
		spend := newKeyPair()
		pbuf := spend.Pub.ToBytes()
		scratch := append(monero.Uint64ToBytes(0x2bb39a), pbuf[:]...)
		slug := monero.EncodeMoneroBase58(scratch)
		if slug[6:6+len(vanity)] == vanity {
			c <- spend
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
	c := make(chan *keyPair)
	for i := 0; i < threads; i++ {
		go worker(c, flag.Arg(0))
	}

	w.SpendKey = <-c
	w.ViewKey = newKeyPair()
	w.Print()
}
