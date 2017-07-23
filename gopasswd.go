package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func lowcase(b byte) byte {
	return 'a' + b%26
}

// 33-64 special
// 48-57 - numbers

const numbers = "0123456789"
const specials = `!#$%^&*+=?/\_`

func number(n uint64) string {
	n = n % uint64(len(numbers))
	return numbers[n : n+1]
}
func special(n uint64) string {
	n = n % uint64(len(specials))
	return specials[n : n+1]
}

func pwchar(b byte) (s string) {
	s = string([]byte{lowcase(b)})
	return
}

func alphaword(w string) bool {
	for _, b := range strings.ToLower(w) {
		if b < 'a' || b > 'z' {
			return false
		}

	}
	return true
}

func dict(fn string) (words []string) {
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		w := scanner.Text()
		if alphaword(w) {
			words = append(words, w)
		}
	}
	return words
}

type Password struct {
	Salt string
}

type Hash string

func (h *Hash) set(val string) {
	s := sha256.New()
	s.Write([]byte(val))
	*h = Hash(s.Sum(nil))
}

func (h *Hash) get(num int) (val uint64) {
	for num > 0 {
		val = (val << 8) | uint64((*h)[0])
		num = num - 1
		*h = (*h)[1:]
	}
	return
}

func main() {

	var args = struct {
		name    string
		len     int
		salt    string
		dict    string
		passwd  string
		verbose bool
	}{}

	flag.StringVar(&args.dict, "dict", "/usr/share/dict/american-english", "")
	flag.StringVar(&args.name, "name", "", "")
	flag.StringVar(&args.salt, "salt", "", "")
	flag.StringVar(&args.passwd, "passwd", "", "")
	flag.IntVar(&args.len, "len", 8, "")
	flag.BoolVar(&args.verbose, "verbose", false, "")

	flag.Parse()

	pws := make(map[string][]Password)

	content, err := ioutil.ReadFile("config.json")
	if err == nil {
		err = json.Unmarshal(content, &pws)
	}

	salt := args.salt
	entries := []Password{Password{Salt: args.salt}}
	if e, ok := pws[args.name]; ok {
		entries = e
	}

	if len(entries) > 0 {
		if salt == "" {
			salt = entries[len(entries)-1].Salt
		}
	}

	if salt != "" && len(entries) > 0 && salt != entries[len(entries)-1].Salt {
		ne := []Password{}
		for _, e := range entries {
			if e.Salt != salt {
				ne = append(ne, e)
			}
		}
		ne = append(ne, Password{Salt: salt})
		entries = ne
	}

	//fmt.Printf("%s\n", salt)
	words := dict(args.dict)
	nw := uint64(len(words))

	hh := Hash("")
	hh.set(salt + args.name + args.passwd)

	pw := ""

	pos := hh.get(3)
	w := words[pos%nw]
	pw += strings.ToUpper(w[:1]) + w[1:]

	pos = hh.get(3)
	w = words[pos%nw]
	pw += strings.ToUpper(w[:1]) + w[1:]

	pw += special(hh.get(1))
	pw += number(hh.get(1))

	fmt.Printf("%s\n", pw)

	pws[args.name] = entries
	content, err = json.MarshalIndent(pws, "", "\t")
	if err == nil {
		ioutil.WriteFile("config.json", content, 0600)
	}

	if args.verbose {
		fmt.Printf("words %d, specials %d\n", len(words), len(specials))
	}
}
