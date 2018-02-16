package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	flag "github.com/ogier/pflag"
	"hash/crc32"
	monero "github.com/paxos-bankchain/moneroutil"
	"regexp"
	"runtime"
	"strings"
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

func worker(k chan *keyPair, s chan struct{}, numeral int, vanity string) {
	generated := 0
	nc := fmt.Sprintf("%d", numeral)

	for {
		spend := newKeyPair()
		pbuf := spend.Pub.ToBytes()
		scratch := append(monero.Uint64ToBytes(0x2bb39a), pbuf[:]...)
		slug := monero.EncodeMoneroBase58(scratch[:])

		if slug[6:6+len(vanity)] == vanity && (numeral == 0 || slug[5:5+1] == nc) {
			k <- spend
			return
		}
		generated++
		if generated >= 100 {
			s <- struct{}{}
			generated = 0
		}
	}
}

func (w *wallet) Address() string {
	prefix := monero.Uint64ToBytes(0x2bb39a)
	csum := monero.GetChecksum(prefix, w.SpendKey.Pub[:], w.ViewKey.Pub[:])
	return monero.EncodeMoneroBase58(prefix, w.SpendKey.Pub[:],
		w.ViewKey.Pub[:], csum[:4])
}

func create_checksum_index(words []string) int {
	h := crc32.NewIEEE()
	for _, word := range words {
		// uniq_prefix EN=3
		if len(word) > 3 {
			h.Write([]byte(word[:3])) 
		} else {
			h.Write([]byte(word))
		}
	}
	crc := h.Sum(nil)
	crci := int(binary.BigEndian.Uint32(crc))
	return crci % 24
}

func create_checksum_index2(words []string) int {
	h := crc32.NewIEEE()
	for _, word := range words {
		// uniq_prefix EN=3
		if len(word) > 3 {
			h.Write([]byte(word[:3])) 
		} else {
			h.Write([]byte(word))
		}
	}
	crc := h.Sum(nil)
	crci := int(binary.BigEndian.Uint32(crc))
	return crci % len(electrum_words)
}

func (w *wallet) Print() {
	spbuf := w.SpendKey.Priv.ToBytes()
	vpbuf := w.ViewKey.Priv.ToBytes()
	fmt.Printf("[!] Address: %s\n", w.Address())
	fmt.Printf("[!] SpendKey: %s ViewKey: %s\n",
		hex.EncodeToString(spbuf[:]),
		hex.EncodeToString(vpbuf[:]))
	fmt.Println("Seed: ");
	b := spbuf[:]
	var words []string
	for i := 0; i < len(b); i+=4 {
		val := int(binary.LittleEndian.Uint32([]byte{b[i], b[i+1], b[i+2], b[i+3]}))
		w1 := val % len(electrum_words)
		w2 := ((val / len(electrum_words)) + w1) % len(electrum_words)
		w3 := (((val / len(electrum_words)) / len(electrum_words)) + w2) % len(electrum_words)
		
		words = append(words, electrum_words[w1], electrum_words[w2], electrum_words[w3])
		
	}
	words = append(words, words[create_checksum_index(words)])
	words = append(words, electrum_words[create_checksum_index2(words)])
	
	fmt.Println(strings.Join(words, " "))
}

func main() {
	var w wallet
	var threads int
	var cores int
	numeral := 0

	flag.IntVar(&threads, "threads", runtime.GOMAXPROCS(0),
		"set the number of threads to use")
	flag.IntVar(&cores, "cores", 0,
		"set the number of cores the machine has (not usually required)")
	flag.IntVar(&numeral, "numeral", 0, "set the leading numeral in the address")
	flag.Parse()

	if numeral >= 7 || numeral < 0 {
		fmt.Printf("Cannot produce addresses with a leading numeral of %d\n", numeral)
		return
	}
	
	if cores > 0 {
		runtime.GOMAXPROCS(cores)
	}

	re := regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]+$`).MatchString
	if re(flag.Arg(0)) == false {
		fmt.Printf("Slug has illegal characters: %s\n", flag.Arg(0))
		return
	}

	fmt.Printf("[*] Threads: %d Cores: %d\n", threads, runtime.GOMAXPROCS(0))
	if numeral == 0 {
		fmt.Printf("[*] Searching for address starting with Sumoo#%s\n", flag.Arg(0))
	} else {
		fmt.Printf("[*] Searching for address starting with Sumoo%d%s\n", numeral, flag.Arg(0))
	}

	s := make(chan struct{})
	k := make(chan *keyPair)
	for i := 0; i < threads; i++ {
		go worker(k, s, numeral, flag.Arg(0))
	}

	t := time.NewTicker(250 * time.Millisecond)
	generated := 0.0
	for {
		select {
		case w.SpendKey = <-k:
			
			// Generate Determenistic View Key
			w.ViewKey = &keyPair{Priv: &monero.Key{}, Pub: nil} // newKeyPair()
			sb := w.SpendKey.Priv.ToBytes()
			w.ViewKey.Priv.FromBytes(monero.Keccak256(sb[:]))
			monero.ScReduce32(w.ViewKey.Priv)
			w.ViewKey.Pub = w.ViewKey.Priv.PubKey()
			
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
