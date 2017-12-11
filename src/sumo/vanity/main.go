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

func worker(c chan keyPair, vanity string) {
	for {
		priv, pub := monero.NewKeyPair()
		pbuf := pub.ToBytes()
		scratch := append(monero.Uint64ToBytes(0x2bb39a), pbuf[:]...)
		slug := monero.EncodeMoneroBase58(scratch)
		if slug[6:6+len(vanity)] == vanity {
			c <- keyPair{priv, pub}
		}
	}
}

func main() {
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
	c := make(chan keyPair)
	for i := 0; i < threads; i++ {
		go worker(c, flag.Arg(0))
	}
	skeys := <-c
	vpriv, vpub := monero.NewKeyPair()
	sbuf := skeys.Pub.ToBytes()
	vbuf := vpub.ToBytes()
	spbuf := skeys.Priv.ToBytes()
	vpbuf := vpriv.ToBytes()
	var buf []byte
	buf = append(buf, monero.Uint64ToBytes(0x2bb39a)...)
	buf = append(buf, sbuf[:]...)
	buf = append(buf, vbuf[:]...)
	scratch := append(monero.Uint64ToBytes(0x2bb39a), buf[:]...)
	csum := monero.GetChecksum(scratch)
	address := fmt.Sprintf("%s%s",
		monero.EncodeMoneroBase58(buf[:]),
		monero.EncodeMoneroBase58(csum[:4]))
	fmt.Printf("ADDRESS=%s\n", address)
	fmt.Printf("SPENDKEY=%s VIEWKEY=%s\n",
		hex.EncodeToString(spbuf[:]),
		hex.EncodeToString(vpbuf[:]))
	fmt.Printf("PUBLIC KEYS=%s / %s\n",
		hex.EncodeToString(sbuf[:]),
		hex.EncodeToString(vbuf[:]))
}
